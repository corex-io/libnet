package tcp

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/corex-io/log"
)

const endl = '\n'

// Server tcp server
type Server struct {
	opts   Options
	listen *net.TCPListener
}

// New new tcp server
func New(opts ...Option) *Server {
	options := newOptions(opts...)
	return &Server{
		opts: options,
	}
}

// Init init
func (s *Server) Init(opts ...Option) error {
	for _, o := range opts {
		o(&s.opts)
	}

	tcpaddr, err := net.ResolveTCPAddr("tcp", s.opts.Addr)
	if err != nil {
		return err
	}
	s.listen, err = net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		return err
	}

	log.Infof("listen %v", tcpaddr.String())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		conn, err := s.listen.AcceptTCP()
		if err != nil {
			return err
		}
		log.Infof("accept %s...", conn.RemoteAddr().String())
		go func() {
			if err := s.process(ctx, conn); err != nil {
				log.Errorf("process: %v", err)
			}
		}()
	}
}

func (s *Server) process(ctx context.Context, conn *net.TCPConn) error {
	stime := time.Now()
	log.Infof("recv: %v, %p", stime.Nanosecond(), conn)
	defer func() {
		conn.Close()
		log.Infof("close Recv %v, cost=%v", stime.Nanosecond(), time.Since(stime))
	}()
	buf := bufio.NewReader(conn)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			bytes, err := buf.ReadBytes(endl)
			if err != nil {
				return err
			}
			resp, err := s.opts.handleFunc(bytes)
			if err != nil {
				return fmt.Errorf("handle: %w", err)
			}
			if len(resp) == 0 {
				continue
			}
			if _, err := conn.Write(resp); err != nil {
				return fmt.Errorf("resp: %w", err)
			}
		}
	}
}
