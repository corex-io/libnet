package tcp_test

import (
	"testing"

	"github.com/corex-io/libnet/tcp"
)

func Test_Tcpd(t *testing.T) {
	tcpd := tcp.New()
	tcpd.Init()
}
