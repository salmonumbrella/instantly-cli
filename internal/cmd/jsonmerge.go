package cmd

import (
	"encoding/json"
	"fmt"
)

func readJSONObjectInput(dataJSON, dataFile string) (map[string]any, error) {
	raw, err := readJSONInput(dataJSON, dataFile)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return map[string]any{}, nil
	}

	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected a JSON object")
	}
	return m, nil
}

func mergeMaps(dst, src map[string]any) {
	for k, v := range src {
		if vMap, ok := v.(map[string]any); ok {
			if existing, ok := dst[k].(map[string]any); ok {
				mergeMaps(existing, vMap)
				dst[k] = existing
				continue
			}
		}
		dst[k] = v
	}
}
