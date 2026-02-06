package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/instantly-cli/internal/agentfmt"
	"github.com/salmonumbrella/instantly-cli/internal/api"
	"github.com/salmonumbrella/instantly-cli/internal/filter"
	"github.com/salmonumbrella/instantly-cli/internal/outfmt"
)

var jsonMarshal = json.Marshal
var jsonUnmarshal = json.Unmarshal

func printResult(cmd *cobra.Command, kind string, resp any, meta map[string]any) error {
	mode := outfmt.ModeFrom(cmd.Context())

	jqExpr, err := effectiveJQExpression()
	if err != nil {
		return printError(cmd, kind, err, meta)
	}

	switch mode {
	case outfmt.Text:
		// For now: keep text output minimal and predictable.
		// Agents should use --output json/agent.
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "ok")
		return nil
	case outfmt.Agent:
		value := agentfmt.Envelope(kind, resp, meta)
		if jqExpr != "" {
			value, err = normalizeForJQ(value)
			if err == nil {
				value, err = filter.Apply(value, jqExpr)
			}
			if err != nil {
				return printError(cmd, kind, err, meta)
			}
		}
		return outfmt.PrintJSON(cmd.OutOrStdout(), value)
	case outfmt.JSONL:
		value := resp
		if jqExpr != "" {
			value, err = normalizeForJQ(value)
			if err == nil {
				value, err = filter.Apply(value, jqExpr)
			}
			if err != nil {
				return printError(cmd, kind, err, meta)
			}
		}
		return outfmt.PrintJSONL(cmd.OutOrStdout(), value)
	default:
		value := resp
		if jqExpr != "" {
			value, err = normalizeForJQ(value)
			if err == nil {
				value, err = filter.Apply(value, jqExpr)
			}
			if err != nil {
				return printError(cmd, kind, err, meta)
			}
		}
		return outfmt.PrintJSON(cmd.OutOrStdout(), value)
	}
}

func normalizeForJQ(v any) (any, error) {
	// gojq only supports JSON-compatible input types.
	// Use a JSON roundtrip to normalize structs and typed maps/slices.
	b, err := jsonMarshal(v)
	if err != nil {
		return nil, err
	}
	var out any
	if err := jsonUnmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func printError(cmd *cobra.Command, kind string, err error, meta map[string]any) error {
	mode := outfmt.ModeFrom(cmd.Context())
	if mode == outfmt.Text {
		// Text mode: print to stderr.
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
		return err
	}

	// Improve agent debuggability: include status code for API errors.
	var apiErr *api.APIError
	if errors.As(err, &apiErr) && apiErr != nil {
		if meta == nil {
			meta = map[string]any{}
		}
		meta["http_status"] = apiErr.Status
	}

	payload := map[string]any{
		"kind":  kind,
		"error": err.Error(),
	}
	if meta != nil {
		payload["meta"] = meta
	}

	// Best effort: emit structured error to stdout then return original error for exit code.
	// Even on failure, emitting JSON is useful for agents orchestrating retries.
	if mode == outfmt.JSONL {
		_ = outfmt.PrintJSONL(cmd.OutOrStdout(), payload)
	} else {
		_ = outfmt.PrintJSON(cmd.OutOrStdout(), payload)
	}
	return err
}
