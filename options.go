package proteus

import (
	"io"

	"github.com/simplesurance/proteus/sources"
)

// Option allows specifying options when creating a configuration parser.
type Option func(*settings)

type settings struct {
	providers []sources.Provider
	loggerFn  Logger

	// auto-usage (aka --help)
	autoUsageExitFn   func()
	autoUsageHeadline string
	autoUsageWriter   io.Writer
}

func (s *settings) apply(options ...Option) {
	for _, o := range options {
		o(s)
	}
}

// WithProviders specifies from where the configuration should be read.
func WithProviders(s ...sources.Provider) Option {
	return func(p *settings) {
		p.providers = append(p.providers, s...)
	}
}

// WithAutoUsage will parse the `--help` command-line flag. If it is present,
// will show usage and exit using the provided exitFn. If the exitFn is nil
// it exists with os.Exit(0).
func WithAutoUsage(writer io.Writer, headline string, exitFn func()) Option {
	return func(p *settings) {
		p.autoUsageExitFn = exitFn
		p.autoUsageHeadline = headline
		p.autoUsageWriter = writer
	}
}

// WithLogger allows providing a custom logger. By default logs are suppressed.
func WithLogger(l Logger) Option {
	return func(p *settings) {
		p.loggerFn = l
	}
}

// Logger is the function used to output human-readable diagnostics
// information. Depth can optionally be used to determine the real caller of
// the log function, by skipping the correct number of intermediate frames
// in the stacktrace.
type Logger func(msg string, depth int)
