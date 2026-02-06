package filter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itchyny/gojq"
)

var jsonMarshal = json.Marshal
var jsonUnmarshal = json.Unmarshal

var runQuery = func(q *gojq.Query, data any) gojq.Iter { return q.Run(data) }

// NormalizeExpression fixes shell-escaped operators in jq expressions.
// Zsh escapes ! to \! even in single quotes, breaking operators like !=.
func NormalizeExpression(expr string) string {
	return strings.ReplaceAll(expr, `\!`, `!`)
}

// Apply applies a JQ filter expression to the input data.
func Apply(data any, expression string) (result any, err error) {
	if strings.TrimSpace(expression) == "" {
		return data, nil
	}

	// Normalize data to JSON-compatible types. gojq panics on unknown types.
	data, err = normalize(data)
	if err != nil {
		return nil, err
	}

	expression = NormalizeExpression(expression)
	query, err := gojq.Parse(expression)
	if err != nil {
		return nil, fmt.Errorf("invalid jq expression: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			// Convert gojq panics (usually "invalid type") into errors.
			err = fmt.Errorf("jq panic: %v", r)
			result = nil
		}
	}()

	iter := runQuery(query, data)
	var results []any
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, fmt.Errorf("jq error: %w", err)
		}
		results = append(results, v)
	}

	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}

func normalize(v any) (any, error) {
	// A JSON roundtrip normalizes structs and typed slices/maps into the
	// map[string]any / []any types that gojq supports.
	b, err := jsonMarshal(v)
	if err != nil {
		return nil, fmt.Errorf("normalize jq input: %w", err)
	}
	var out any
	if err := jsonUnmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("normalize jq input: %w", err)
	}
	return out, nil
}

// ApplyToJSON applies a JQ filter to JSON bytes and returns filtered JSON bytes.
func ApplyToJSON(jsonData []byte, expression string) ([]byte, error) {
	if strings.TrimSpace(expression) == "" {
		return jsonData, nil
	}

	var data any
	if err := jsonUnmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	filtered, err := Apply(data, expression)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(filtered, "", "  ")
}
