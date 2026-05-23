package main

import (
	"fmt"
	"os"
	"strings"

	"charm.land/huh/v2"
	"github.com/cabljac/huhx"
	"github.com/spf13/cobra"
)

func main() {
	var (
		name        string
		environment string
		notes       string
		regions     []string
		allRegions  bool
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy an app",
		RunE: func(cmd *cobra.Command, args []string) error {
			form := huhx.NewForm(
				huhx.NewGroup(
					huhx.NewInput().
						Key("name").
						Title("App name").
						Value(&name).
						Validate(func(s string) error {
							if s == "" {
								return fmt.Errorf("name is required")
							}
							return nil
						}),
					huhx.NewSelect[string]().
						Key("environment").
						Title("Target environment").
						Options(
							huh.NewOption("staging", "staging"),
							huh.NewOption("prod", "prod"),
						).
						Value(&environment),
				),
				huhx.NewGroup(
					huhx.NewText().
						Key("notes").
						Title("Release notes").
						Value(&notes).
						Optional(),
					huhx.NewMultiSelect[string]().
						Key("regions").
						Title("Target regions").
						Options(
							huh.NewOption("us-east-1", "us-east-1"),
							huh.NewOption("us-west-2", "us-west-2"),
							huh.NewOption("eu-west-1", "eu-west-1"),
						).
						Value(&regions),
				),
				huhx.NewGroup(
					huhx.NewConfirm().
						Key("all-regions").
						Title("Deploy to all regions?").
						Value(&allRegions),
				).WithHideFunc(func() bool {
					return environment != "prod"
				}),
			)

			answerFile, _ := cmd.Flags().GetString("answer-file")
			runner := huhx.New(form,
				huhx.WithEnvPrefix("DEPLOY"),
				huhx.WithCobraFlags(cmd),
				huhx.WithAnswerFile(answerFile),
			)
			if err := runner.Run(); err != nil {
				return err
			}

			fmt.Printf("Deploying %q to %s (regions: %s; all regions: %v)\n",
				name, environment, strings.Join(regions, ","), allRegions)
			if notes != "" {
				fmt.Printf("Notes: %s\n", notes)
			}
			return nil
		},
	}

	flags := cmd.Flags()
	flags.String("name", "", "app name")
	flags.String("environment", "", "target environment (staging|prod)")
	flags.String("notes", "", "release notes")
	flags.String("regions", "", "comma-separated target regions")
	flags.Bool("all-regions", false, "deploy to all regions")
	flags.StringArray("answer", nil, "additional answers in key=val form (repeatable)")
	flags.String("answer-file", "", "path to YAML/JSON answer file")
	flags.Bool("non-interactive", false, "force non-interactive mode")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
