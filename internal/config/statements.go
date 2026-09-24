package config

// OTTLStatement represents an individual OTTL statement
// extracted from a Collector transform processor.
type OTTLStatement struct {
	Signal  string
	Context string
	Code    string
}

// ExtractStatements extracts OTTL statements from all
// transform processors in the Collector configuration.
func ExtractStatements(cfg *CollectorConfig) []OTTLStatement {
	var result []OTTLStatement

	for _, processor := range cfg.Processors {

		// Logs
		for _, group := range processor.LogStatements {
			for _, statement := range group.Statements {
				result = append(result, OTTLStatement{
					Signal:  "logs",
					Context: group.Context,
					Code:    statement,
				})
			}
		}

		// Metrics
		for _, group := range processor.MetricStatements {
			for _, statement := range group.Statements {
				result = append(result, OTTLStatement{
					Signal:  "metrics",
					Context: group.Context,
					Code:    statement,
				})
			}
		}

		// Traces
		for _, group := range processor.TraceStatements {
			for _, statement := range group.Statements {
				result = append(result, OTTLStatement{
					Signal:  "traces",
					Context: group.Context,
					Code:    statement,
				})
			}
		}
	}

	return result
}
