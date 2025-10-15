package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Open documentation in browser",
	Long:  `Open Gastro documentation in your web browser`,
	RunE:  runDocs,
}

func init() {
	rootCmd.AddCommand(docsCmd)
}

func runDocs(cmd *cobra.Command, args []string) error {
	url := "https://github.com/gastro/gastro"
	
	if !silent {
		fmt.Printf("📚 Opening docs: %s\n", url)
	}
	
	openBrowser(url)
	return nil
}
