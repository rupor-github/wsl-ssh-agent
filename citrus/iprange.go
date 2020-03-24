package citrus

// This comes directly from https://github.com/pocke/go-iprange
// I simply could not stand using "InlucdeConn"..., literately

import (
	"net"
	"strings"
)

// Range defines slice of checkable addresses.
type Range struct {
	allows []*net.IPNet
}

// NewRange parses comma delimited list of addresses and allocates new range.
func NewRange(ipStr string) (*Range, error) {

	IPs := strings.Split(ipStr, ",")
	r := &Range{
		allows: make([]*net.IPNet, 0, len(IPs)),
	}

	for _, i := range IPs {
		if !strings.Contains(i, "/") {
			if strings.Contains(i, ".") { // IPv4
				i += "/32"
			} else { // IPv6
				i += "/128"
			}
		}

		_, mask, err := net.ParseCIDR(i)
		if err != nil {
			return nil, err
		}
		r.allows = append(r.allows, mask)
	}
	return r, nil
}

// IsStrIn checks is address (as a string) is included.
func (r *Range) IsStrIn(addr string) bool {
	return r.IsIn(net.ParseIP(addr))
}

// IsIn checks is address is included.
func (r *Range) IsIn(addr net.IP) bool {
	for _, m := range r.allows {
		if m.Contains(addr) {
			return true
		}
	}
	return false
}

// IsConnIn checks is connection's remote address is included.
func (r *Range) IsConnIn(conn net.Conn) bool {
	addr, _ := conn.RemoteAddr().(*net.TCPAddr)
	return r.IsIn(addr.IP)
}
