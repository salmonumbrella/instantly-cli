package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/instantly-cli/internal/api"
)

var stdinReader io.Reader = os.Stdin

func newAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "api",
		Aliases: []string{"raw"},
		Short:   "Low-level Instantly API access (escape hatch)",
	}

	cmd.AddCommand(newAPIMethodCmd("get"))
	cmd.AddCommand(newAPIMethodCmd("post"))
	cmd.AddCommand(newAPIMethodCmd("patch"))
	cmd.AddCommand(newAPIMethodCmd("delete"))

	return cmd
}

func newAPIMethodCmd(method string) *cobra.Command {
	var (
		queryPairs []string
		data       string
		dataFile   string
	)

	cmd := &cobra.Command{
		Use:   method + " <path>",
		Short: strings.ToUpper(method) + " an arbitrary Instantly API path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "api."+method, err, nil)
			}

			path := strings.TrimSpace(args[0])
			if path == "" {
				return printError(cmd, "api."+method, fmt.Errorf("path is required"), nil)
			}

			q := url.Values{}
			for _, pair := range queryPairs {
				pair = strings.TrimSpace(pair)
				if pair == "" {
					continue
				}
				k, v, ok := strings.Cut(pair, "=")
				if !ok || strings.TrimSpace(k) == "" {
					return printError(cmd, "api."+method, fmt.Errorf("invalid --query %q (expected key=value)", pair), nil)
				}
				q.Add(strings.TrimSpace(k), strings.TrimSpace(v))
			}

			var payload any
			if method == "post" || method == "patch" {
				raw, err := readJSONInput(data, dataFile)
				if err != nil {
					return printError(cmd, "api."+method, err, nil)
				}
				if raw != nil {
					if err := json.Unmarshal(raw, &payload); err != nil {
						return printError(cmd, "api."+method, fmt.Errorf("invalid JSON payload: %w", err), nil)
					}
				}
			}

			var resp any
			var meta *api.Meta
			switch method {
			case "get":
				resp, meta, err = client.GetJSON(cmdContext(cmd), path, q)
			case "post":
				resp, meta, err = client.PostJSON(cmdContext(cmd), path, q, payload)
			case "patch":
				resp, meta, err = client.PatchJSON(cmdContext(cmd), path, q, payload)
			case "delete":
				resp, meta, err = client.DeleteJSON(cmdContext(cmd), path, q)
			default:
				err = fmt.Errorf("unsupported method %q", method)
			}
			if err != nil {
				return printError(cmd, "api."+method, err, metaFrom(meta, nil))
			}

			return printResult(cmd, "api."+method, resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Query param (repeatable): key=value")
	cmd.Flags().StringVar(&data, "data", "", "JSON payload as a string (for post/patch)")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "JSON payload file path, or '-' for stdin (for post/patch)")
	return cmd
}

func readJSONInput(data, dataFile string) ([]byte, error) {
	if strings.TrimSpace(data) != "" && strings.TrimSpace(dataFile) != "" {
		return nil, fmt.Errorf("--data and --data-file cannot be used together")
	}
	if strings.TrimSpace(data) != "" {
		return []byte(data), nil
	}
	if strings.TrimSpace(dataFile) == "" {
		return nil, nil
	}
	if dataFile == "-" {
		b, err := io.ReadAll(stdinReader)
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		return b, nil
	}
	b, err := os.ReadFile(dataFile)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", dataFile, err)
	}
	return b, nil
}
