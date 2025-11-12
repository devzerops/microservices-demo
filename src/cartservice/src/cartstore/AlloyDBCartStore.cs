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
using System.Threading.Tasks;
using Google.Api.Gax.ResourceNames;
using Google.Cloud.SecretManager.V1;
 
namespace cartservice.cartstore
{
    public class AlloyDBCartStore : ICartStore
    {
        private readonly string tableName;
        private readonly string connectionString;
        private readonly string readConnectionString;

        public AlloyDBCartStore(IConfiguration configuration)
        {
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
            connectionString = "Host="          +
                               primaryIPAddress +
                               ";Username="     +
                               alloyDBUser      +
                               ";Password="     +
                               alloyDBPassword  +
                               ";Database="     +
                               databaseName;

            // Optional: Read replica connection for read-heavy workloads
            // If ALLOYDB_READ_IP is configured, read operations can be directed to read pool
            // This improves performance and reduces load on the primary instance
            string readIPAddress = configuration["ALLOYDB_READ_IP"];
            if (!string.IsNullOrEmpty(readIPAddress))
            {
                readConnectionString = "Host="          +
                                      readIPAddress    +
                                      ";Username="     +
                                      alloyDBUser      +
                                      ";Password="     +
                                      alloyDBPassword  +
                                      ";Database="     +
                                      databaseName;
                Console.WriteLine($"AlloyDB read pool configured at {readIPAddress}");
            }
            else
            {
                // If no read replica configured, use primary for all operations
                readConnectionString = connectionString;
            }

            tableName = configuration["ALLOYDB_TABLE_NAME"];
        }


        public async Task AddItemAsync(string userId, string productId, int quantity)
        {
            Console.WriteLine($"AddItemAsync for {userId} called");
            try
            {
                await using var dataSource = NpgsqlDataSource.Create(connectionString);

                // Fetch the current quantity for our userId/productId tuple
                var fetchCmd = $"SELECT quantity FROM {tableName} WHERE userID='{userId}' AND productID='{productId}'";
                var currentQuantity = 0;
                var cmdRead = dataSource.CreateCommand(fetchCmd);
                await using (var reader = await cmdRead.ExecuteReaderAsync())
                {
                    while (await reader.ReadAsync())
                        currentQuantity += reader.GetInt32(0);
                }
                var totalQuantity = quantity + currentQuantity;

                var insertCmd = $"INSERT INTO {tableName} (userId, productId, quantity) VALUES ('{userId}', '{productId}', {totalQuantity})";
                await using (var cmdInsert = dataSource.CreateCommand(insertCmd))
                {
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
            Console.WriteLine($"GetCartAsync called for userId={userId}");
            Hipstershop.Cart cart = new();
            cart.UserId = userId;
            try
            {
                // Use read connection for read-only operations to leverage read pool
                await using var dataSource = NpgsqlDataSource.Create(readConnectionString);

                var cartFetchCmd = $"SELECT productId, quantity FROM {tableName} WHERE userId = '{userId}'";
                var cmd = dataSource.CreateCommand(cartFetchCmd);
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
            Console.WriteLine($"EmptyCartAsync called for userId={userId}");

            try
            {
                await using var dataSource = NpgsqlDataSource.Create(connectionString);
                var deleteCmd = $"DELETE FROM {tableName} WHERE userID = '{userId}'";
                await using (var cmd = dataSource.CreateCommand(deleteCmd))
                {
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
                return true;
            }
            catch (Exception)
            {
                return false;
            }
        }
    }
}

