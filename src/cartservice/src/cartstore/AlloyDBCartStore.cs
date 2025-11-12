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
        private readonly NpgsqlDataSource dataSource;
        private readonly NpgsqlDataSource readDataSource;
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

            // Use dedicated application user instead of superuser for least privilege access
            // Default to "postgres" for backward compatibility, but should be configured
            // with a dedicated user (e.g., "cartservice_user") in production
            string alloyDBUser = configuration["ALLOYDB_USER"] ?? "postgres";
            string databaseName = configuration["ALLOYDB_DATABASE_NAME"];

            // Primary connection string for write operations
            string primaryIPAddress = configuration["ALLOYDB_PRIMARY_IP"];
            string connectionString = "Host="          +
                                      primaryIPAddress +
                                      ";Username="     +
                                      alloyDBUser      +
                                      ";Password="     +
                                      alloyDBPassword  +
                                      ";Database="     +
                                      databaseName     +
                                      ";Timeout=30;Command Timeout=30";

            // Create primary data source with connection pooling
            dataSource = NpgsqlDataSource.Create(connectionString);

            // Optional: Read replica connection for read-heavy workloads
            // If ALLOYDB_READ_IP is configured, read operations can be directed to read pool
            // This improves performance and reduces load on the primary instance
            string readIPAddress = configuration["ALLOYDB_READ_IP"];
            if (!string.IsNullOrEmpty(readIPAddress))
            {
                string readConnectionString = "Host="          +
                                              readIPAddress    +
                                              ";Username="     +
                                              alloyDBUser      +
                                              ";Password="     +
                                              alloyDBPassword  +
                                              ";Database="     +
                                              databaseName     +
                                              ";Timeout=30;Command Timeout=30";
                readDataSource = NpgsqlDataSource.Create(readConnectionString);
                _logger.LogInformation("AlloyDB read pool configured at {ReadIP}", readIPAddress);
            }
            else
            {
                // If no read replica configured, use primary for all operations
                readDataSource = dataSource;
            }

            tableName = configuration["ALLOYDB_TABLE_NAME"];
        }


        public async Task AddItemAsync(string userId, string productId, int quantity)
        {
            _logger.LogInformation("AddItemAsync called for userId={UserId}", userId);
            try
            {
                // Fetch the current quantity for our userId/productId tuple
                // Use parameterized query to prevent SQL injection
                var fetchCmd = $"SELECT quantity FROM {tableName} WHERE userID = $1 AND productID = $2";
                var currentQuantity = 0;
                await using (var cmdRead = dataSource.CreateCommand(fetchCmd))
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
                var insertCmd = $"INSERT INTO {tableName} (userId, productId, quantity) VALUES ($1, $2, $3)";
                await using (var cmdInsert = this.dataSource.CreateCommand(insertCmd))
                {
                    cmdInsert.Parameters.AddWithValue(userId);
                    cmdInsert.Parameters.AddWithValue(productId);
                    cmdInsert.Parameters.AddWithValue(totalQuantity);
                    await cmdInsert.ExecuteNonQueryAsync();
                }
            }
            catch (Exception ex)
            {
                throw new RpcException(
                    new Status(StatusCode.FailedPrecondition, $"Can't access cart storage. {ex}"));
            }
        }


        public async Task<Hipstershop.Cart> GetCartAsync(string userId)
        {
            _logger.LogInformation("GetCartAsync called for userId={UserId}", userId);
            Hipstershop.Cart cart = new();
            cart.UserId = userId;
            try
            {
                // Use read connection for read-only operations to leverage read pool
                // Use parameterized query to prevent SQL injection
                var cartFetchCmd = $"SELECT productId, quantity FROM {tableName} WHERE userId = $1";
                await using (var cmd = readDataSource.CreateCommand(cartFetchCmd))
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
            }
            catch (Exception ex)
            {
                throw new RpcException(
                    new Status(StatusCode.FailedPrecondition, $"Can't access cart storage. {ex}"));
            }
            return cart;
        }


        public async Task EmptyCartAsync(string userId)
        {
            _logger.LogInformation("EmptyCartAsync called for userId={UserId}", userId);

            try
            {
                // Use parameterized query to prevent SQL injection
                var deleteCmd = $"DELETE FROM {tableName} WHERE userID = $1";
                await using (var cmd = dataSource.CreateCommand(deleteCmd))
                {
                    cmd.Parameters.AddWithValue(userId);
                    await cmd.ExecuteNonQueryAsync();
                }
            }
            catch (Exception ex)
            {
                throw new RpcException(
                    new Status(StatusCode.FailedPrecondition, $"Can't access cart storage. {ex}"));
            }
        }

        public bool Ping()
        {
            try
            {
                using var connection = dataSource.CreateConnection();
                connection.Open();
                return connection.State == System.Data.ConnectionState.Open;
            }
            catch (Exception)
            {
                return false;
            }
        }
    }
}

