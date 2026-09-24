package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"otel-ottl-preview/internal/config"
	"otel-ottl-preview/internal/executor"
	"otel-ottl-preview/internal/telemetry"
)

func main() {
	configPath := flag.String(
		"config",
		"examples/collector.yaml",
		"Path to the OpenTelemetry Collector configuration",
	)

	dataPath := flag.String(
		"data",
		"examples/sample.json",
		"Path to the sample telemetry JSON",
	)

	flag.Parse()

	// --------------------------------
	// Load Collector configuration
	// --------------------------------

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	statements := config.ExtractStatements(cfg)

	fmt.Println("================================")
	fmt.Println("       OTTL Preview Tool")
	fmt.Println("================================")
	fmt.Println()

	// --------------------------------
	// Display OTTL statements
	// --------------------------------

	fmt.Println("OTTL Statements")
	fmt.Println("-------------------------------")

	if len(statements) == 0 {
		fmt.Println("No OTTL statements found.")
	} else {
		for i, statement := range statements {
			fmt.Printf("Statement %d\n", i+1)
			fmt.Printf("Signal:  %s\n", statement.Signal)
			fmt.Printf("Context: %s\n", statement.Context)
			fmt.Printf("OTTL:    %s\n", statement.Code)
			fmt.Println()
		}
	}

	// --------------------------------
	// Load telemetry
	// --------------------------------

	logs, err := telemetry.LoadLogs(*dataPath)
	if err != nil {
		log.Fatalf("failed to load telemetry: %v", err)
	}

	// --------------------------------
	// Create OTTL engine
	// --------------------------------

	engine, err := executor.New()
	if err != nil {
		log.Fatalf("failed to create OTTL engine: %v", err)
	}

	// --------------------------------
	// BEFORE
	// --------------------------------

	fmt.Println("BEFORE")
	fmt.Println("-------------------------------")

	before, err := telemetry.LogsToJSON(logs)
	if err != nil {
		log.Fatalf("failed to encode telemetry: %v", err)
	}

	prettyBefore, err := telemetry.PrettyJSON(before)
	if err != nil {
		log.Fatalf("failed to format BEFORE telemetry: %v", err)
	}

	fmt.Println(string(prettyBefore))

	// --------------------------------
	// EXECUTE OTTL
	// --------------------------------

	fmt.Println()
	fmt.Println("EXECUTION")
	fmt.Println("-------------------------------")

	for i, statement := range statements {
		if statement.Signal == "logs" {
			fmt.Printf("%d. %s\n", i+1, statement.Code)
		}
	}

	if err := engine.ExecuteLogs(logs, statements); err != nil {
		log.Fatalf("OTTL execution failed: %v", err)
	}

	// --------------------------------
	// AFTER
	// --------------------------------

	fmt.Println()
	fmt.Println("AFTER")
	fmt.Println("-------------------------------")

	after, err := telemetry.LogsToJSON(logs)
	if err != nil {
		log.Fatalf("failed to encode telemetry: %v", err)
	}

	prettyAfter, err := telemetry.PrettyJSON(after)
	if err != nil {
		log.Fatalf("failed to format AFTER telemetry: %v", err)
	}

	fmt.Println(string(prettyAfter))

	fmt.Println()
	fmt.Println("CHANGES")
	fmt.Println("────────────────────────────────────────")

	changes, err := telemetry.DiffJSON(before, after)
	if err != nil {
		log.Fatalf("failed to calculate changes: %v", err)
	}

	if len(changes) == 0 {
		fmt.Println("No changes.")
	} else {
		for _, change := range changes {
			fmt.Println()
			fmt.Println(change.Path)

			fmt.Println()
			fmt.Println("BEFORE")
			fmt.Println("-------------------------------")

			beforeValue, _ := json.MarshalIndent(
				change.Before,
				"",
				"  ",
			)

			fmt.Println(string(beforeValue))

			fmt.Println()
			fmt.Println("AFTER")
			fmt.Println("-------------------------------")

			afterValue, _ := json.MarshalIndent(
				change.After,
				"",
				"  ",
			)

			fmt.Println(string(afterValue))
		}
	}

	fmt.Println()
	fmt.Println("DIFF")
	fmt.Println("────────────────────────────────────────")

	fmt.Print(
		telemetry.UnifiedDiff(
			prettyBefore,
			prettyAfter,
		),
	)

	fmt.Println()
	fmt.Println("OTTL execution successful.")
}
