// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

using System;
using System.Text.RegularExpressions;
using Grpc.Core;
using Npgsql;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;
using System.Threading.Tasks;
using Google.Api.Gax.ResourceNames;
using Google.Cloud.SecretManager.V1;

namespace cartservice.cartstore
{
    public class AlloyDBCartStore : ICartStore
    {
        private readonly string tableName;
        private readonly string connectionString;
        private readonly ILogger<AlloyDBCartStore> _logger;

        public AlloyDBCartStore(IConfiguration configuration, ILogger<AlloyDBCartStore> logger)
        {
            _logger = logger;

            // Create a Cloud Secrets client.
            SecretManagerServiceClient client = SecretManagerServiceClient.Create();
            var projectId = configuration["PROJECT_ID"];
            var secretId = configuration["ALLOYDB_SECRET_NAME"];
            SecretVersionName secretVersionName = new SecretVersionName(projectId, secretId, "latest");

            AccessSecretVersionResponse result = client.AccessSecretVersion(secretVersionName);
            // Convert the payload to a string. Payloads are bytes by default.
            string alloyDBPassword = result.Payload.Data.ToStringUtf8().TrimEnd('\r', '\n');

            // TODO: Create a separate user for connecting within the application
            // rather than using our superuser
            string alloyDBUser = "postgres";
            string databaseName = configuration["ALLOYDB_DATABASE_NAME"];
            // TODO: Consider splitting workloads into read vs. write and take
            // advantage of the AlloyDB read pools
            string primaryIPAddress = configuration["ALLOYDB_PRIMARY_IP"];
            connectionString = "Host="          +
                               primaryIPAddress +
                               ";Username="     +
                               alloyDBUser      +
                               ";Password="     +
                               alloyDBPassword  +
                               ";Database="     +
                               databaseName;

            tableName = configuration["ALLOYDB_TABLE_NAME"];
            ValidateTableName(tableName);
        }

        /// <summary>
        /// Validates that a table name contains only safe characters to prevent SQL injection.
        /// Table names cannot be parameterized in SQL, so they must be validated.
        /// </summary>
        /// <param name="name">The table name to validate</param>
        /// <exception cref="ArgumentException">Thrown when table name is invalid</exception>
        private void ValidateTableName(string name)
        {
            if (string.IsNullOrEmpty(name))
            {
                throw new ArgumentException("Table name cannot be null or empty");
            }

            // Allow only alphanumeric characters and underscores, must start with letter or underscore
            var validTableName = new Regex(@"^[a-zA-Z_][a-zA-Z0-9_]*$");
            if (!validTableName.IsMatch(name))
            {
                throw new ArgumentException(
                    "Invalid table name: must contain only letters, numbers, and underscores, and start with a letter or underscore");
            }

            if (name.Length > 63)
            {
                throw new ArgumentException("Table name too long: maximum 63 characters");
            }
        }


        public async Task AddItemAsync(string userId, string productId, int quantity)
        {
            _logger.LogInformation("AddItemAsync called for userId={UserId}", userId);
            try
            {
                await using var dataSource = NpgsqlDataSource.Create(connectionString);

                // Fetch the current quantity for our userId/productId tuple
                // Use parameterized query to prevent SQL injection
                var fetchQuery = $"SELECT quantity FROM {tableName} WHERE userID = $1 AND productID = $2";
                var currentQuantity = 0;
                await using (var cmdRead = dataSource.CreateCommand(fetchQuery))
                {
                    cmdRead.Parameters.AddWithValue(userId);
                    cmdRead.Parameters.AddWithValue(productId);
                    await using (var reader = await cmdRead.ExecuteReaderAsync())
                    {
                        while (await reader.ReadAsync())
                            currentQuantity += reader.GetInt32(0);
                    }
                }
                var totalQuantity = quantity + currentQuantity;

                // Use parameterized query to prevent SQL injection
                var insertQuery = $"INSERT INTO {tableName} (userId, productId, quantity) VALUES ($1, $2, $3)";
                await using (var cmdInsert = dataSource.CreateCommand(insertQuery))
                {
                    cmdInsert.Parameters.AddWithValue(userId);
                    cmdInsert.Parameters.AddWithValue(productId);
                    cmdInsert.Parameters.AddWithValue(totalQuantity);
                    await Task.Run(() =>
                    {
                        return cmdInsert.ExecuteNonQueryAsync();
                    });
                }
            }
            catch (Exception ex)
            {
                throw new RpcException(
                    new Status(StatusCode.FailedPrecondition, $"Can't access cart storage at {connectionString}. {ex}"));
            }
        }


        public async Task<Hipstershop.Cart> GetCartAsync(string userId)
        {
            _logger.LogInformation("GetCartAsync called for userId={UserId}", userId);
            Hipstershop.Cart cart = new();
            cart.UserId = userId;
            try
            {
                await using var dataSource = NpgsqlDataSource.Create(connectionString);

                // Use parameterized query to prevent SQL injection
                var cartFetchQuery = $"SELECT productId, quantity FROM {tableName} WHERE userId = $1";
                await using (var cmd = dataSource.CreateCommand(cartFetchQuery))
                {
                    cmd.Parameters.AddWithValue(userId);
                    await using (var reader = await cmd.ExecuteReaderAsync())
                    {
                        while (await reader.ReadAsync())
                        {
                            Hipstershop.CartItem item = new()
                            {
                                ProductId = reader.GetString(0),
                                Quantity = reader.GetInt32(1)
                            };
                            cart.Items.Add(item);
                        }
                    }
                }
                await Task.Run(() =>
                {
                    return cart;
                });
            }
            catch (Exception ex)
            {
                throw new RpcException(
                    new Status(StatusCode.FailedPrecondition, $"Can't access cart storage at {connectionString}. {ex}"));
            }
            return cart;
        }


        public async Task EmptyCartAsync(string userId)
        {
            _logger.LogInformation("EmptyCartAsync called for userId={UserId}", userId);

            try
            {
                await using var dataSource = NpgsqlDataSource.Create(connectionString);
                // Use parameterized query to prevent SQL injection
                var deleteQuery = $"DELETE FROM {tableName} WHERE userID = $1";
                await using (var cmd = dataSource.CreateCommand(deleteQuery))
                {
                    cmd.Parameters.AddWithValue(userId);
                    await Task.Run(() =>
                    {
                        return cmd.ExecuteNonQueryAsync();
                    });
                }
            }
            catch (Exception ex)
            {
                throw new RpcException(
                    new Status(StatusCode.FailedPrecondition, $"Can't access cart storage at {connectionString}. {ex}"));
            }
        }

        public bool Ping()
        {
            try
            {
                using var dataSource = NpgsqlDataSource.Create(connectionString);
                using var connection = dataSource.CreateConnection();
                connection.Open();
                return connection.State == System.Data.ConnectionState.Open;
            }
            catch (Exception ex)
            {
                _logger.LogWarning("Ping failed: {Error}", ex.Message);
                return false;
            }
        }
    }
}

