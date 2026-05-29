package huhx

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// answerFileCache memoizes parsed answer files for the lifetime of the
// process. A CLI that drives several forms in one run constructs a fresh
// Runner per form, so without this cache the same file is re-read and
// re-parsed once per form. Entries are keyed on path and invalidated when
// the file's mod time or size changes.
var (
	answerFileMu    sync.Mutex
	answerFileCache = map[string]answerFileEntry{}
)

type answerFileEntry struct {
	modTime time.Time
	size    int64
	answers map[string]string
}

// loadAnswerFile reads a YAML/JSON file into a flat map of string answers.
// Returns an empty (non-nil) map if no path is configured so callers can
// always index safely. Results are memoized per path (see answerFileCache);
// each call returns an independent copy so callers may mutate it freely.
func loadAnswerFile(path string) (map[string]string, error) {
	if path == "" {
		return map[string]string{}, nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("read answer file %q: %w", path, err)
	}

	answerFileMu.Lock()
	defer answerFileMu.Unlock()

	if e, ok := answerFileCache[path]; ok && e.modTime.Equal(info.ModTime()) && e.size == info.Size() {
		return cloneAnswers(e.answers), nil
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

	answerFileCache[path] = answerFileEntry{
		modTime: info.ModTime(),
		size:    info.Size(),
		answers: out,
	}
	return cloneAnswers(out), nil
}

// cloneAnswers returns a shallow copy so cached maps are never shared with
// (and mutated by) callers.
func cloneAnswers(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// cobraAnswerPairs reads --answer key=val pairs from the cobra command's
// StringArray flag named "answer". StringArray is required rather than
// StringSlice because StringSlice splits each value on commas, which
// would corrupt MultiSelect answers like
// --answer regions=us-east-1,us-west-2. Returns an empty (non-nil) map
// when no command or no flag is configured so callers can always index
// safely.
func cobraAnswerPairs(cmd *cobra.Command) (map[string]string, error) {
	if cmd == nil {
		return map[string]string{}, nil
	}
	f := cmd.Flags().Lookup("answer")
	if f == nil {
		return map[string]string{}, nil
	}
	raw, err := cmd.Flags().GetStringArray("answer")
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
