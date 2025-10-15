package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/AlecAivazis/survey/v2"
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
	fmt.Println("Installing Tailwind CSS...")
	
	cmd := exec.Command("npm", "install", "-D", "tailwindcss")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("npx", "tailwindcss", "init")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Println("\n✅ Tailwind CSS added!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Configure tailwind.config.js")
	fmt.Println("  2. Add Tailwind directives to your CSS")
	return nil
}

func addSitemap() error {
	fmt.Println("ℹ️  Sitemap integration not yet implemented")
	fmt.Println("   This would generate sitemap.xml during build")
	return nil
}
