package tcp_test

import (
	"testing"

	"github.com/corex-io/libnet/tcp"
)

func Test_tcpd(t *testing.T) {
	tcpd := tcp.New(
		tcp.Max(1),
		tcp.HandlePacket(func(recv []byte) ([]byte, error) {
			return recv, nil
		}))
	// tcpd.HandleFunc(func(conn *net.TCPConn) {
	//     conn.Write([]byte("12345"))
	//     conn.Close()
	// })

	tcpd.Init()
}
