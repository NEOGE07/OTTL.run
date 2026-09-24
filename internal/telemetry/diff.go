package telemetry

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Change struct {
	Path   string
	Before any
	After  any
}

// DiffJSON finds meaningful changes between two JSON documents.
func DiffJSON(before, after []byte) ([]Change, error) {
	var b any
	var a any

	if err := json.Unmarshal(before, &b); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(after, &a); err != nil {
		return nil, err
	}

	var changes []Change

	findChanges(b, a, "", &changes)

	return changes, nil
}

func findChanges(
	before any,
	after any,
	path string,
	changes *[]Change,
) {
	switch b := before.(type) {

	case map[string]any:
		a, ok := after.(map[string]any)
		if !ok {
			*changes = append(*changes, Change{
				Path:   path,
				Before: before,
				After:  after,
			})
			return
		}

		for key, beforeValue := range b {
			afterValue, exists := a[key]

			nextPath := key
			if path != "" {
				nextPath = path + "." + key
			}

			if !exists {
				*changes = append(*changes, Change{
					Path:   nextPath,
					Before: beforeValue,
					After:  nil,
				})
				continue
			}

			findChanges(
				beforeValue,
				afterValue,
				nextPath,
				changes,
			)
		}

	case []any:
		a, ok := after.([]any)
		if !ok {
			*changes = append(*changes, Change{
				Path:   path,
				Before: before,
				After:  after,
			})
			return
		}

		max := len(b)
		if len(a) > max {
			max = len(a)
		}

		for i := 0; i < max; i++ {

			nextPath := fmt.Sprintf("%s[%d]", path, i)

			if i >= len(b) {
				*changes = append(*changes, Change{
					Path:   nextPath,
					Before: nil,
					After:  a[i],
				})
				continue
			}

			if i >= len(a) {
				*changes = append(*changes, Change{
					Path:   nextPath,
					Before: b[i],
					After:  nil,
				})
				continue
			}

			findChanges(
				b[i],
				a[i],
				nextPath,
				changes,
			)
		}

	default:
		if fmt.Sprintf("%v", before) != fmt.Sprintf("%v", after) {
			*changes = append(*changes, Change{
				Path:   path,
				Before: before,
				After:  after,
			})
		}
	}
}

// FormatChange converts a change into a readable BEFORE/AFTER block.
func FormatChange(change Change) string {
	var before strings.Builder
	var after strings.Builder

	writeJSONValue(&before, change.Before)
	writeJSONValue(&after, change.After)

	return fmt.Sprintf(
		"%s\n\nBEFORE\n%s\n\nAFTER\n%s",
		change.Path,
		before.String(),
		after.String(),
	)
}

func writeJSONValue(out *strings.Builder, value any) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fmt.Fprintf(out, "%v", value)
		return
	}

	out.Write(data)
}
