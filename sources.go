package huhx

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// loadAnswerFile reads a YAML/JSON file into a flat map of string answers.
// Returns an empty (non-nil) map if no path is configured so callers can
// always index safely.
func loadAnswerFile(path string) (map[string]string, error) {
	if path == "" {
		return map[string]string{}, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read answer file %q: %w", path, err)
	}
	var generic map[string]any
	if err := yaml.Unmarshal(raw, &generic); err != nil {
		return nil, fmt.Errorf("parse answer file %q: %w", path, err)
	}
	out := make(map[string]string, len(generic))
	for k, v := range generic {
		out[k] = fmt.Sprintf("%v", v)
	}
	return out, nil
}

// cobraAnswerPairs reads --answer key=val pairs from the cobra command's
// StringSlice flag named "answer". Returns an empty (non-nil) map when
// no command or no flag is configured so callers can always index safely.
func cobraAnswerPairs(cmd *cobra.Command) (map[string]string, error) {
	if cmd == nil {
		return map[string]string{}, nil
	}
	f := cmd.Flags().Lookup("answer")
	if f == nil {
		return map[string]string{}, nil
	}
	raw, err := cmd.Flags().GetStringSlice("answer")
	if err != nil {
		return nil, fmt.Errorf("read --answer: %w", err)
	}
	out := make(map[string]string, len(raw))
	for _, pair := range raw {
		k, v, found := strings.Cut(pair, "=")
		if !found {
			return nil, fmt.Errorf("invalid --answer %q: expected key=val", pair)
		}
		out[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return out, nil
}

// envKey turns a field key + prefix into an environment variable name.
// "name" with prefix "MYCLI" becomes "MYCLI_NAME". "all-regions" becomes
// "MYCLI_ALL_REGIONS".
func envKey(prefix, key string) string {
	if prefix == "" {
		return ""
	}
	k := strings.ReplaceAll(key, "-", "_")
	return strings.ToUpper(prefix + "_" + k)
}
