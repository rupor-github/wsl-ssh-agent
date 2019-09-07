package citrus

import (
	"net"
	"regexp"
	"strings"

	"github.com/atotto/clipboard"
)

// Clipboard is used by "lemonade" to rpc clipboard content.
type Clipboard struct {
	connCh chan net.Conn
	params ParamsValue
}

// NewClipboard initializes Clipboard structure.
func NewClipboard(ch chan net.Conn, params ParamsValue) *Clipboard {
	return &Clipboard{
		connCh: ch,
		params: params,
	}
}

// Copy is implementation of "lemonade" rpc "copy" command.
func (c *Clipboard) Copy(text string, _ *struct{}) error {
	<-c.connCh
	// Logger instance needs to be passed here somehow?
	return clipboard.WriteAll(convertLineEnding(text, c.params.LE))
}

// Paste is implementation of "lemonade" rpc "paste" command.
func (c *Clipboard) Paste(_ struct{}, resp *string) error {
	<-c.connCh
	t, err := clipboard.ReadAll()
	*resp = t
	return err
}

func convertLineEnding(text, option string) string {
	switch {
	case strings.EqualFold(option, "lf"):
		text = strings.Replace(text, "\r\n", "\n", -1)
		return strings.Replace(text, "\r", "\n", -1)
	case strings.EqualFold(option, "crlf"):
		text = regexp.MustCompile(`\r(.)|\r$`).ReplaceAllString(text, "\r\n$1")
		text = regexp.MustCompile(`([^\r])\n|^\n`).ReplaceAllString(text, "$1\r\n")
		return text
	default:
		return text
	}
}
