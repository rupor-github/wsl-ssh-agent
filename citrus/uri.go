package citrus

import (
	"log"
	"net"
	"net/url"
	"regexp"

	"github.com/skratchdot/open-golang/open"
)

// URI is used by "lemonade" to rpc open command.
type URI struct {
	connCh chan net.Conn
	params ParamsValue
	debug  bool
}

// NewURI initializes URI structure.
func NewURI(ch chan net.Conn, params ParamsValue, debug bool) *URI {
	return &URI{
		connCh: ch,
		params: params,
		debug:  debug,
	}
}

// OpenParam defines rpc parameter for "lemonade" open command".
type OpenParam struct {
	URI           string
	TransLoopback bool
}

// Open is implementation of "lemonade" rpc "open" command.
func (u *URI) Open(param *OpenParam, _ *struct{}) error {
	conn := <-u.connCh
	if u.debug {
		log.Printf("lemonade URI parameters received: '%v'", *param)
	}
	uri := param.URI
	if param.TransLoopback {
		uri = translateLoopbackIP(param.URI, conn)
	}
	if u.debug {
		log.Printf("lemonade run URI: '%s'", uri)
	}
	return open.Run(uri)
}

func removeIPv6Brackets(ip string) string {
	if regexp.MustCompile(`^\[.+\]$`).MatchString(ip) {
		return ip[1 : len(ip)-1]
	}
	return ip
}

func splitHostPort(hostPort string) []string {
	portRe := regexp.MustCompile(`:(\d+)$`)
	portSlice := portRe.FindStringSubmatch(hostPort)
	if len(portSlice) == 0 {
		return []string{removeIPv6Brackets(hostPort)}
	}
	port := portSlice[1]
	host := hostPort[:len(hostPort)-len(port)-1]
	return []string{removeIPv6Brackets(host), port}
}

func translateLoopbackIP(uri string, conn net.Conn) string {

	parsed, err := url.Parse(uri)
	if err != nil {
		log.Printf("Translating loopback, url parse error: %s", err.Error())
		return uri
	}

	const (
		host = iota
		port
	)
	parts := splitHostPort(parsed.Host)

	ip := net.ParseIP(parts[host])
	if ip == nil || !ip.IsLoopback() {
		return uri
	}

	addr := conn.RemoteAddr().(*net.TCPAddr).IP.String()
	if len(parts) == 1 {
		parsed.Host = addr
	} else {
		parsed.Host = net.JoinHostPort(addr, parts[port])
	}
	return parsed.String()
}
