package prompts

import (
	"github.com/AlecAivazis/survey/v2"
)

type ProjectConfig struct {
	Name           string
	Template       string
	InstallDeps    bool
	InitGit        bool
	PackageManager string
}

func AskProjectDetails(defaultName string) (*ProjectConfig, error) {
	config := &ProjectConfig{}

	questions := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Project name:",
				Default: defaultName,
			},
			Validate: survey.Required,
		},
		{
			Name: "template",
			Prompt: &survey.Select{
				Message: "Select a template:",
				Options: []string{
					"minimal",
					"blog",
					"portfolio",
					"documentation",
				},
				Default: "minimal",
			},
		},
		{
			Name: "packageManager",
			Prompt: &survey.Select{
				Message: "Package manager:",
				Options: []string{"npm", "pnpm", "yarn", "bun"},
				Default: "npm",
			},
		},
		{
			Name: "installDeps",
			Prompt: &survey.Confirm{
				Message: "Install dependencies?",
				Default: true,
			},
		},
		{
			Name: "initGit",
			Prompt: &survey.Confirm{
				Message: "Initialize git repository?",
				Default: true,
			},
		},
	}

	err := survey.Ask(questions, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
