package huhx

// Mode controls how the runner chooses between interactive and
// non-interactive execution.
type Mode int

const (
	// AutoDetect chooses non-interactive when env (NON_INTERACTIVE/CI),
	// stdin (not a TTY), or the --non-interactive flag say so; otherwise
	// interactive.
	AutoDetect Mode = iota
	// Always forces non-interactive mode.
	Always
	// Never forces interactive mode.
	Never
)
