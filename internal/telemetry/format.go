package telemetry

import (
	"bytes"
	"encoding/json"
)

// PrettyJSON formats JSON with indentation.
func PrettyJSON(data []byte) ([]byte, error) {
	var out bytes.Buffer

	if err := json.Indent(&out, data, "", "  "); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
