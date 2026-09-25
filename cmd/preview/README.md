# Instructions to run ottl.run --> local package to validate otel configs and data

1. edit otel config yaml inside ./examples. How to configure dummy data? Telemetry gen??
2. once ready to test, run these commands in sequence (within the root project dir):
> go fmt ./...
> go build ./...
> go run ./cmd/preview/
3. Will be able to see output, cross check changes and validate if settings are correct
