package proteus

import (
	"io"
	"strings"

	"github.com/simplesurance/proteus/plog"
	"github.com/simplesurance/proteus/sources"
)

// Option specifes options when creating a configuration parser.
type Option func(*settings)

type settings struct {
	providers       []sources.Provider
	loggerFn        plog.Logger
	onelineDesc     string
	valueFormatting ValueFormattingOptions

	// auto-usage (aka --help)
	autoUsageExitFn func()
	autoUsageWriter io.Writer

	// version (aka --version)
	version string
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

// WithPrintfLogger use the printf-style logFn function as logger.
func WithPrintfLogger(logFn func(format string, v ...any)) Option {
	return func(p *settings) {
		p.loggerFn = func(e plog.Entry) {
			logFn("%-5s %s:%d %s\n",
				e.Severity, e.Caller.File, e.Caller.LineNumber, e.Message)
		}
	}
}

// WithValueFormatting specifies options for pre-processing values before
// using them. See ValueFormattingOptions for more details.
func WithValueFormatting(o ValueFormattingOptions) Option {
	return func(p *settings) {
		p.valueFormatting = o
	}
}

func WithVersion(version string) Option {
	return func(s *settings) {
		s.version = version
	}
}

// ValueFormattingOptions specifies how values of parameters are "trimmed".
type ValueFormattingOptions struct {
	// TrimSpace instructs proteus to trim leading and trailing spaces from
	// values of parameters.
	TrimSpace bool
}

func (t ValueFormattingOptions) apply(v string) string {
	if t.TrimSpace {
		v = strings.TrimSpace(v)
	}

	return v
}
