package icmp

import (
	"io"
	"math"
	"time"
)

// Options options
type Options struct {
	count   int
	size    int
	Log     io.Writer
	timeout time.Duration
}

// Option Options function
type Option func(*Options)

func newOptions(opts []Option, extends ...Option) Options {
	opt := Options{
		size:    56,
		count:   4,
		Log:     io.Discard,
		timeout: 3 * time.Second,
	}
	for _, o := range opts {
		o(&opt)
	}
	for _, o := range extends {
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
func Timeout(timeout time.Duration) Option {
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

// Log ..
func Log(w io.Writer) Option {
	return func(o *Options) {
		o.Log = w
	}
}
