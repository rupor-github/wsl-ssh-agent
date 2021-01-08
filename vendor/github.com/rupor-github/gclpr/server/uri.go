package server

import (
	"log"

	"github.com/skratchdot/open-golang/open"
)

// URI is used to rpc open command.
type URI struct {
	// placeholder
}

// NewURI initializes URI structure.
func NewURI() *URI {
	return &URI{}
}

// Open is implementation of "lemonade" rpc "open" command.
func (u *URI) Open(uri string, _ *struct{}) error {
	log.Printf("URI Open received: '%s'", uri)
	return open.Run(uri)
}
