// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package bindata

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type assetTree struct {
	Asset    Asset
	Children map[string]*assetTree
}

func newAssetTree() *assetTree {
	tree := &assetTree{}
	tree.Children = make(map[string]*assetTree)
	return tree
}

func (root *assetTree) child(name string) *assetTree {
	rv, ok := root.Children[name]
	if !ok {
		rv = newAssetTree()
		root.Children[name] = rv
	}
	return rv
}

func (root *assetTree) Add(route []string, asset Asset) {
	for _, name := range route {
		root = root.child(name)
	}
	root.Asset = asset
}

func ident(w io.Writer, n int) (err error) {
	for i := 0; i < n; i++ {
		_, err = w.Write([]byte{'\t'})
		if err != nil {
			return
		}
	}
	return
}

func (root *assetTree) funcOrNil() string {
	if root.Asset.Func == "" {
		return "nil"
	}

	return root.Asset.Func
}

//
// getFilenames will return all files sorted, to make output stable between
// invocations.
//
func (root *assetTree) getFilenames() (filenames []string) {
	filenames = make([]string, len(root.Children))
	x := 0
	for filename := range root.Children {
		filenames[x] = filename
		x++
	}
	sort.Strings(filenames)

	return
}

func (root *assetTree) writeGoMap(w io.Writer, nident int) (err error) {
	fmt.Fprintf(w, tmplBinTreeValues, root.funcOrNil())

	if len(root.Children) > 0 {
		_, err = io.WriteString(w, "\n")
		if err != nil {
			return
		}

		filenames := root.getFilenames()

		for _, p := range filenames {
			err = ident(w, nident+1)
			if err != nil {
				return
			}
			fmt.Fprintf(w, `"%s": `, p)

			err = root.Children[p].writeGoMap(w, nident+1)
			if err != nil {
				return
			}
		}

		err = ident(w, nident)
		if err != nil {
			return
		}
	}

	_, err = io.WriteString(w, "}}")
	if err != nil {
		return
	}

	if nident > 0 {
		_, err = io.WriteString(w, ",")
		if err != nil {
			return
		}
	}

	_, err = io.WriteString(w, "\n")

	return
}

func (root *assetTree) WriteAsGoMap(w io.Writer) (err error) {
	_, err = fmt.Fprint(w, tmplTypeBintree)
	if err != nil {
		return
	}

	return root.writeGoMap(w, 0)
}

func writeTOCTree(w io.Writer, toc []Asset) error {
	_, err := fmt.Fprint(w, tmplFuncAssetDir)
	if err != nil {
		return err
	}
	tree := newAssetTree()
	for i := range toc {
		pathList := strings.Split(toc[i].Name, "/")
		tree.Add(pathList, toc[i])
	}

	return tree.WriteAsGoMap(w)
}

//
// getLongestAssetNameLen will return length of the longest asset name in toc.
//
func getLongestAssetNameLen(toc []Asset) (longest int) {
	for _, asset := range toc {
		lenName := len(asset.Name)
		if lenName > longest {
			longest = lenName
		}
	}

	return
}

// writeTOC writes the table of contents file.
func writeTOC(w io.Writer, toc []Asset) (err error) {
	_, err = fmt.Fprint(w, tmplFuncAsset)
	if err != nil {
		return err
	}

	longestNameLen := getLongestAssetNameLen(toc)

	for i := range toc {
		err = writeTOCAsset(w, &toc[i], longestNameLen)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprint(w, "}\n")

	return
}

// writeTOCAsset write a TOC entry for the given asset.
func writeTOCAsset(w io.Writer, asset *Asset, longestNameLen int) (err error) {
	toWrite := " "

	for x := 0; x < longestNameLen-len(asset.Name); x++ {
		toWrite += " "
	}

	toWrite = "\t\"" + asset.Name + "\":" + toWrite + asset.Func + ",\n"

	_, err = io.WriteString(w, toWrite)

	return
}
