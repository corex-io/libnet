package tcp

import (
	"net"

	"github.com/corex-io/log"
)

// Handler tcp handler
type Handler interface {
	ServeTCP(*net.TCPConn)
}

// HandlerFunc handler func
type HandlerFunc func(*net.TCPConn)

// ServeTCP implement Handler interface
func (f HandlerFunc) ServeTCP(conn *net.TCPConn) {
	f(conn)
}

// Server tcp server
type Server struct {
	opts          Options
	limit         chan struct{}
	listen        *net.TCPListener
	Handler       Handler
}

// New new tcp server
func New(opts ...Option) *Server {
	options := newOptions(opts...)
	server := &Server{
		opts:  options,
		limit: make(chan struct{}, options.Max),
		Handler: &defaultHandler{
			delim:        options.Endl,
			HandlePacket: options.handlePacket,
		},
	}
	return server
}

// Handle set handle
func (s *Server) Handle(handler Handler) {
	s.Handler = handler
}

// HandleFunc set Handle func
func (s *Server) HandleFunc(handlerFunc HandlerFunc) {
	s.Handler = handlerFunc
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
	defer s.listen.Close()
	log.Infof("listen %s://%s", tcpaddr.Network(), tcpaddr.String())

	for {
		conn, err := s.listen.AcceptTCP()
		if err != nil {
			return err
		}
		log.Debugf("accept %s...", conn.RemoteAddr().String())

		select {
		case s.limit <- struct{}{}:
			go func() {
				defer func() {
					<-s.limit
					conn.Close()
				}()
				s.Handler.ServeTCP(conn)
			}()
		default:
			if _, err = conn.Write([]byte("Too many connections.")); err != nil {
				log.Errorf("write %w", err)
			}
			conn.Close()
		}
	}
}
// Close close
func (s *Server) Close() error {
	return s.listen.Close()
}
