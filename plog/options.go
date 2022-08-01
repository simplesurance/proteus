package plog

type Option func(*options)

func applyOptions(o ...Option) options {
	ret := options{
		skipCallers: 2,
	}

	for _, o := range o {
		o(&ret)
	}

	return ret
}

type options struct {
	skipCallers int
}

// SkipCallers must be used when the caller to be registered on the log
// message is not the immediate caller of the log function. The provided
// number specify how many callers must be skipped.
func SkipCallers(i int) Option {
	return func(o *options) {
		o.skipCallers += i
	}
}
