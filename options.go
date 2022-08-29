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

	// dry-mode
	autoDryModeFn func(*Parsed)
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

// WithDryMode add a special parameter "--dry-mode", that can only be provided
// by command-line flags, and can be used to validate parameters without
// executing the application. The provided callback function will be invoked
// and will have access to the Parsed object, than can then be used to
// print errors, dump variable values or exit with some status code.
//
// If the callback function does not exit, proteus will terminate the
// application with 0 or 1, depending on the validation being successful or
// not.
//
//	parsed, err := proteus.MustParse(&params,
//		proteus.WithDryMode(func(parsed *proteus.Parsed) {
//			if err := parsed.Valid(); err != nil {
//				parsed.ErrUsage(os.Stdout, err)
//				os.Exit(42)
//			}
//
//			fmt.Println("The following parameters are found to be valid:")
//			parsed.Dump(os.Stdout)
//			os.Exit(0)
//		}))
func WithDryMode(f func(*Parsed)) Option {
	return func(p *settings) {
		p.autoDryModeFn = f
	}
}
