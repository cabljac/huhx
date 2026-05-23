package main

import (
	"fmt"
	"os"

	"charm.land/huh/v2"
	"github.com/cabljac/huhx"
	"github.com/spf13/cobra"
)

func main() {
	var (
		name        string
		environment string
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

			fmt.Printf("Deploying %q to %s (all regions: %v)\n", name, environment, allRegions)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.String("name", "", "app name")
	flags.String("environment", "", "target environment (staging|prod)")
	flags.Bool("all-regions", false, "deploy to all regions")
	flags.StringSlice("answer", nil, "additional answers in key=val form (repeatable)")
	flags.String("answer-file", "", "path to YAML/JSON answer file")
	flags.Bool("non-interactive", false, "force non-interactive mode")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
