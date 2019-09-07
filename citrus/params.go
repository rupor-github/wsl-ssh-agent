package citrus

import (
	"fmt"
	"strconv"
	"strings"
)

// ParamsValue implements special flag to parse "lemonade" server arguments.
type ParamsValue struct {
	wasSet int
	Port   int
	Allow  string
	LE     string
}

func (v *ParamsValue) String() string {

	if v == nil || v.wasSet == 0 {
		return ""
	}

	var res string
	if v.Port > 0 {
		res += strconv.FormatInt(int64(v.Port), 10)
	}
	if v.wasSet > 1 {
		res += ";"
	}
	if len(v.Allow) > 0 {
		res += v.Allow
	}
	if v.wasSet > 2 {
		res += ";"
	}
	if len(v.LE) > 0 {
		res += v.LE
	}
	return res
}

// Set parses "lemonade" arguments and stores them for later use.
func (v *ParamsValue) Set(s string) error {

	const (
		port = iota
		allow
		le
		maxParams
	)

	parts := strings.Split(s, ";")

	if len(parts) > maxParams {
		return fmt.Errorf("wrong lemonade parameters number: %d, should be 3", len(parts))
	}
	v.wasSet = len(parts)
	for i := range parts {
		switch i {
		case port:
			if len(parts[port]) > 0 {
				p, err := strconv.Atoi(parts[port])
				if err != nil {
					return fmt.Errorf("unable to parse lemonade port: '%s' error: %w", parts[port], err)
				}
				v.Port = p
			}
		case allow:
			if len(parts[allow]) > 0 {
				v.Allow = parts[allow]
			}
		case le:
			if len(parts[le]) > 0 {
				v.LE = parts[le]
			}
		}
	}
	return nil
}

// IsSet returns true is Set() method was called on value.
func (v *ParamsValue) IsSet() bool {
	return v.wasSet > 0
}
