// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package bindata

import (
	"os"
	"path/filepath"
	"unicode"
)

//
// Asset holds information about a single asset to be processed.
//
type Asset struct {
	// Path contains full file path.
	Path string

	// Name contains key used in TOC -- name by which asset is referenced.
	Name string

	// Function name for the procedure returning the asset contents.
	Func string

	// fi field contains the file information (to minimize calling os.Stat
	// on the same file while processing).
	fi os.FileInfo
}

func normalize(in string) (out string) {
	up := true
	for _, r := range in {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if up {
				out += string(unicode.ToUpper(r))
				up = false
			} else {
				out += string(r)
			}
			continue
		}
		if r == '/' {
			up = true
		}
	}

	return
}

//
// NewAsset will create, initialize, and return new asset based on file
// path and real path if its symlink.
//
func NewAsset(path, name, realPath string, fi os.FileInfo) (a Asset) {
	a = Asset{
		Path: path,
		Name: filepath.ToSlash(name),
		fi:   fi,
	}

	if len(realPath) == 0 {
		a.Func = "bindata" + normalize(name)
	} else {
		a.Func = "bindata" + normalize(realPath)
	}

	return
}
