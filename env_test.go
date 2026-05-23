package huhx

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestEnvTruthy(t *testing.T) {
	cases := []struct {
		val  string
		want bool
	}{
		{"1", true},
		{"true", true},
		{"TRUE", true},
		{"True", true},
		{"yes", true},
		{"YES", true},
		{"0", false},
		{"false", false},
		{"no", false},
		{"", false},
		{"maybe", false},
	}
	for _, tc := range cases {
		t.Run(tc.val, func(t *testing.T) {
			t.Setenv("TEST_KEY", tc.val)
			got := envTruthy("TEST_KEY")
			if got != tc.want {
				t.Errorf("envTruthy(%q)=%v, want %v", tc.val, got, tc.want)
			}
		})
	}
}

func TestIsNonInteractive(t *testing.T) {
	t.Run("Always forces true", func(t *testing.T) {
		t.Setenv("NON_INTERACTIVE", "")
		t.Setenv("CI", "")
		r := New(NewForm(), WithNonInteractive(Always))
		if !r.isNonInteractive() {
			t.Error("expected true with Mode=Always")
		}
	})

	t.Run("Never overrides CI", func(t *testing.T) {
		t.Setenv("NON_INTERACTIVE", "")
		t.Setenv("CI", "1")
		r := New(NewForm(), WithNonInteractive(Never))
		if r.isNonInteractive() {
			t.Error("expected false with Mode=Never despite CI=1")
		}
	})

	t.Run("AutoDetect with NON_INTERACTIVE=1", func(t *testing.T) {
		t.Setenv("NON_INTERACTIVE", "1")
		t.Setenv("CI", "")
		r := New(NewForm())
		if !r.isNonInteractive() {
			t.Error("expected true with NON_INTERACTIVE=1")
		}
	})

	t.Run("AutoDetect with CI=1", func(t *testing.T) {
		t.Setenv("NON_INTERACTIVE", "")
		t.Setenv("CI", "1")
		r := New(NewForm())
		if !r.isNonInteractive() {
			t.Error("expected true with CI=1")
		}
	})

	t.Run("AutoDetect with cobra --non-interactive flag", func(t *testing.T) {
		t.Setenv("NON_INTERACTIVE", "")
		t.Setenv("CI", "")
		cmd := &cobra.Command{Use: "t"}
		cmd.Flags().Bool("non-interactive", false, "")
		if err := cmd.ParseFlags([]string{"--non-interactive"}); err != nil {
			t.Fatal(err)
		}
		r := New(NewForm(), WithCobraFlags(cmd))
		if !r.isNonInteractive() {
			t.Error("expected true with --non-interactive flag set")
		}
	})
}
