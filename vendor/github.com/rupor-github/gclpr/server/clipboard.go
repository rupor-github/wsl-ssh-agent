package server

import (
	"log"

	"github.com/atotto/clipboard"
)

// Clipboard is used to rpc clipboard content.
type Clipboard struct {
	leOP string
}

// NewClipboard initializes Clipboard structure.
func NewClipboard(le string) *Clipboard {
	return &Clipboard{leOP: le}
}

// Copy is implementation of rpc "copy" command.
func (c *Clipboard) Copy(text string, _ *struct{}) error {
	log.Printf("Copy request received len: %d\n", len(text))
	return clipboard.WriteAll(ConvertLE(text, c.leOP))
}

// Paste is implementation of rpc "paste" command.
func (c *Clipboard) Paste(_ struct{}, resp *string) error {
	t, err := clipboard.ReadAll()
	log.Printf("Paste request received len: %d, error: '%+v'\n", len(t), err)
	*resp = t
	return err
}
