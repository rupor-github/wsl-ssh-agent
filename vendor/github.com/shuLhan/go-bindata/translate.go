// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package bindata

import (
	"os"
)

// Translate reads assets from an input directory, converts them
// to Go code and writes new files to the output specified
// in the given configuration.
func Translate(c *Config) (err error) {
	c.cwd, err = os.Getwd()
	if err != nil {
		return ErrCWD
	}

	// Ensure our configuration has sane values.
	err = c.validate()
	if err != nil {
		return
	}

	scanner := NewFSScanner(c)

	assets := make([]Asset, 0)

	// Locate all the assets.
	for _, input := range c.Input {
		err = scanner.Scan(input.Path, "", input.Recursive)
		if err != nil {
			return
		}

		assets = append(assets, scanner.assets...)

		scanner.Reset()
	}

	if c.Split {
		return translateToDir(c, assets)
	}

	return translateToFile(c, assets)
}
