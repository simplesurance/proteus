package proteus

import (
	"io"

	"github.com/simplesurance/proteus/sources"
)

// Option allows specifying options when creating a configuration parser.
type Option func(*settings)

type settings struct {
	sources           []sources.Source
	autoUsageExitFn   func()
	autoUsageHeadline string
	autoUsageWriter   io.Writer
	loggerFn          Logger
}

func (s *settings) apply(options ...Option) {
	for _, o := range options {
		o(s)
	}
}

// WithSources adds configuration sources to be used to read values for the
// application parameters. It can be called multiple times. If a parameter is
// found in more than one source, the source that was added first takes
// precedence.
func WithSources(s ...sources.Source) Option {
	return func(p *settings) {
		p.sources = append(p.sources, s...)
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
