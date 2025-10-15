package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gastro/gastro/pkg/build"
	"github.com/spf13/cobra"
)

var (
	buildOutDir string
	buildMode   string
	buildForce  bool
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build your project for production",
	Long:  `Build your project and write it to disk`,
	RunE:  runBuild,
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVar(&buildOutDir, "outDir", "./dist", "output directory")
	buildCmd.Flags().StringVar(&buildMode, "mode", "production", "build mode")
	buildCmd.Flags().BoolVar(&buildForce, "force", false, "clear cache and rebuild")
}

func runBuild(cmd *cobra.Command, args []string) error {
	start := time.Now()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if rootDir != "" {
		cwd = rootDir
	}

	pagesDir := filepath.Join(cwd, "pages")
	publicDir := filepath.Join(cwd, "public")
	outDir := buildOutDir
	if !filepath.IsAbs(outDir) {
		outDir = filepath.Join(cwd, outDir)
	}

	if _, err := os.Stat(pagesDir); os.IsNotExist(err) {
		return fmt.Errorf("pages directory not found: %s", pagesDir)
	}

	if !silent {
		fmt.Println("🔨 Building for production...")
		if verbose {
			fmt.Printf("📁 Pages: %s\n", pagesDir)
			fmt.Printf("📦 Public: %s\n", publicDir)
			fmt.Printf("📤 Output: %s\n", outDir)
		}
		fmt.Println()
	}

	builder := build.NewSSGBuilder(pagesDir, outDir, publicDir)

	if err := builder.Build(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	duration := time.Since(start)
	
	if !silent {
		fmt.Printf("\n✅ Build complete in %v\n", duration.Round(time.Millisecond))
		fmt.Printf("📂 Output: %s\n", outDir)
	}

	return nil
}
