package tcp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"

	"github.com/corex-io/limit"

	"github.com/corex-io/log"
)

const endl = '\n'

// Server tcp server
type Server struct {
	opts   Options
	limit  *limit.Limit
	listen *net.TCPListener
}

// New new tcp server
func New(opts ...Option) *Server {
	options := newOptions(opts...)
	return &Server{
		opts:  options,
		limit: limit.New(limit.Max(options.Max)),
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

	log.Debugf("listen %v", tcpaddr.String())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		s.limit.Add(1)
		conn, err := s.listen.AcceptTCP()
		if err != nil {
			return err
		}
		log.Debugf("accept %s...", conn.RemoteAddr().String())
		go func() {
			defer func() {
				s.limit.Done()
				conn.Close()
			}()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					if err := s.process(ctx, conn); err != nil {
						log.Errorf("process: %v", err)
						if err == io.EOF {
							return
						}
					}
				}
			}

		}()
	}
}

func (s *Server) process(ctx context.Context, conn *net.TCPConn) error {
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
				log.Errorf("handle: resp=%s, err=%v, bytes=%s", string(resp), err, string(bytes))
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
