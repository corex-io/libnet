package tcp

import (
	"bufio"
	"fmt"
	"net"

	"github.com/corex-io/log"
)

type defaultHandler struct {
	delim        byte
	HandlePacket func([]byte) ([]byte, error)
}

func (t *defaultHandler) ServeTCP(conn *net.TCPConn) {
	if err := t.serve(conn); err != nil {
		log.Errorf("%v", err)
	}
}

func (t *defaultHandler) serve(conn *net.TCPConn) error {
	buf := bufio.NewReader(conn)
	for {
		recv, err := buf.ReadBytes(t.delim)
		if err != nil {
			return err
		}
		res, err := t.HandlePacket(recv)
		if err != nil {
			return fmt.Errorf("handle: %w", err)
		}
		if len(res) == 0 {
			continue
		}
		if _, err := conn.Write(res); err != nil {
			return fmt.Errorf("resp: %w", err)
		}
	}
}
