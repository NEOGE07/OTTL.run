package telemetry

import (
	"strings"
)

type diffLine struct {
	Prefix string
	Text   string
}

func UnifiedDiff(before, after []byte) string {
	beforeLines := strings.Split(string(before), "\n")
	afterLines := strings.Split(string(after), "\n")

	lines := calculateLineDiff(beforeLines, afterLines)

	var out strings.Builder

	for _, line := range lines {
		out.WriteString(line.Prefix)
		out.WriteString(line.Text)
		out.WriteString("\n")
	}

	return out.String()
}

func calculateLineDiff(before, after []string) []diffLine {
	m := len(before)
	n := len(after)

	// LCS table.
	dp := make([][]int, m+1)

	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := m - 1; i >= 0; i-- {
		for j := n - 1; j >= 0; j-- {

			if before[i] == after[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	var result []diffLine

	i := 0
	j := 0

	for i < m && j < n {

		if before[i] == after[j] {
			result = append(result, diffLine{
				Prefix: "  ",
				Text:   before[i],
			})

			i++
			j++

		} else if dp[i+1][j] >= dp[i][j+1] {
			result = append(result, diffLine{
				Prefix: "- ",
				Text:   before[i],
			})

			i++

		} else {
			result = append(result, diffLine{
				Prefix: "+ ",
				Text:   after[j],
			})

			j++
		}
	}

	for i < m {
		result = append(result, diffLine{
			Prefix: "- ",
			Text:   before[i],
		})

		i++
	}

	for j < n {
		result = append(result, diffLine{
			Prefix: "+ ",
			Text:   after[j],
		})

		j++
	}

	return result
}
