package cmd

import (
	"fmt"
	"regexp"
	"strings"
)

var fieldNameRe = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)*$`)

func effectiveJQExpression() (string, error) {
	if strings.TrimSpace(flags.JQ) != "" && strings.TrimSpace(flags.Fields) != "" {
		return "", fmt.Errorf("--jq and --fields cannot be used together")
	}
	if strings.TrimSpace(flags.JQ) != "" {
		return flags.JQ, nil
	}
	if strings.TrimSpace(flags.Fields) != "" {
		fields, err := parseFields(flags.Fields)
		if err != nil {
			return "", err
		}
		return buildFieldsQuery(fields), nil
	}
	return "", nil
}

func parseFields(input string) ([]string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty --fields")
	}

	raw := strings.Split(input, ",")
	out := make([]string, 0, len(raw))
	seen := map[string]bool{}
	for _, f := range raw {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if !fieldNameRe.MatchString(f) {
			return nil, fmt.Errorf("invalid field %q (expected: a, a_b, a.b, a.b_c)", f)
		}
		if !seen[f] {
			out = append(out, f)
			seen[f] = true
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no valid fields in --fields")
	}
	return out, nil
}

func buildFieldsQuery(fields []string) string {
	// Works for both:
	// - agent output envelope: {kind, items|item|data, meta}
	// - raw API output: {items, next_starting_after} etc.
	//
	// Agent envelopes keep their envelope keys; raw outputs are projected directly.
	objectExpr := buildObjectExpr(fields)
	return fmt.Sprintf(`
if (type == "object") then
  if (has("kind")) then
    if (has("items") and (.items|type=="array")) then
      {kind: .kind, meta: .meta, items: [.items[] | %s]}
    elif (has("item") and (.item|type=="object")) then
      {kind: .kind, meta: .meta, item: (.item | %s)}
    elif (has("data") and (.data|type=="object")) then
      {kind: .kind, meta: .meta, data: (.data | %s)}
    else
      %s
    end
  else
    if (has("items") and (.items|type=="array")) then
      [.items[] | %s]
    else
      %s
    end
  end
else
  .
end
`, objectExpr, objectExpr, objectExpr, objectExpr, objectExpr, objectExpr)
}

func buildObjectExpr(fields []string) string {
	parts := make([]string, 0, len(fields))
	for _, f := range fields {
		key := f
		if i := strings.LastIndex(f, "."); i >= 0 {
			key = f[i+1:]
		}
		parts = append(parts, fmt.Sprintf("%s: .%s", key, f))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}
