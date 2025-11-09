# OpenTelemetry Setup Guide

This guide explains how to configure and use OpenTelemetry distributed tracing and metrics in the microservices-demo application.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Configuration](#configuration)
- [Service-Specific Implementation](#service-specific-implementation)
- [Deployment](#deployment)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)

---

## Overview

OpenTelemetry is implemented across all microservices to provide:
- **Distributed Tracing**: Track requests across service boundaries
- **Metrics Collection**: Monitor service performance and health
- **Unified Observability**: Consistent instrumentation across polyglot services

### Current Status

✅ **Fully Instrumented Services**:
- shippingservice (Go)
- productcatalogservice (Go)
- frontend (Go)
- checkoutservice (Go)
- adservice (Java)
- recommendationservice (Python) - pre-existing
- emailservice (Python) - pre-existing

---

## Architecture

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────┐
│  Services   │────▶│ OTLP Exporters   │────▶│  Collector  │
│ (Polyglot)  │     │  (gRPC/HTTP)     │     │  (port 4317)│
└─────────────┘     └──────────────────┘     └─────────────┘
                                                      │
                                                      ▼
                                              ┌─────────────┐
                                              │  Backends   │
                                              │ (Jaeger/    │
                                              │  Zipkin/    │
                                              │  Prometheus)│
                                              └─────────────┘
```

### Components

1. **Services**: Generate telemetry data (traces, metrics)
2. **OTLP Exporters**: Send data using OpenTelemetry Protocol
3. **Collector**: Receives, processes, and exports telemetry
4. **Backends**: Store and visualize telemetry data

---

## Configuration

### Environment Variables

All services support the following environment variables:

#### Required
- `COLLECTOR_SERVICE_ADDR` - OpenTelemetry Collector endpoint
  - Format: `hostname:port` (e.g., `otelcol:4317`)
  - Default: `localhost:4317`

#### Optional
- `DISABLE_TRACING` - Set to any value to disable tracing
- `DISABLE_STATS` - Set to any value to disable metrics collection
- `DISABLE_PROFILER` - Set to any value to disable Google Cloud Profiler

### Example Configuration

**Docker Compose**:
```yaml
environment:
  - COLLECTOR_SERVICE_ADDR=otelcol:4317
  - ENABLE_TRACING=1
```

**Kubernetes**:
```yaml
env:
- name: COLLECTOR_SERVICE_ADDR
  value: "opentelemetry-collector.observability:4317"
```

---

## Service-Specific Implementation

### Go Services

#### Common Pattern

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func initTracing() {
    collectorEndpoint := os.Getenv("COLLECTOR_SERVICE_ADDR")
    if collectorEndpoint == "" {
        collectorEndpoint = "localhost:4317"
    }

    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint(collectorEndpoint),
        otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()),
    )

    res, _ := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceName("servicename"),
            semconv.ServiceVersion("1.0.0"),
        ),
    )

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
    )

    otel.SetTracerProvider(tp)
}
```

#### gRPC Server Instrumentation

```go
import "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

srv := grpc.NewServer(
    grpc.StatsHandler(otelgrpc.NewServerHandler()),
)
```

#### Services Using This Pattern
- `shippingservice/main.go`
- `productcatalogservice/server.go`
- `checkoutservice/main.go`
- `frontend/main.go`

---

### Java Services

#### adservice Implementation

**Dependencies** (`build.gradle`):
```gradle
def openTelemetryVersion = "1.42.1"

implementation "io.opentelemetry:opentelemetry-api:${openTelemetryVersion}"
implementation "io.opentelemetry:opentelemetry-sdk:${openTelemetryVersion}"
implementation "io.opentelemetry:opentelemetry-exporter-otlp:${openTelemetryVersion}"
```

**Initialization**:
```java
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.sdk.OpenTelemetrySdk;
import io.opentelemetry.sdk.resources.Resource;
import io.opentelemetry.exporter.otlp.trace.OtlpGrpcSpanExporter;

private static void initTracing() {
    String collectorEndpoint = System.getenv("COLLECTOR_SERVICE_ADDR");
    if (collectorEndpoint == null || collectorEndpoint.isEmpty()) {
        collectorEndpoint = "localhost:4317";
    }

    Resource resource = Resource.getDefault()
        .merge(Resource.create(Attributes.of(
            ResourceAttributes.SERVICE_NAME, "adservice",
            ResourceAttributes.SERVICE_VERSION, "1.0.0"
        )));

    OtlpGrpcSpanExporter spanExporter = OtlpGrpcSpanExporter.builder()
        .setEndpoint("http://" + collectorEndpoint)
        .build();

    SdkTracerProvider sdkTracerProvider = SdkTracerProvider.builder()
        .addSpanProcessor(BatchSpanProcessor.builder(spanExporter).build())
        .setResource(resource)
        .build();

    OpenTelemetry openTelemetry = OpenTelemetrySdk.builder()
        .setTracerProvider(sdkTracerProvider)
        .buildAndRegisterGlobal();
}
```

---

### Python Services

#### Pre-existing Implementation

**recommendationservice** and **emailservice** already have OpenTelemetry configured.

**Dependencies** (`requirements.txt`):
```
opentelemetry-api
opentelemetry-sdk
opentelemetry-instrumentation-grpc
opentelemetry-exporter-otlp-proto-grpc
```

**Initialization**:
```python
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.grpc import GrpcInstrumentorServer

if os.environ.get("ENABLE_TRACING") == "1":
    trace.set_tracer_provider(TracerProvider())
    otel_endpoint = os.getenv("COLLECTOR_SERVICE_ADDR", "localhost:4317")
    trace.get_tracer_provider().add_span_processor(
        BatchSpanProcessor(
            OTLPSpanExporter(
                endpoint=otel_endpoint,
                insecure=True
            )
        )
    )

    grpc_server_instrumentor = GrpcInstrumentorServer()
    grpc_server_instrumentor.instrument()
```

---

## Deployment

### Local Development with Docker Compose

1. **Start OpenTelemetry Collector**:

Create `otel-collector-config.yaml`:
```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 1s

exporters:
  logging:
    loglevel: debug
  jaeger:
    endpoint: jaeger:14250
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [logging, jaeger]
```

2. **Add to docker-compose.yaml**:

```yaml
services:
  otelcol:
    image: otel/opentelemetry-collector:0.91.0
    command: ["--config=/etc/otel-collector-config.yaml"]
    volumes:
      - ./otel-collector-config.yaml:/etc/otel-collector-config.yaml
    ports:
      - "4317:4317"   # OTLP gRPC receiver
      - "4318:4318"   # OTLP HTTP receiver
      - "13133:13133" # health_check

  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686" # Jaeger UI
      - "14250:14250" # gRPC

  frontend:
    environment:
      - COLLECTOR_SERVICE_ADDR=otelcol:4317
    depends_on:
      - otelcol
```

3. **Access Jaeger UI**:
```
http://localhost:16686
```

---

### Kubernetes Deployment

1. **Deploy OpenTelemetry Operator** (recommended):

```bash
kubectl apply -f https://github.com/open-telemetry/opentelemetry-operator/releases/latest/download/opentelemetry-operator.yaml
```

2. **Create OpenTelemetry Collector**:

```yaml
apiVersion: opentelemetry.io/v1alpha1
kind: OpenTelemetryCollector
metadata:
  name: otel
  namespace: observability
spec:
  mode: deployment
  config: |
    receivers:
      otlp:
        protocols:
          grpc:
            endpoint: 0.0.0.0:4317

    processors:
      batch:

    exporters:
      logging:
        loglevel: debug
      jaeger:
        endpoint: jaeger-collector.observability:14250
        tls:
          insecure: true

    service:
      pipelines:
        traces:
          receivers: [otlp]
          processors: [batch]
          exporters: [logging, jaeger]
```

3. **Update Service Deployments**:

```yaml
env:
- name: COLLECTOR_SERVICE_ADDR
  value: "otel-collector.observability.svc.cluster.local:4317"
```

---

## Troubleshooting

### Traces Not Appearing

**Check 1: Collector Connectivity**
```bash
# Test collector endpoint
telnet otelcol 4317
```

**Check 2: Service Logs**
Look for initialization messages:
```
OpenTelemetry tracing initialized with collector at otelcol:4317
```

**Check 3: Collector Logs**
```bash
kubectl logs -n observability deployment/otel-collector
```

### Common Issues

#### "Failed to create trace exporter"
- **Cause**: Collector endpoint unreachable
- **Solution**: Verify COLLECTOR_SERVICE_ADDR is correct and collector is running

#### "Context deadline exceeded"
- **Cause**: Network timeout
- **Solution**: Check firewall rules and network policies

#### "Permission denied"
- **Cause**: TLS/SSL certificate issues
- **Solution**: Use `insecure.NewCredentials()` for development

### Debug Mode

Enable verbose logging:

**Go**:
```go
log.SetLevel(logrus.DebugLevel)
```

**Java**:
```java
System.setProperty("otel.traces.exporter", "logging")
```

**Python**:
```python
logging.basicConfig(level=logging.DEBUG)
```

---

## Best Practices

### Sampling Strategies

**Development**: `AlwaysSample()`
```go
sdktrace.WithSampler(sdktrace.AlwaysSample())
```

**Production**: Probabilistic or parent-based
```go
sdktrace.WithSampler(sdktrace.ParentBased(
    sdktrace.TraceIDRatioBased(0.1), // 10% sampling
))
```

### Resource Attributes

Always include:
- `service.name`: Identifies the service
- `service.version`: Tracks deployments
- `deployment.environment`: Separates dev/staging/prod

Example:
```go
resource.WithAttributes(
    semconv.ServiceName("checkoutservice"),
    semconv.ServiceVersion("1.2.3"),
    attribute.String("deployment.environment", "production"),
)
```

### Span Naming

Follow semantic conventions:
- HTTP: `HTTP {method}`
- gRPC: `{package}.{service}/{method}`
- Database: `{db.system} {db.operation}`

### Performance Considerations

1. **Use BatchSpanProcessor**: Reduces network calls
2. **Set reasonable timeouts**: Default 30s may be too long
3. **Monitor collector performance**: Add metrics and health checks
4. **Size resource attributes**: Avoid large payloads

---

## Advanced Topics

### Custom Spans

**Go**:
```go
tracer := otel.Tracer("myservice")
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()

span.SetAttributes(attribute.String("key", "value"))
```

### Baggage Propagation

```go
import "go.opentelemetry.io/otel/baggage"

member, _ := baggage.NewMember("user.id", "12345")
bag, _ := baggage.New(member)
ctx = baggage.ContextWithBaggage(ctx, bag)
```

### Metrics (Coming Soon)

Future implementation will include:
- Request rates
- Error rates
- Latency histograms
- Custom business metrics

---

## References

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [OpenTelemetry Java](https://opentelemetry.io/docs/languages/java/)
- [OpenTelemetry Python](https://opentelemetry.io/docs/languages/python/)
- [OTLP Specification](https://opentelemetry.io/docs/specs/otlp/)
- [Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/)

---

## Support

For issues or questions:
1. Check service logs for initialization errors
2. Verify collector configuration
3. Review [Troubleshooting](#troubleshooting) section
4. Consult OpenTelemetry documentation

---

**Last Updated**: November 2025
**Version**: 1.0.0
