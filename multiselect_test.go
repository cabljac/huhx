package huhx

import (
	"errors"
	"reflect"
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

// TestRunner_MultiSelectValidatorOnInjected confirms that the validator
// runs on the resolved []T after parsing comma-separated answers and
// that a returned error is wrapped with the field key.
func TestRunner_MultiSelectValidatorOnInjected(t *testing.T) {
	var tags []string
	want := errors.New("too-many-tags")

	form := NewForm(NewGroup(
		NewMultiSelect[string]().Key("tags").
			Options(
				huh.NewOption("a", "a"),
				huh.NewOption("b", "b"),
				huh.NewOption("c", "c"),
			).
			Value(&tags).
			Validate(func(v []string) error {
				if len(v) > 2 {
					return want
				}
				return nil
			}),
	))

	r := New(form,
		WithNonInteractive(Always),
		WithAnswers(map[string]any{"tags": "a,b,c"}),
	)

	err := r.Run()
	if err == nil {
		t.Fatal("expected validator error")
	}
	if !strings.Contains(err.Error(), `field "tags"`) {
		t.Errorf("expected field-prefixed error, got %q", err.Error())
	}
	if !errors.Is(err, want) {
		t.Errorf("expected validator sentinel wrapped, got %v", err)
	}
}

// TestRunner_MultiSelectCommaEdgeCases pins documented parser behavior:
// empty parts (consecutive commas or leading/trailing comma) are skipped
// rather than causing an "is not a valid option" error.
func TestRunner_MultiSelectCommaEdgeCases(t *testing.T) {
	cases := map[string]struct {
		answer string
		want   []string
	}{
		"consecutive commas":     {"a,,b", []string{"a", "b"}},
		"leading comma":          {",a,b", []string{"a", "b"}},
		"trailing comma":         {"a,b,", []string{"a", "b"}},
		"surrounding whitespace": {" a ,  b ", []string{"a", "b"}},
		"all empty":              {",,,", []string{}},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var got []string
			form := NewForm(NewGroup(
				NewMultiSelect[string]().Key("xs").
					Options(
						huh.NewOption("a", "a"),
						huh.NewOption("b", "b"),
					).
					Value(&got).
					Optional(),
			))

			r := New(form,
				WithNonInteractive(Always),
				WithAnswers(map[string]any{"xs": tc.answer}),
			)

			if err := r.Run(); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestMultiSelect_Forwarders(t *testing.T) {
	m := NewMultiSelect[string]().
		Key("k").
		Title("t").
		TitleFunc(func() string { return "tf" }, nil).
		Description("d").
		DescriptionFunc(func() string { return "df" }, nil).
		Options(huh.NewOption("a", "a"), huh.NewOption("b", "b")).
		Limit(2).
		Filterable(false).
		Filtering(false).
		Width(40).
		Height(10)
	if m == nil {
		t.Fatal("expected non-nil multiselect after forwarder chain")
	}
}

func TestMultiSelect_AccessorWritesValue(t *testing.T) {
	var dst []string
	acc := huh.NewPointerAccessor(&dst)
	form := NewForm(NewGroup(
		NewMultiSelect[string]().
			Key("tags").
			Options(huh.NewOption("a", "a"), huh.NewOption("b", "b"), huh.NewOption("c", "c")).
			Accessor(acc),
	))
	r := New(form, WithNonInteractive(Always), WithAnswers(map[string]any{"tags": "a,c"}))
	if err := r.Run(); err != nil {
		t.Fatal(err)
	}
	if len(dst) != 2 || dst[0] != "a" || dst[1] != "c" {
		t.Errorf("expected accessor to receive [a c], got %v", dst)
	}
}
