using System;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Diagnostics.HealthChecks;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Diagnostics.HealthChecks;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using cartservice.cartstore;
using cartservice.services;
using Microsoft.Extensions.Caching.StackExchangeRedis;

namespace cartservice
{
    public class Startup
    {
        private readonly ILogger<Startup> _logger;

        public Startup(IConfiguration configuration, ILogger<Startup> logger)
        {
            Configuration = configuration;
            _logger = logger;
        }

        public IConfiguration Configuration { get; }
        
        // This method gets called by the runtime. Use this method to add services to the container.
        // For more information on how to configure your application, visit https://go.microsoft.com/fwlink/?LinkID=398940
        public void ConfigureServices(IServiceCollection services)
        {
            string redisAddress = Configuration["REDIS_ADDR"];
            string spannerProjectId = Configuration["SPANNER_PROJECT"];
            string spannerConnectionString = Configuration["SPANNER_CONNECTION_STRING"];
            string alloyDBConnectionString = Configuration["ALLOYDB_PRIMARY_IP"];

            if (!string.IsNullOrEmpty(redisAddress))
            {
                services.AddStackExchangeRedisCache(options =>
                {
                    // Enable SSL for production Redis connections
                    // Format: "host:port,ssl=true,abortConnect=false"
                    string redisConfig = redisAddress;
                    if (!redisAddress.Contains("ssl="))
                    {
                        redisConfig = $"{redisAddress},ssl=true,abortConnect=false";
                    }
                    options.Configuration = redisConfig;
                    _logger.LogInformation("Redis configured with SSL enabled: {RedisAddress}", redisAddress);
                });
                services.AddSingleton<ICartStore, RedisCartStore>();
            }
            else if (!string.IsNullOrEmpty(spannerProjectId) || !string.IsNullOrEmpty(spannerConnectionString))
            {
                services.AddSingleton<ICartStore, SpannerCartStore>();
            }
            else if (!string.IsNullOrEmpty(alloyDBConnectionString))
            {
                _logger.LogInformation("Creating AlloyDB cart store");
                services.AddSingleton<ICartStore, AlloyDBCartStore>();
            }
            else
            {
                _logger.LogInformation("Redis cache host was not specified. Starting cart service using in-memory store");
                services.AddDistributedMemoryCache();
                services.AddSingleton<ICartStore, RedisCartStore>();
            }


            services.AddGrpc();
        }

        // This method gets called by the runtime. Use this method to configure the HTTP request pipeline.
        public void Configure(IApplicationBuilder app, IWebHostEnvironment env)
        {
            if (env.IsDevelopment())
            {
                app.UseDeveloperExceptionPage();
            }

            app.UseRouting();

            app.UseEndpoints(endpoints =>
            {
                endpoints.MapGrpcService<CartService>();
                endpoints.MapGrpcService<cartservice.services.HealthCheckService>();

                endpoints.MapGet("/", async context =>
                {
                    await context.Response.WriteAsync("Communication with gRPC endpoints must be made through a gRPC client. To learn how to create a client, visit: https://go.microsoft.com/fwlink/?linkid=2086909");
                });
            });
        }
    }
}
