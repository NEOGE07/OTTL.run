package config

// CollectorConfig represents the parts of an
// OpenTelemetry Collector configuration that
// our preview tool currently cares about.
type CollectorConfig struct {
	Processors map[string]TransformProcessor `yaml:"processors"`
}

// TransformProcessor represents an OTTL transform
// processor configuration.
type TransformProcessor struct {
	ErrorMode        string           `yaml:"error_mode"`
	LogStatements    []StatementGroup `yaml:"log_statements"`
	MetricStatements []StatementGroup `yaml:"metric_statements"`
	TraceStatements  []StatementGroup `yaml:"trace_statements"`
}

// StatementGroup represents one OTTL statement group.
//
// Example:
//
//   - context: log
//     statements:
//   - set(attributes["environment"], "production")
type StatementGroup struct {
	Context    string   `yaml:"context"`
	Statements []string `yaml:"statements"`
}
