package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/BurntSushi/toml"
	"github.com/galaxy/galaxy/pkg/config"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [integration]",
	Short: "Add an integration to your project",
	Long:  `Add integrations like frameworks, adapters, or features`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

var availableIntegrations = []string{
	"react",
	"vue",
	"svelte",
	"tailwind",
	"sitemap",
}

func runAdd(cmd *cobra.Command, args []string) error {
	var integration string

	if len(args) > 0 {
		integration = args[0]
	} else {
		prompt := &survey.Select{
			Message: "Select an integration:",
			Options: availableIntegrations,
		}
		if err := survey.AskOne(prompt, &integration); err != nil {
			return err
		}
	}

	fmt.Printf("\n📦 Adding %s integration...\n", integration)

	switch integration {
	case "react", "vue", "svelte":
		return addFramework(integration)
	case "tailwind":
		return addTailwind()
	case "sitemap":
		return addSitemap()
	default:
		return fmt.Errorf("unknown integration: %s", integration)
	}
}

func addFramework(framework string) error {
	fmt.Printf("ℹ️  Framework integrations not yet implemented\n")
	fmt.Printf("   This would install %s support for component islands\n", framework)
	return nil
}

func addTailwind() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if rootDir != "" {
		cwd = rootDir
	}

	configPath := filepath.Join(cwd, "galaxy.config.toml")
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	pkgManager := cfg.PackageManager
	if pkgManager == "" {
		pkgManager = detectPackageManager(cwd)
	}

	packageJSONPath := filepath.Join(cwd, "package.json")
	if _, err := os.Stat(packageJSONPath); os.IsNotExist(err) {
		fmt.Println("Creating package.json...")
		cmd := exec.Command(pkgManager, "init", "-y")
		cmd.Dir = cwd
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create package.json: %w", err)
		}
	}

	fmt.Println("Installing Tailwind CSS...")

	var cmd *exec.Cmd
	if pkgManager == "pnpm" {
		cmd = exec.Command(pkgManager, "add", "-D", "tailwindcss")
	} else {
		cmd = exec.Command(pkgManager, "install", "-D", "tailwindcss")
	}
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Println("Initializing Tailwind config...")
	var initCmd *exec.Cmd
	if pkgManager == "pnpm" || pkgManager == "yarn" || pkgManager == "bun" {
		initCmd = exec.Command(pkgManager, "exec", "tailwindcss", "init")
	} else {
		initCmd = exec.Command("npx", "tailwindcss", "init")
	}
	cmd = initCmd
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	hasPlugin := false
	for _, p := range cfg.Plugins {
		if p.Name == "tailwindcss" {
			hasPlugin = true
			break
		}
	}

	if !hasPlugin {
		cfg.Plugins = append(cfg.Plugins, config.PluginConfig{
			Name:   "tailwindcss",
			Config: make(map[string]interface{}),
		})

		f, err := os.Create(configPath)
		if err != nil {
			return fmt.Errorf("open config: %w", err)
		}
		defer f.Close()

		if err := toml.NewEncoder(f).Encode(cfg); err != nil {
			return fmt.Errorf("write config: %w", err)
		}

		fmt.Println("  ✓ Added tailwindcss plugin to galaxy.config.toml")
	}

	fmt.Println("\n✅ Tailwind CSS added!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Configure tailwind.config.js")
	fmt.Println("  2. Add Tailwind directives to your CSS")
	return nil
}

func detectPackageManager(cwd string) string {
	if _, err := os.Stat(filepath.Join(cwd, "pnpm-lock.yaml")); err == nil {
		return "pnpm"
	}
	if _, err := os.Stat(filepath.Join(cwd, "yarn.lock")); err == nil {
		return "yarn"
	}
	if _, err := os.Stat(filepath.Join(cwd, "bun.lockb")); err == nil {
		return "bun"
	}
	return "npm"
}

func addSitemap() error {
	fmt.Println("ℹ️  Sitemap integration not yet implemented")
	fmt.Println("   This would generate sitemap.xml during build")
	return nil
}
