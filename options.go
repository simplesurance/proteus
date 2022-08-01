package proteus

import (
	"io"

	"github.com/simplesurance/proteus/plog"
	"github.com/simplesurance/proteus/sources"
)

// Option specifes options when creating a configuration parser.
type Option func(*settings)

type settings struct {
	providers   []sources.Provider
	loggerFn    plog.Logger
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
// If not specified, proteus will use the equivalent to:
//
//	WithEnv(cfgflags.New(), cfgenv.New("CFG"))
//
// Providing this option override any previous configuration for providers.
func WithProviders(s ...sources.Provider) Option {
	return func(p *settings) {
		p.providers = s
	}
}

// WithShortDescription species a short one-line description for
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

// WithLogger provides a custom logger. By default logs are suppressed.
//
// Warning: the "Logger" interface is expected to change in the stable release.
func WithLogger(l plog.Logger) Option {
	return func(p *settings) {
		p.loggerFn = l
	}
}
