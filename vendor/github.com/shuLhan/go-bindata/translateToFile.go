package bindata

import (
	"bufio"
	"fmt"
	"os"
)

// translateToFile generates one single file
func translateToFile(c *Config, toc []Asset) (err error) {
	// Create output file.
	fd, err := os.Create(c.Output)
	if err != nil {
		return err
	}

	// Create a buffered writer for better performance.
	bfd := bufio.NewWriter(fd)

	err = writeHeader(bfd, c, toc)
	if err != nil {
		goto out
	}

	// Write package declaration.
	_, err = fmt.Fprintf(bfd, "\npackage %s\n\n", c.Package)
	if err != nil {
		goto out
	}

	// Write assets.
	if c.Debug || c.Dev {
		err = writeDebug(bfd, c, toc)
	} else {
		err = writeRelease(bfd, c, toc)
	}

	if err != nil {
		goto out
	}

	// Write table of contents
	err = writeTOC(bfd, toc)
	if err != nil {
		goto out
	}

	// Write hierarchical tree of assets
	err = writeTOCTree(bfd, toc)
	if err != nil {
		return err
	}

	// Write restore procedure
	err = writeRestore(bfd)
out:
	return flushAndClose(fd, bfd, err)
}
