package server

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"strings"

	"golang.org/x/crypto/nacl/sign"
)

const DefaultPort = 2850

type secConn struct {
	conn  net.Conn
	pkeys map[[32]byte][32]byte
	magic []byte
}

func (sc *secConn) Read(p []byte) (n int, err error) {

	var (
		hpk, pk [32]byte
		in      = make([]byte, len(p)+len(hpk)+len(sc.magic)+sign.Overhead)
	)

	n, err = sc.conn.Read(in)
	if err != nil {
		return
	}

	if n <= len(sc.magic)+len(hpk)+sign.Overhead {
		log.Printf("Message is too short: %d", n)
		return 0, io.ErrUnexpectedEOF
	}

	magic := make([]byte, len(sc.magic))
	copy(magic, in[0:len(magic)])

	// check first 6 bytes of magic - signature and major version number
	if !bytes.Equal(magic[0:6], sc.magic[0:6]) {
		log.Printf("Bad signature or incompatible versions: server [%x], client [%x]", sc.magic, magic)
		return 0, rpc.ErrShutdown
	}

	copy(hpk[:], in[len(magic):len(magic)+len(hpk)])

	var ok bool
	if pk, ok = sc.pkeys[hpk]; !ok {
		log.Printf("Call with unauthorized key: %s", hex.EncodeToString(hpk[:]))
		return 0, rpc.ErrShutdown
	}

	out, ok := sign.Open([]byte{}, in[len(magic)+len(hpk):n], &pk)
	if !ok {
		log.Printf("Call fails verification with key: %s", hex.EncodeToString(pk[:]))
		return 0, rpc.ErrShutdown
	}
	copy(p, out)
	return len(out), nil
}

func (sc *secConn) Write(p []byte) (n int, err error) {
	return sc.conn.Write(p)
}

func (sc *secConn) Close() error {
	return sc.conn.Close()
}

// Serve handles backend rpc calls.
func Serve(ctx context.Context, port int, le string, pkeys map[[32]byte][32]byte, magic []byte) error {

	if err := rpc.Register(NewURI()); err != nil {
		return fmt.Errorf("unable to register URI rpc object: %w", err)
	}
	if err := rpc.Register(NewClipboard(le)); err != nil {
		return fmt.Errorf("unable to register Clipboard rpc object: %w", err)
	}

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return fmt.Errorf("unable to resolve address: %w", err)
	}

	log.Printf("gclpr server listens on '%s'\n", addr)

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return fmt.Errorf("unable to listen on '%s': %w", addr, err)
	}

	// This will break the loop
	go func() {
		<-ctx.Done()
		l.Close()
	}()

	log.Print("gclpr server is ready\n")
	for {
		conn, err := l.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				return fmt.Errorf("gclpr server is unable to accept requests: %w", err)
			}
			log.Print("gclpr server is shutting down\n")
			return nil
		}
		go func(sc *secConn) {
			defer sc.Close()
			log.Printf("gclpr server accepted request from '%s'", sc.conn.RemoteAddr())
			rpc.ServeConn(sc)
			log.Printf("gclpr server handled request from '%s'", sc.conn.RemoteAddr())
		}(&secConn{
			conn:  conn,
			pkeys: pkeys,
			magic: magic,
		})
	}
}
