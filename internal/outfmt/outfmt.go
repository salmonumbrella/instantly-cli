package outfmt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

type Mode string

const (
	Text  Mode = "text"
	JSON  Mode = "json"
	JSONL Mode = "jsonl"
	Agent Mode = "agent"
)

type ctxKey int

const (
	modeKey ctxKey = iota
)

func WithMode(ctx context.Context, mode Mode) context.Context {
	return context.WithValue(ctx, modeKey, mode)
}

func ModeFrom(ctx context.Context) Mode {
	if v := ctx.Value(modeKey); v != nil {
		if m, ok := v.(Mode); ok {
			return m
		}
	}
	return JSON
}

func ParseMode(s string) (Mode, error) {
	switch Mode(s) {
	case "", JSON:
		return JSON, nil
	case Text, JSONL, Agent:
		return Mode(s), nil
	default:
		return "", fmt.Errorf("invalid --output %q (expected text, json, jsonl, agent)", s)
	}
}

func PrintJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func PrintJSONL(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	return enc.Encode(v)
}
