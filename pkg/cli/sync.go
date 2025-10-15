package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync project configuration and types",
	Long:  `Generate TypeScript types and sync project metadata`,
	RunE:  runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	if !silent {
		fmt.Println("🔄 Syncing project...")
		fmt.Println("ℹ️  Type generation not yet implemented")
		fmt.Println("✅ Sync complete")
	}
	return nil
}
