package huhx

import (
	"strings"
	"testing"

	"charm.land/huh/v2"
)

func TestRunner_MultiSelectOptionsFunc(t *testing.T) {
	servicesByRegion := map[string][]huh.Option[string]{
		"us": {
			huh.NewOption("ec2", "ec2"),
			huh.NewOption("s3", "s3"),
			huh.NewOption("lambda", "lambda"),
		},
		"eu": {
			huh.NewOption("ec2", "ec2"),
			huh.NewOption("s3", "s3"),
			huh.NewOption("rds", "rds"),
		},
	}

	build := func(region *string, services *[]string) *Form {
		return NewForm(
			NewGroup(
				NewSelect[string]().Key("region").
					Options(
						huh.NewOption("us", "us"),
						huh.NewOption("eu", "eu"),
					).Value(region),
			),
			NewGroup(
				NewMultiSelect[string]().Key("services").
					OptionsFunc(func() []huh.Option[string] {
						return servicesByRegion[*region]
					}, region).
					Value(services),
			),
		)
	}

	t.Run("happy path", func(t *testing.T) {
		var region string
		var services []string
		form := build(&region, &services)

		r := New(form,
			WithNonInteractive(Always),
			WithAnswers(map[string]any{
				"region":   "us",
				"services": "ec2,lambda",
			}),
		)
		if err := r.Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if region != "us" {
			t.Errorf("expected region=us, got %q", region)
		}
		if len(services) != 2 || services[0] != "ec2" || services[1] != "lambda" {
			t.Errorf("expected [ec2 lambda], got %v", services)
		}
	})

	t.Run("rejects unknown in dynamic list", func(t *testing.T) {
		var region string
		var services []string
		form := build(&region, &services)

		r := New(form,
			WithNonInteractive(Always),
			WithAnswers(map[string]any{
				"region":   "us",
				"services": "ec2,rds",
			}),
		)
		err := r.Run()
		if err == nil {
			t.Fatal("expected error for option not present in dynamic list")
		}
		if !strings.Contains(err.Error(), `field "services"`) {
			t.Errorf("expected field-prefixed error, got %q", err.Error())
		}
		if !strings.Contains(err.Error(), `"rds" is not a valid option`) {
			t.Errorf("expected invalid-option message, got %q", err.Error())
		}
	})
}
