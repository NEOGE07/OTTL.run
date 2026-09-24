package executor

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/ottlfuncs"

	"otel-ottl-preview/internal/config"
)

type Executor struct {
	parser ottl.Parser[*ottllog.TransformContext]
}

func New() (*Executor, error) {
	settings := component.TelemetrySettings{
		Logger: zap.NewNop(),
	}

	functions := ottlfuncs.StandardFuncs[*ottllog.TransformContext]()

	parser, err := ottllog.NewParser(
		functions,
		settings,
		//		ottllog.EnablePathContextNames(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTTL parser: %w", err)
	}

	return &Executor{
		parser: parser,
	}, nil
}

func (e *Executor) ExecuteLogs(
	logs plog.Logs,
	statements []config.OTTLStatement,
) error {

	for _, statement := range statements {

		if statement.Signal != "logs" {
			continue
		}

		parsed, err := e.parser.ParseStatement(statement.Code)
		if err != nil {
			return fmt.Errorf(
				"failed to parse OTTL statement %q: %w",
				statement.Code,
				err,
			)
		}

		sequence := ottllog.NewStatementSequence(
			[]*ottl.Statement[*ottllog.TransformContext]{
				parsed,
			},
			component.TelemetrySettings{
				Logger: zap.NewNop(),
			},
		)

		for i := 0; i < logs.ResourceLogs().Len(); i++ {
			resourceLogs := logs.ResourceLogs().At(i)

			for j := 0; j < resourceLogs.ScopeLogs().Len(); j++ {
				scopeLogs := resourceLogs.ScopeLogs().At(j)

				for k := 0; k < scopeLogs.LogRecords().Len(); k++ {
					logRecord := scopeLogs.LogRecords().At(k)

					transformContext := ottllog.NewTransformContextPtr(
						resourceLogs,
						scopeLogs,
						logRecord,
					)

					err := sequence.Execute(
						context.Background(),
						transformContext,
					)

					transformContext.Close()

					if err != nil {
						return fmt.Errorf(
							"failed to execute OTTL statement %q: %w",
							statement.Code,
							err,
						)
					}
				}
			}
		}
	}

	return nil
}
