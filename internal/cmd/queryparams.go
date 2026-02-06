package cmd

import (
	"fmt"
	"net/url"
	"strings"
)

// applyQueryPairs merges key=value pairs into q. Later values overwrite earlier ones.
func applyQueryPairs(q url.Values, pairs []string) error {
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		k, v, ok := strings.Cut(pair, "=")
		if !ok || strings.TrimSpace(k) == "" {
			return fmt.Errorf("invalid --query %q (expected key=value)", pair)
		}
		q.Set(strings.TrimSpace(k), strings.TrimSpace(v))
	}
	return nil
}
