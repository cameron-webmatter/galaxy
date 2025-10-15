package cli

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	previewPort int
	previewHost string
	previewOpen bool
)

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Preview your production build",
	Long:  `Preview your build locally before deploying`,
	RunE:  runPreview,
}

func init() {
	rootCmd.AddCommand(previewCmd)
	previewCmd.Flags().IntVar(&previewPort, "port", 4323, "port to run server on")
	previewCmd.Flags().StringVar(&previewHost, "host", "localhost", "host to bind to")
	previewCmd.Flags().BoolVar(&previewOpen, "open", false, "open browser on start")
}

func runPreview(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if rootDir != "" {
		cwd = rootDir
	}

	distDir := filepath.Join(cwd, "dist")
	if _, err := os.Stat(distDir); os.IsNotExist(err) {
		return fmt.Errorf("dist directory not found. Run 'galaxy build' first")
	}

	fs := http.FileServer(http.Dir(distDir))
	http.Handle("/", fs)

	addr := fmt.Sprintf("%s:%d", previewHost, previewPort)

	if !silent {
		fmt.Printf("🔍 Preview server running at http://%s\n", addr)
		fmt.Printf("📂 Serving: %s\n", distDir)
		fmt.Println("\nPress Ctrl+C to stop")
	}

	if previewOpen {
		go openBrowser(fmt.Sprintf("http://%s", addr))
	}

	return http.ListenAndServe(addr, nil)
}
