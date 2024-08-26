package icmp

import (
	"context"
	"os"
	"testing"
)

func TestPingPing(t *testing.T) {
	ping := New(Count(4))
	stat, err := ping.Ping(context.Background(), "qq.com")
	if err != nil {
		t.Fatal(err)
	}
	stat.Print(os.Stdout)
	t.Logf("%#v, %#v", stat, err)
}
