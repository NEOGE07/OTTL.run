package main

import (
	"bufio"
	//"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"otel-ottl-preview/internal/config"
	"otel-ottl-preview/internal/executor"
	"otel-ottl-preview/internal/telemetry"
)

func main() {
	// --------------------------------
	// Command-line flags
	// --------------------------------

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

	fmt.Println("========================================")
	fmt.Println("          OTTL Preview Tool")
	fmt.Println("========================================")
	fmt.Println()

	fmt.Printf("Configuration: %s\n", *configPath)
	fmt.Printf("Sample data:   %s\n", *dataPath)
	fmt.Println()

	// --------------------------------
	// Display OTTL statements
	// --------------------------------

	fmt.Println("OTTL STATEMENTS")
	fmt.Println("----------------------------------------")

	if len(statements) == 0 {
		fmt.Println("No OTTL statements found.")
	} else {
		for i, statement := range statements {
			fmt.Printf("[%d] Signal:  %s\n", i+1, statement.Signal)
			fmt.Printf("    Context: %s\n", statement.Context)
			fmt.Printf("    OTTL:    %s\n", statement.Code)
			fmt.Println()
		}
	}

	// --------------------------------
	// Load sample telemetry
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
	// Capture BEFORE state
	// --------------------------------

	before, err := telemetry.LogsToJSON(logs)
	if err != nil {
		log.Fatalf("failed to encode BEFORE telemetry: %v", err)
	}

	prettyBefore, err := telemetry.PrettyJSON(before)
	if err != nil {
		log.Fatalf("failed to format BEFORE telemetry: %v", err)
	}

	// --------------------------------
	// Execute OTTL
	// --------------------------------

	fmt.Println("EXECUTION")
	fmt.Println("----------------------------------------")

	executableCount := 0

	for i, statement := range statements {
		if statement.Signal == "logs" {
			fmt.Printf("[%d] %s\n", i+1, statement.Code)
			executableCount++
		}
	}

	if executableCount == 0 {
		fmt.Println("No log OTTL statements to execute.")
	}

	if err := engine.ExecuteLogs(logs, statements); err != nil {
		log.Fatalf("OTTL execution failed: %v", err)
	}

	// --------------------------------
	// Capture AFTER state
	// --------------------------------

	after, err := telemetry.LogsToJSON(logs)
	if err != nil {
		log.Fatalf("failed to encode AFTER telemetry: %v", err)
	}

	prettyAfter, err := telemetry.PrettyJSON(after)
	if err != nil {
		log.Fatalf("failed to format AFTER telemetry: %v", err)
	}

	// --------------------------------
	// Calculate changes
	// --------------------------------
	/*
		changes, err := telemetry.DiffJSON(before, after)
		if err != nil {
			log.Fatalf("failed to calculate changes: %v", err)
		}




		// --------------------------------
		// Display concise summary
		// --------------------------------

		fmt.Println()
		fmt.Println("CHANGES")
		fmt.Println("========================================")

		if len(changes) == 0 {
			fmt.Println("No changes detected.")
		} else {
			fmt.Printf("Changed fields: %d\n", len(changes))

			for _, change := range changes {
				fmt.Println()
				fmt.Println(change.Path)

				beforeValue, err := json.Marshal(change.Before)
				if err != nil {
					log.Fatalf("failed to encode previous value: %v", err)
				}

				afterValue, err := json.Marshal(change.After)
				if err != nil {
					log.Fatalf("failed to encode new value: %v", err)
				}

				fmt.Printf("  - %s\n", beforeValue)
				fmt.Printf("  + %s\n", afterValue)
			}
		}
	*/

	if err := telemetry.PrintChangeSummary(prettyBefore, prettyAfter); err != nil {
		log.Fatalf("failed to compare telemetry: %v", err)
	}

	// --------------------------------
	// Ask whether to show full JSON
	// --------------------------------

	fmt.Println()
	fmt.Println("----------------------------------------")
	fmt.Print("View full BEFORE and AFTER telemetry? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)

	answer, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		log.Fatalf("failed to read user input: %v", err)
	}

	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer == "y" || answer == "yes" {
		telemetry.PrintNumberedJSON("BEFORE JSON", prettyBefore)
		telemetry.PrintNumberedJSON("AFTER JSON", prettyAfter)

		/*
			fmt.Println()
			fmt.Println("BEFORE TRANSFORMATION")
			fmt.Println("========================================")
			fmt.Println(string(prettyBefore))

			fmt.Println()
			fmt.Println("AFTER TRANSFORMATION")
			fmt.Println("========================================")
			fmt.Println(string(prettyAfter))
		*/
	} else {
		fmt.Println("Full telemetry output skipped.")
	}

	// --------------------------------
	// Finish
	// --------------------------------

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("OTTL preview completed successfully.")
	fmt.Println("========================================")
}
