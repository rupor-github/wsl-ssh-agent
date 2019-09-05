package bindata

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

// translateToDir generates splited file
func translateToDir(c *Config, toc []Asset) error {
	if err := generateCommonFile(c, toc); err != nil {
		return err
	}

	for i := range toc {
		if err := generateOneAsset(c, &toc[i]); err != nil {
			return err
		}
	}

	return nil
}

func generateCommonFile(c *Config, toc []Asset) (err error) {
	// Create output file.
	fd, err := os.Create(filepath.Join(c.Output, DefOutputName))
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
		err = writeDebugHeader(bfd)
	} else {
		err = writeReleaseHeader(bfd, c)
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
		goto out
	}

	// Write restore procedure
	err = writeRestore(bfd)

out:
	return flushAndClose(fd, bfd, err)
}

func generateOneAsset(c *Config, a *Asset) (err error) {
	// Create output file.
	fd, err := os.Create(filepath.Join(c.Output, a.Func+".go"))
	if err != nil {
		return err
	}

	// Create a buffered writer for better performance.
	bfd := bufio.NewWriter(fd)

	// Write the header. This makes e.g. Github ignore diffs in generated files.
	_, err = fmt.Fprint(bfd, headerGeneratedBy)
	if err != nil {
		goto out
	}

	if _, err = fmt.Fprint(bfd, "// source: "); err != nil {
		goto out
	}

	if _, err = fmt.Fprintln(bfd, a.Path); err != nil {
		goto out
	}

	// Write build tags, if applicable.
	if len(c.Tags) > 0 {
		if _, err = fmt.Fprintf(bfd, "// +build %s\n\n", c.Tags); err != nil {
			goto out
		}
	}

	// Write package declaration.
	_, err = fmt.Fprintf(bfd, "package %s\n\n", c.Package)
	if err != nil {
		goto out
	}

	// Write assets.
	if c.Debug || c.Dev {
		err = writeOneFileDebug(bfd, c, a)
	} else {
		err = writeOneFileRelease(bfd, c, a)
	}
out:
	return flushAndClose(fd, bfd, err)
}
