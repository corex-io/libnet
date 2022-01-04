package icmp

import (
	"context"
	"testing"
)

func TestPing(t *testing.T) {
	ping := New()
	stats, err := ping.Send(context.Background(), "127.0.0.1")
	t.Logf("%#v, %#v", stats, err)
}
