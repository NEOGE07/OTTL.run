# OpenTelemetry Collector Configuration
This README is a summary of the OpenTelemetry (OTel) Collector Documentation. For more detailed information, refer to the official documentation in the [OTel website](https://opentelemetry.io/docs/collector/configuration/):

> https://opentelemetry.io/docs/collector/configuration/

<br>


## Basic OpenTelemetry Documentation


OpenTelemetry (OTel) Collectors have have **4 main** configuration categories: 
1. [**Receivers**](#1-receivers): Configure how to receive data from tenant applications

2. [**Processors**](#2-processors): Configure signal transformation and/or additional enriching

3. [**Exporters**](#3-exporters): Configure how to send data to backend Observability platforms 

4. [**Connectors**](#4-connectors): Configure collector to connect different pipelines together 

<br>

## Basic Configuration Example

This segment serves as as **introductory example** of a basic OTel collector. <br>
More in-depth configuration details will be covered in later segments. 

The main segments in **this example** are: 
- ```receivers```
- ```exporters```
- ```extensions```
- ```service```

### Receivers
- Can configure multiple receivers of same type **(e.g. otlp)**
- Each receiver can have multiple receiving endpoints **(e.g. http & grpc)**
- Can configure endpoint based on protocol **(e.g. http: Port 4318)**
```
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318
  otlp/2:
    protocols:
      grpc:
        endpoint: 0.0.0.0:55690
```

### Exporters
- Can have multiple export channels by configuring endpoint and port number
```
exporters:
  otlp_grpc:
    endpoint: otelcol:4317
    sending_queue:
      batch:
  otlp_grpc/2:
    endpoint: otelcol2:4317
    sending_queue:
      batch:
```

### Extensions
- **Optional & additional components** that provide more info on collector (e.g. health check)
```
extensions:
  health_check:
    endpoint: 0.0.0.0:13133
  pprof:
    endpoint: 0.0.0.0:1777
  zpages:
    endpoint: 0.0.0.0:55679
```

### Services
- Configure signal types within collector, namely
  - **Logs**: Structured event records that took place
  - **Metrics**: Specific values for processes (e.g. CPU usage)
  - **Traces**: End-to-end process tracking
- Configure receivers and exporters for each category of OTel signals
```
service:
  extensions: [health_check, pprof, zpages]
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [otlp_grpc]
    traces/2:
      receivers: [otlp/2]
      exporters: [otlp_grpc/2]
    metrics:
      receivers: [otlp]
      exporters: [otlp_grpc]
    logs:
      receivers: [otlp]
      exporters: [otlp_grpc]
```
<br><br>




# Specific Configuration Details  

## 1. Receivers 

> https://opentelemetry.io/docs/collector/configuration/#receivers



## 2. Processors 

> https://opentelemetry.io/docs/collector/configuration/#processors



## 3. Exporters 

> https://opentelemetry.io/docs/collector/configuration/#exporters



## 4. Connectors

> https://opentelemetry.io/docs/collector/configuration/#connectors





