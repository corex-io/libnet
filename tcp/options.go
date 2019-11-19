package tcp

// Options options
type Options struct {
    Addr       string
    handleFunc func([]byte) ([]byte, error)
}

// Option option func
type Option func(*Options)

func newOptions(opts ...Option) Options {
    opt := Options{
        Addr: "127.0.0.1:9090",
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
    return func(opts *Options) {
        opts.Addr = addr
    }
}

// HandleFunc handlefunc
func HandleFunc(f func([]byte) ([]byte, error)) Option {
    return func(opts *Options) {
        opts.handleFunc = f
    }
}
