// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain
// Dedication license.  Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package bindata

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// ByName implement sort.Interface for []os.FileInfo based on Name()
type ByName []os.FileInfo

func (v ByName) Len() int           { return len(v) }
func (v ByName) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v ByName) Less(i, j int) bool { return v[i].Name() < v[j].Name() }

//
// FSScanner implement the file system scanner.
//
type FSScanner struct {
	cfg         *Config
	knownFuncs  map[string]int
	visitedDirs map[string]bool
	assets      []Asset
}

//
// NewFSScanner will create and initialize new file system scanner.
//
func NewFSScanner(cfg *Config) (fss *FSScanner) {
	fss = &FSScanner{
		cfg: cfg,
	}

	fss.Reset()

	return
}

//
// Reset will clear all previous mapping and assets.
//
func (fss *FSScanner) Reset() {
	fss.knownFuncs = make(map[string]int)
	fss.visitedDirs = make(map[string]bool)
	fss.assets = make([]Asset, 0)
}

//
// isIgnored will return,
// (1) true, if `path` is matched with one of ignore-pattern,
// (2) false, if `path` is matched with one of include-pattern,
// (3) true, if include-pattern is defined but no matched found.
//
func (fss *FSScanner) isIgnored(path string) bool {
	// (1)
	for _, re := range fss.cfg.Ignore {
		if re.MatchString(path) {
			return true
		}
	}

	// (2)
	for _, re := range fss.cfg.Include {
		if re.MatchString(path) {
			return false
		}
	}

	// (3)
	return len(fss.cfg.Include) > 0
}

func (fss *FSScanner) cleanPrefix(path string) string {
	if fss.cfg.Prefix == nil {
		return path
	}

	return fss.cfg.Prefix.ReplaceAllString(path, "")
}

//
// addAsset will add new asset based on path, realPath, and file info.  The
// path can be a directory or file. Realpath reference to the original path if
// path is symlink, if path is not symlink then path and realPath will be equal.
//
func (fss *FSScanner) addAsset(path, realPath string, fi os.FileInfo) {
	name := fss.cleanPrefix(path)

	asset := NewAsset(path, name, realPath, fi)

	num, ok := fss.knownFuncs[asset.Func]
	if ok {
		fss.knownFuncs[asset.Func] = num + 1
		asset.Func = fmt.Sprintf("%s_%d", asset.Func, num)
	} else {
		fss.knownFuncs[asset.Func] = 2
	}

	fss.assets = append(fss.assets, asset)
}

//
// getListFileInfo will return list of files in `path`.
//
// (1) set the path visited status to true,
// (2) read all files inside directory into list, and
// (3) sort the list to make output stable between invocations.
//
func (fss *FSScanner) getListFileInfo(path string) (
	list []os.FileInfo, err error,
) {
	// (1)
	fss.visitedDirs[path] = true

	// (2)
	fd, err := os.Open(path)
	if err != nil {
		_ = fd.Close()
		return
	}

	list, err = fd.Readdir(0)
	if err != nil {
		_ = fd.Close()
		return
	}

	err = fd.Close()
	if err != nil {
		return
	}

	// (3)
	sort.Sort(ByName(list))

	return
}

//
// scanSymlink reads the real-path from symbolic link file and converts the path
// and real-path into a relative path.
//
func (fss *FSScanner) scanSymlink(path string, recursive bool) (
	err error,
) {
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return
	}

	fi, err := os.Lstat(realPath)
	if err != nil {
		return
	}

	if fi.Mode().IsRegular() {
		fss.addAsset(path, realPath, fi)
		return
	}

	if !recursive {
		return
	}

	_, ok := fss.visitedDirs[realPath]
	if ok {
		return
	}

	list, err := fss.getListFileInfo(path)
	if err != nil {
		return
	}

	for _, fi = range list {
		filePath := filepath.Join(path, fi.Name())
		fileRealPath := filepath.Join(realPath, fi.Name())

		err = fss.Scan(filePath, fileRealPath, recursive)
		if err != nil {
			return
		}
	}

	return
}

//
// Scan will scan the file or content of directory in `path`.
//
func (fss *FSScanner) Scan(path, realPath string, recursive bool) (err error) {
	path = filepath.Clean(path)

	if fss.isIgnored(path) {
		return
	}

	fi, err := os.Lstat(path)
	if err != nil {
		return
	}

	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		err = fss.scanSymlink(path, recursive)
		return
	}

	if fi.Mode().IsRegular() {
		fss.addAsset(path, realPath, fi)
		return
	}

	if !recursive {
		return
	}

	_, ok := fss.visitedDirs[path]
	if ok {
		return
	}

	list, err := fss.getListFileInfo(path)
	if err != nil {
		return
	}

	for _, fi = range list {
		filePath := filepath.Join(path, fi.Name())

		err = fss.Scan(filePath, "", recursive)
		if err != nil {
			return
		}
	}

	return
}
