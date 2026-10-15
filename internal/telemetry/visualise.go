package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type jsonChange struct {
	Path        string
	Before      interface{}
	After       interface{}
	BeforeLine  int
	AfterLine   int
	BeforeExist bool
	AfterExist  bool
}

type treeNode struct {
	children map[string]*treeNode
	changed  bool
}

// PrintNumberedJSON displays JSON with line numbers.
// The original JSON data is not modified.
func PrintNumberedJSON(title string, data []byte) {
	fmt.Printf("\n%s\n", title)
	fmt.Println(strings.Repeat("-", len(title)))

	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	width := len(strconv.Itoa(len(lines)))

	for i, line := range lines {
		fmt.Printf("%*d | %s\n", width, i+1, line)
	}
}

// PrintChangeSummary compares two JSON documents and prints:
// 1. Changed JSON paths and their line numbers.
// 2. A nested tree showing where changes occurred.
func PrintChangeSummary(before, after []byte) error {
	var beforeValue interface{}
	var afterValue interface{}

	if err := json.Unmarshal(before, &beforeValue); err != nil {
		return fmt.Errorf("parse BEFORE JSON: %w", err)
	}
	if err := json.Unmarshal(after, &afterValue); err != nil {
		return fmt.Errorf("parse AFTER JSON: %w", err)
	}

	beforeLines, err := jsonPathLines(before)
	if err != nil {
		return fmt.Errorf("locate BEFORE JSON lines: %w", err)
	}
	afterLines, err := jsonPathLines(after)
	if err != nil {
		return fmt.Errorf("locate AFTER JSON lines: %w", err)
	}

	changes := make([]jsonChange, 0)
	compareValues(
		beforeValue,
		afterValue,
		"$",
		beforeLines,
		afterLines,
		&changes,
	)

	fmt.Println("\nCHANGE SUMMARY")
	fmt.Println("==================================================")

	if len(changes) == 0 {
		fmt.Println("No JSON fields changed.")
		fmt.Println("\nCHANGED ELEMENT TREE")
		fmt.Println("No changes to display.")
		return nil
	}

	for i, change := range changes {
		fmt.Printf("\n%d. %s\n", i+1, change.Path)

		if change.BeforeExist {
			fmt.Printf("   BEFORE line %d: %s\n",
				change.BeforeLine, formatValue(change.Before))
		} else {
			fmt.Println("   BEFORE: <field did not exist>")
		}

		if change.AfterExist {
			fmt.Printf("   AFTER  line %d: %s\n",
				change.AfterLine, formatValue(change.After))
		} else {
			fmt.Println("   AFTER: <field removed>")
		}
	}

	fmt.Println("\nCHANGED ELEMENT TREE")
	fmt.Println("==================================================")

	root := &treeNode{children: make(map[string]*treeNode)}

	for _, change := range changes {
		insertTreePath(root, change.Path)
	}

	printTree(root, "", true, "$", true)

	return nil
}

func compareValues(
	before, after interface{},
	path string,
	beforeLines, afterLines map[string]int,
	changes *[]jsonChange,
) {
	beforeMap, beforeIsMap := before.(map[string]interface{})
	afterMap, afterIsMap := after.(map[string]interface{})

	if beforeIsMap && afterIsMap {
		keys := make(map[string]bool)
		for key := range beforeMap {
			keys[key] = true
		}
		for key := range afterMap {
			keys[key] = true
		}

		sortedKeys := make([]string, 0, len(keys))
		for key := range keys {
			sortedKeys = append(sortedKeys, key)
		}
		sort.Strings(sortedKeys)

		for _, key := range sortedKeys {
			oldValue, oldExists := beforeMap[key]
			newValue, newExists := afterMap[key]
			childPath := path + "." + escapePathKey(key)

			switch {
			case oldExists && newExists:
				compareValues(
					oldValue, newValue, childPath,
					beforeLines, afterLines, changes,
				)
			case oldExists:
				*changes = append(*changes, jsonChange{
					Path:        childPath,
					Before:      oldValue,
					BeforeLine:  beforeLines[childPath],
					AfterLine:   afterLines[childPath],
					BeforeExist: true,
				})
			case newExists:
				*changes = append(*changes, jsonChange{
					Path:       childPath,
					After:      newValue,
					BeforeLine: beforeLines[childPath],
					AfterLine:  afterLines[childPath],
					AfterExist: true,
				})
			}
		}
		return
	}

	beforeArray, beforeIsArray := before.([]interface{})
	afterArray, afterIsArray := after.([]interface{})

	if beforeIsArray && afterIsArray {
		maxLen := len(beforeArray)
		if len(afterArray) > maxLen {
			maxLen = len(afterArray)
		}

		for i := 0; i < maxLen; i++ {
			childPath := fmt.Sprintf("%s[%d]", path, i)

			switch {
			case i < len(beforeArray) && i < len(afterArray):
				compareValues(
					beforeArray[i], afterArray[i], childPath,
					beforeLines, afterLines, changes,
				)
			case i < len(beforeArray):
				*changes = append(*changes, jsonChange{
					Path:        childPath,
					Before:      beforeArray[i],
					BeforeLine:  beforeLines[childPath],
					AfterLine:   afterLines[childPath],
					BeforeExist: true,
				})
			default:
				*changes = append(*changes, jsonChange{
					Path:       childPath,
					After:      afterArray[i],
					BeforeLine: beforeLines[childPath],
					AfterLine:  afterLines[childPath],
					AfterExist: true,
				})
			}
		}
		return
	}

	if !equalJSONValues(before, after) {
		*changes = append(*changes, jsonChange{
			Path:        path,
			Before:      before,
			After:       after,
			BeforeLine:  beforeLines[path],
			AfterLine:   afterLines[path],
			BeforeExist: true,
			AfterExist:  true,
		})
	}
}

func equalJSONValues(a, b interface{}) bool {
	aJSON, errA := json.Marshal(a)
	bJSON, errB := json.Marshal(b)
	return errA == nil && errB == nil && bytes.Equal(aJSON, bJSON)
}

func formatValue(value interface{}) string {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(data)
}

func escapePathKey(key string) string {
	// Dot notation is convenient for ordinary OTel attribute names.
	// Quote keys containing dots or spaces to avoid ambiguous paths.
	if strings.ContainsAny(key, ". []\"") {
		quoted, _ := json.Marshal(key)
		return "[" + string(quoted) + "]"
	}
	return key
}

// jsonPathLines associates JSON paths with their line numbers in
// the pretty-printed JSON. Object fields map to their key lines.
func jsonPathLines(data []byte) (map[string]int, error) {
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}

	// The decoder walks the original formatted document so that
	// line numbers refer to the displayed BEFORE/AFTER JSON.
	decoder := json.NewDecoder(bytes.NewReader(data))
	lines := make(map[string]int)

	if err := walkJSON(decoder, data, "$", lines); err != nil {
		return nil, err
	}
	return lines, nil
}

func walkJSON(
	decoder *json.Decoder,
	data []byte,
	path string,
	lines map[string]int,
) error {
	start := decoder.InputOffset()
	token, err := decoder.Token()
	if err != nil {
		return err
	}

	switch token {
	case json.Delim('{'):
		lines[path] = lineAt(data, start)

		for decoder.More() {
			keyOffset := decoder.InputOffset()
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}

			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("expected JSON object key")
			}

			childPath := path + "." + escapePathKey(key)
			lines[childPath] = lineAt(data, keyOffset)

			if err := walkJSON(decoder, data, childPath, lines); err != nil {
				return err
			}
		}

		_, err = decoder.Token() // closing }
		return err

	case json.Delim('['):
		lines[path] = lineAt(data, start)
		index := 0

		for decoder.More() {
			childPath := fmt.Sprintf("%s[%d]", path, index)
			if err := walkJSON(decoder, data, childPath, lines); err != nil {
				return err
			}
			index++
		}

		_, err = decoder.Token() // closing ]
		return err

	default:
		lines[path] = lineAt(data, start)
		return nil
	}
}

func lineAt(data []byte, offset int64) int {
	if offset < 0 {
		return 1
	}
	if offset > int64(len(data)) {
		offset = int64(len(data))
	}
	return bytes.Count(data[:offset], []byte("\n")) + 1
}

func insertTreePath(root *treeNode, path string) {
	current := root

	for _, part := range splitJSONPath(path) {
		if current.children[part] == nil {
			current.children[part] = &treeNode{
				children: make(map[string]*treeNode),
			}
		}
		current = current.children[part]
	}
	current.changed = true
}

func splitJSONPath(path string) []string {
	parts := make([]string, 0)
	var current strings.Builder

	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '$':
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		case '.':
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		case '[':
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			j := i
			for j < len(path) && path[j] != ']' {
				j++
			}
			if j < len(path) {
				parts = append(parts, path[i:j+1])
				i = j
			}
		default:
			current.WriteByte(path[i])
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

func printTree(
	node *treeNode,
	prefix string,
	last bool,
	label string,
	isRoot bool,
) {
	if !isRoot {
		connector := "├── "
		if last {
			connector = "└── "
		}

		marker := ""
		if node.changed {
			marker = "  [CHANGED]"
		}

		fmt.Printf("%s%s%s%s\n", prefix, connector, label, marker)

		if last {
			prefix += "    "
		} else {
			prefix += "│   "
		}
	}

	keys := make([]string, 0, len(node.children))
	for key := range node.children {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for i, key := range keys {
		printTree(
			node.children[key],
			prefix,
			i == len(keys)-1,
			key,
			false,
		)
	}
}
