package tcp

// Options options
type Options struct {
	Addr         string
	Max          int
	Endl         byte
	handlePacket func([]byte) ([]byte, error)
}

// Option option func
type Option func(*Options)

func newOptions(opts ...Option) Options {
	opt := Options{
		Addr: "127.0.0.1:9090",
		Max:  1,
		Endl: '\n',
		handlePacket: func(buf []byte) ([]byte, error) {
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

// Max concurrent
func Max(max int) Option {
	return func(o *Options) {
		o.Max = max
	}
}

// HandlePacket using default Handler, But if Handler set ,HandlePacket useless
func HandlePacket(f func([]byte) ([]byte, error)) Option {
	return func(o *Options) {
		o.handlePacket = f
	}
}

// Endl using default Handler
func Endl(delim byte) Option {
	return func(o *Options) {
		o.Endl = delim
	}
}
