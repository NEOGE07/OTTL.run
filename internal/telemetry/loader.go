package telemetry

import (
	"fmt"
	"os"

	"go.opentelemetry.io/collector/pdata/plog"
)

// LoadLogs reads an OTLP/JSON log payload from a file.
func LoadLogs(path string) (plog.Logs, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return plog.Logs{}, err
	}

	var unmarshaler plog.JSONUnmarshaler

	logs, err := unmarshaler.UnmarshalLogs(data)
	if err != nil {
		return plog.Logs{}, err
	}

	fmt.Printf("Loaded ResourceLogs: %d\n", logs.ResourceLogs().Len())

	return logs, nil
}

// LogsToJSON converts pdata logs back into OTLP/JSON.
func LogsToJSON(logs plog.Logs) ([]byte, error) {
	var marshaler plog.JSONMarshaler

	return marshaler.MarshalLogs(logs)
}
