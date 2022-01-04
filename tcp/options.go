package tcp

// Options options
type Options struct {
	Addr       string
	Max        int
	handleFunc func([]byte) ([]byte, error)
}

// Option option func
type Option func(*Options)

func newOptions(opts ...Option) Options {
	opt := Options{
		Addr: "127.0.0.1:9090",
		Max: 1,
		handleFunc: func(buf []byte) ([]byte, error) {
			return buf, nil
		},
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

// Addr addr
func Addr(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}

// HandleFunc handlefunc
func HandleFunc(f func([]byte) ([]byte, error)) Option {
	return func(o *Options) {
		o.handleFunc = f
	}
}

// Max concurrent connect
func Max(max int) Option {
	return func(o *Options) {
		o.Max = max
	}
}
