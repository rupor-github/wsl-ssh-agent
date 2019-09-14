package citrus

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
)

// Citrus holds "lemonade" server state.
type Citrus struct {
	connCh   chan net.Conn
	params   ParamsValue
	ra       *Range
	listener *net.TCPListener
	debug    bool
}

// NewCitrus initializes "lemonade" server.
func NewCitrus(params ParamsValue, debug bool) (*Citrus, error) {

	res := &Citrus{
		connCh: make(chan net.Conn, 1),
		params: params,
		debug:  debug,
	}

	if err := rpc.Register(NewURI(res.connCh, res.params, res.debug)); err != nil {
		return nil, fmt.Errorf("unable to register URI rpc: %w", err)
	}
	if err := rpc.Register(NewClipboard(res.connCh, res.params, res.debug)); err != nil {
		return nil, fmt.Errorf("unable to register Clipboard rpc: %w", err)
	}

	ra, err := NewRange(params.Allow)
	if err != nil {
		return nil, fmt.Errorf("unable to process allowed IP ranges: %w", err)
	}
	res.ra = ra

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", params.Port))
	if err != nil {
		return nil, fmt.Errorf("ResolveTCPAddr error: '%w'", err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("ListenTCP error: '%w'", err)
	}
	res.listener = l

	return res, nil
}

// Serve starts "lemonade" server backend.
func (c *Citrus) Serve(debug bool) {
	defer c.listener.Close()
	for {
		conn, err := c.listener.Accept()
		if err != nil {
			log.Printf("lemonade Accept error '%s'", err)
			return
		}
		go func(conn net.Conn, debug bool) {
			defer conn.Close()

			if debug {
				log.Printf("lemonade server request from '%s'", conn.RemoteAddr())
			}
			if c.ra.IsConnIn(conn) {
				c.connCh <- conn
				rpc.ServeConn(conn)
				if debug {
					log.Printf("lemonade server done with '%s'", conn.RemoteAddr())
				}
			}
		}(conn, debug)
	}
}
