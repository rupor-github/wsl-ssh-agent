// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain
// Dedication license.  Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package bindata

import (
	"bufio"
	"fmt"
	"io"
)

func writeHeader(bfd io.Writer, c *Config, toc []Asset) (err error) {
	// Write the header. This makes e.g. Github ignore diffs in generated files.
	_, err = fmt.Fprint(bfd, headerGeneratedBy)
	if err != nil {
		return
	}

	if c.Split {
		_, err = fmt.Fprint(bfd, "// -- Common file --\n")
		if err != nil {
			return
		}
	} else {
		_, err = fmt.Fprint(bfd, "// sources:\n")
		if err != nil {
			return
		}

		for _, asset := range toc {
			_, err = fmt.Fprintf(bfd, "// %s\n", asset.Path)
			if err != nil {
				return
			}
		}
	}

	// Write build tags, if applicable.
	if len(c.Tags) > 0 {
		if _, err = fmt.Fprintf(bfd, "// +build %s\n\n", c.Tags); err != nil {
			return
		}
	}

	return
}

//
// flushAndClose will flush the buffered writer `bfd` and close the file `fd`.
//
func flushAndClose(fd io.Closer, bfd *bufio.Writer, errParam error) (err error) {
	err = errParam

	if err == nil {
		err = bfd.Flush()
	}

	errClose := fd.Close()
	if errClose != nil {
		if err == nil {
			err = errClose
		}
	}

	return

}
