package proteus

import (
	"io"

	"github.com/simplesurance/proteus/sources"
)

// Option allows specifying options when creating a configuration parser.
type Option func(*settings)

type settings struct {
	providers   []sources.Provider
	loggerFn    Logger
	onelineDesc string

	// auto-usage (aka --help)
	autoUsageExitFn func()
	autoUsageWriter io.Writer
}

func (s *settings) apply(options ...Option) {
	for _, o := range options {
		o(s)
	}
}

// WithProviders specifies from where the configuration should be read.
func WithProviders(s ...sources.Provider) Option {
	return func(p *settings) {
		p.providers = s
	}
}

// WithShortDescription allows specifying a short one-line description for
// the application. Is used when generating help information.
func WithShortDescription(oneline string) Option {
	return func(p *settings) {
		p.onelineDesc = oneline
	}
}

// WithAutoUsage will change how the --help parameter is parsed, allowing to
// specify a writer, for the usage information and an "exit function".
// If not specified, proteus will use stdout for writer and will use
// os.Exit(0) as exit function.
func WithAutoUsage(writer io.Writer, exitFn func()) Option {
	return func(p *settings) {
		p.autoUsageExitFn = exitFn
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
