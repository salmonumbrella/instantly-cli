package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newTagsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "custom-tags",
		Aliases: []string{"tags", "tag"},
		Short:   "Manage custom tags and tag mappings",
	}

	cmd.AddCommand(newTagsListCmd())
	cmd.AddCommand(newTagsGetCmd())
	cmd.AddCommand(newTagsCreateCmd())
	cmd.AddCommand(newTagsUpdateCmd())
	cmd.AddCommand(newTagsDeleteCmd())
	cmd.AddCommand(newTagsToggleResourceCmd())

	cmd.AddCommand(newTagMappingsListCmd())

	return cmd
}

func newTagsListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		search        string
		queryPairs    []string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List custom tags (GET /custom-tags)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "custom_tags.list", err, nil)
			}
			q := url.Values{}
			if limit > 0 {
				q.Set("limit", fmt.Sprint(limit))
			}
			if strings.TrimSpace(startingAfter) != "" {
				q.Set("starting_after", strings.TrimSpace(startingAfter))
			}
			if strings.TrimSpace(search) != "" {
				q.Set("search", strings.TrimSpace(search))
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "custom_tags.list", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/custom-tags", q)
			if err != nil {
				return printError(cmd, "custom_tags.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "custom_tags.list", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination")
	cmd.Flags().StringVar(&search, "search", "", "Search")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")
	return cmd
}

func newTagsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <tag_id>",
		Short: "Get custom tag (GET /custom-tags/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "custom_tags.get", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "custom_tags.get", fmt.Errorf("tag_id is required"), nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/custom-tags/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "custom_tags.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "custom_tags.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newTagsCreateCmd() *cobra.Command {
	var (
		name     string
		color    string
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create custom tag (POST /custom-tags)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "custom_tags.create", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "custom_tags.create", err, nil)
			}
			// Instantly uses "label" for tag name; accept --name and legacy "name" input for convenience.
			if v, ok := body["name"]; ok && body["label"] == nil {
				body["label"] = v
				delete(body, "name")
			}
			if strings.TrimSpace(name) != "" {
				body["label"] = strings.TrimSpace(name)
			}
			if strings.TrimSpace(color) != "" {
				body["color"] = strings.TrimSpace(color)
			}
			labelVal, ok := body["label"].(string)
			if !ok || strings.TrimSpace(labelVal) == "" {
				return printError(cmd, "custom_tags.create", fmt.Errorf("--name is required (or set in --data-json/--data-file)"), nil)
			}

			resp, meta, err := client.PostJSON(cmdContext(cmd), "/custom-tags", nil, body)
			if err != nil {
				return printError(cmd, "custom_tags.create", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "custom_tags.create", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Tag name")
	cmd.Flags().StringVar(&color, "color", "", "Tag color")
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body (merged with flags)")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON body file, or '-' for stdin (merged with flags)")
	return cmd
}

func newTagsUpdateCmd() *cobra.Command {
	var (
		name     string
		color    string
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "update <tag_id>",
		Short: "Update custom tag (PATCH /custom-tags/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "custom_tags.update", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "custom_tags.update", fmt.Errorf("tag_id is required"), nil)
			}

			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "custom_tags.update", err, nil)
			}
			// Instantly uses "label" for tag name; accept "name" too for convenience.
			if v, ok := body["name"]; ok && body["label"] == nil {
				body["label"] = v
				delete(body, "name")
			}
			if strings.TrimSpace(name) != "" {
				body["label"] = strings.TrimSpace(name)
			}
			if strings.TrimSpace(color) != "" {
				body["color"] = strings.TrimSpace(color)
			}
			if len(body) == 0 {
				return printError(cmd, "custom_tags.update", fmt.Errorf("no fields to update"), nil)
			}

			resp, meta, err := client.PatchJSON(cmdContext(cmd), "/custom-tags/"+url.PathEscape(id), nil, body)
			if err != nil {
				return printError(cmd, "custom_tags.update", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "custom_tags.update", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Tag name")
	cmd.Flags().StringVar(&color, "color", "", "Tag color")
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON patch body (merged with flags)")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON patch body file, or '-' for stdin (merged with flags)")
	return cmd
}

func newTagsDeleteCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "delete <tag_id>",
		Short: "Delete custom tag (DELETE /custom-tags/{id}, requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "custom_tags.delete", fmt.Errorf("refusing to delete without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "custom_tags.delete", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "custom_tags.delete", fmt.Errorf("tag_id is required"), nil)
			}
			resp, meta, err := client.DeleteJSON(cmdContext(cmd), "/custom-tags/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "custom_tags.delete", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "custom_tags.delete", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive action")
	return cmd
}

func newTagsToggleResourceCmd() *cobra.Command {
	var (
		tagID      string
		resourceID string
		resource   string
		enabled    bool
		enabledSet bool
		dataJSON   string
		dataFile   string
	)
	cmd := &cobra.Command{
		Use:   "toggle-resource",
		Short: "Toggle tag mapping (POST /custom-tags/toggle-resource)",
		PreRun: func(cmd *cobra.Command, _ []string) {
			if cmd.Flags().Changed("enabled") {
				enabledSet = true
			}
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "custom_tags.toggle_resource", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "custom_tags.toggle_resource", err, nil)
			}
			if strings.TrimSpace(tagID) != "" {
				body["custom_tag_id"] = strings.TrimSpace(tagID)
			}
			if strings.TrimSpace(resourceID) != "" {
				body["resource_id"] = strings.TrimSpace(resourceID)
			}
			if strings.TrimSpace(resource) != "" {
				body["resource_type"] = strings.TrimSpace(resource)
			}
			if enabledSet {
				body["enabled"] = enabled
			}

			tagIDVal, ok := body["custom_tag_id"].(string)
			if !ok || strings.TrimSpace(tagIDVal) == "" {
				return printError(cmd, "custom_tags.toggle_resource", fmt.Errorf("--tag-id is required (or set custom_tag_id in body)"), nil)
			}
			resourceIDVal, ok := body["resource_id"].(string)
			if !ok || strings.TrimSpace(resourceIDVal) == "" {
				return printError(cmd, "custom_tags.toggle_resource", fmt.Errorf("--resource-id is required (or set resource_id in body)"), nil)
			}
			resourceTypeVal, ok := body["resource_type"].(string)
			if !ok || strings.TrimSpace(resourceTypeVal) == "" {
				return printError(cmd, "custom_tags.toggle_resource", fmt.Errorf("--resource-type is required (or set resource_type in body)"), nil)
			}

			resp, meta, err := client.PostJSON(cmdContext(cmd), "/custom-tags/toggle-resource", nil, body)
			if err != nil {
				return printError(cmd, "custom_tags.toggle_resource", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "custom_tags.toggle_resource", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&tagID, "tag-id", "", "Custom tag ID")
	cmd.Flags().StringVar(&resourceID, "resource-id", "", "Resource ID (entity ID)")
	cmd.Flags().StringVar(&resource, "resource-type", "", "Resource type (e.g. lead, campaign, account)")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable mapping (false to disable)")
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body (merged with flags)")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON body file, or '-' for stdin (merged with flags)")
	return cmd
}

func newTagMappingsListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		queryPairs    []string
	)
	cmd := &cobra.Command{
		Use:   "mappings",
		Short: "List custom tag mappings (GET /custom-tag-mappings)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "custom_tag_mappings.list", err, nil)
			}
			q := url.Values{}
			if limit > 0 {
				q.Set("limit", fmt.Sprint(limit))
			}
			if strings.TrimSpace(startingAfter) != "" {
				q.Set("starting_after", strings.TrimSpace(startingAfter))
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "custom_tag_mappings.list", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/custom-tag-mappings", q)
			if err != nil {
				return printError(cmd, "custom_tag_mappings.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "custom_tag_mappings.list", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")
	return cmd
}
