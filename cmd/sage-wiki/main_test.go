package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunCompile_RejectsBatchWatch(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("watch", false, "")
	cmd.Flags().Bool("batch", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("fresh", false, "")
	cmd.Flags().Bool("re-embed", false, "")
	cmd.Flags().Bool("re-extract", false, "")
	cmd.Flags().Bool("estimate", false, "")
	cmd.Flags().Bool("no-cache", false, "")
	cmd.Flags().Bool("prune", false, "")
	cmd.Flags().StringP("dir", "d", ".", "")
	cmd.Flags().StringP("output", "o", "", "")

	cmd.Flags().Set("watch", "true")
	cmd.Flags().Set("batch", "true")

	err := runCompile(cmd, []string{})
	if err == nil {
		t.Fatal("expected error when --batch and --watch are both set")
	}
	if !strings.Contains(err.Error(), "batch") || !strings.Contains(err.Error(), "watch") {
		t.Errorf("error should mention both 'batch' and 'watch', got: %s", err.Error())
	}
}
