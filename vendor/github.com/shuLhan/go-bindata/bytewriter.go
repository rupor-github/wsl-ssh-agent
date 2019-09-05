// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package bindata

import (
	"fmt"
	"io"
)

var (
	newline    = []byte{'\n'}
	dataindent = []byte{'\t', '\t'}
	space      = []byte{' '}
)

//
// ByteWriter define a writer to write content of file.
//
type ByteWriter struct {
	io.Writer
	c int
}

func (w *ByteWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}

	for n = range p {
		if w.c%12 == 0 {
			_, err = w.Writer.Write(newline)
			if err != nil {
				return
			}

			_, err = w.Writer.Write(dataindent)
			if err != nil {
				return
			}

			w.c = 0
		} else {
			_, err = w.Writer.Write(space)
			if err != nil {
				return
			}
		}

		_, err = fmt.Fprintf(w.Writer, "0x%02x,", p[n])
		if err != nil {
			return
		}
		w.c++
	}

	n++

	return
}
