package icmp

import (
	"math"
)

// Options options
type Options struct {
	count   int
	size    int
	timeout int
}

// Option Options function
type Option func(*Options)

func newOptions(opts ...Option) Options {
	opt := Options{
		size:    56,
		count:   4,
		timeout: 5,
	}
	for _, o := range opts {
		o(&opt)
	}
	return opt
}

// Size set size
func Size(size int) Option {
	return func(o *Options) {
		o.size = size
	}
}

// Timeout set timeout
func Timeout(timeout int) Option {
	return func(o *Options) {
		o.timeout = timeout
	}
}

// Count count
func Count(count int) Option {
	return func(o *Options) {
		if count == 0 {
			count = math.MaxInt64
		}
		o.count = count
	}
}
