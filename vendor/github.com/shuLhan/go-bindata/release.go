// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package bindata

// nolint: gas
import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"unicode/utf8"
)

// writeOneFileRelease writes the release code file for each file (when splited file).
func writeOneFileRelease(w io.Writer, c *Config, a *Asset) (err error) {
	_, err = fmt.Fprint(w, tmplImport)
	if err != nil {
		return
	}

	return writeReleaseAsset(w, c, a)
}

// writeRelease writes the release code file for single file.
func writeRelease(w io.Writer, c *Config, toc []Asset) (err error) {
	err = writeReleaseHeader(w, c)
	if err != nil {
		return
	}

	for i := range toc {
		err = writeReleaseAsset(w, c, &toc[i])
		if err != nil {
			return
		}
	}

	return
}

// writeReleaseHeader writes output file headers.
// This targets release builds.
func writeReleaseHeader(w io.Writer, c *Config) (err error) {
	if c.NoCompress {
		if c.NoMemCopy {
			_, err = fmt.Fprint(w, tmplImportNocompressNomemcopy)
		} else {
			_, err = fmt.Fprint(w, tmplImportNocompressMemcopy)
		}
	} else {
		if c.NoMemCopy {
			_, err = fmt.Fprint(w, tmplImportCompressNomemcopy)
		} else {
			_, err = fmt.Fprint(w, tmplImportCompressMemcopy)
		}
	}
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, tmplReleaseHeader)

	return
}

// writeReleaseAsset write a release entry for the given asset.
// A release entry is a function which embeds and returns
// the file's byte content.
func writeReleaseAsset(w io.Writer, c *Config, asset *Asset) (err error) {
	fd, err := os.Open(asset.Path)
	if err != nil {
		return
	}

	if c.NoCompress {
		if c.NoMemCopy {
			err = nocompressNomemcopy(w, asset, fd)
		} else {
			err = nocompressMemcopy(w, asset, fd)
		}
	} else {
		if c.NoMemCopy {
			err = compressNomemcopy(w, asset, fd)
		} else {
			err = compressMemcopy(w, asset, fd)
		}
	}
	if err != nil {
		_ = fd.Close()
		return
	}

	err = fd.Close()
	if err != nil {
		return
	}

	return assetReleaseCommon(w, c, asset)
}

var (
	backquote = []byte("`")
	bom       = []byte("\xEF\xBB\xBF")
)

// sanitize prepares a valid UTF-8 string as a raw string constant.
// Based on https://code.google.com/p/go/source/browse/godoc/static/makestatic.go?repo=tools
func sanitize(b []byte) []byte {
	var chunks [][]byte
	for i, b := range bytes.Split(b, backquote) {
		if i > 0 {
			chunks = append(chunks, backquote)
		}
		for j, c := range bytes.Split(b, bom) {
			if j > 0 {
				chunks = append(chunks, bom)
			}
			if len(c) > 0 {
				chunks = append(chunks, c)
			}
		}
	}

	var buf bytes.Buffer
	sanitizeChunks(&buf, chunks)
	return buf.Bytes()
}

func sanitizeChunks(buf *bytes.Buffer, chunks [][]byte) {
	n := len(chunks)
	if n >= 2 {
		buf.WriteString("(")
		sanitizeChunks(buf, chunks[:n/2])
		buf.WriteString(" + ")
		sanitizeChunks(buf, chunks[n/2:])
		buf.WriteString(")")
		return
	}
	b := chunks[0]
	if bytes.Equal(b, backquote) {
		buf.WriteString("\"`\"")
		return
	}
	if bytes.Equal(b, bom) {
		buf.WriteString(`"\xEF\xBB\xBF"`)
		return
	}
	buf.WriteString("`")
	buf.Write(b)
	buf.WriteString("`")
}

func compressNomemcopy(w io.Writer, asset *Asset, r io.Reader) (err error) {
	_, err = fmt.Fprintf(w, "var _%s = \"\" +\n\t\"", asset.Func)
	if err != nil {
		return
	}

	gz := gzip.NewWriter(&StringWriter{Writer: w})
	_, err = io.Copy(gz, r)
	if err != nil {
		_ = gz.Close()
		return
	}

	err = gz.Close()
	if err != nil {
		return
	}

	_, err = fmt.Fprintf(w, tmplFuncCompressNomemcopy, asset.Func,
		asset.Func, asset.Name)

	return
}

func compressMemcopy(w io.Writer, asset *Asset, r io.Reader) (err error) {
	_, err = fmt.Fprintf(w, "var _%s = []byte(\n\t\"", asset.Func)
	if err != nil {
		return err
	}

	gz := gzip.NewWriter(&StringWriter{Writer: w})
	_, err = io.Copy(gz, r)
	if err != nil {
		_ = gz.Close()
		return err
	}

	err = gz.Close()
	if err != nil {
		return
	}

	_, err = fmt.Fprintf(w, tmplFuncCompressMemcopy, asset.Func,
		asset.Func, asset.Name)

	return
}

func nocompressNomemcopy(w io.Writer, asset *Asset, r io.Reader) (err error) {
	_, err = fmt.Fprintf(w, `var _%s = "`, asset.Func)
	if err != nil {
		return
	}

	_, err = io.Copy(&StringWriter{Writer: w}, r)
	if err != nil {
		return
	}

	_, err = fmt.Fprintf(w, tmplFuncNocompressNomemcopy, asset.Func,
		asset.Func, asset.Name)

	return
}

func nocompressMemcopy(w io.Writer, asset *Asset, r io.Reader) (err error) {
	_, err = fmt.Fprintf(w, `var _%s = []byte(`, asset.Func)
	if err != nil {
		return
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}

	if utf8.Valid(b) && !bytes.Contains(b, []byte{0}) {
		_, err = w.Write(sanitize(b))
	} else {
		_, err = fmt.Fprintf(w, "%+q", b)
	}
	if err != nil {
		return
	}

	_, err = fmt.Fprintf(w, tmplFuncNocompressMemcopy, asset.Func,
		asset.Func)

	return
}

// nolint: gas
func assetReleaseCommon(w io.Writer, c *Config, asset *Asset) (err error) {
	fi, err := os.Stat(asset.Path)
	if err != nil {
		return
	}

	mode := uint(fi.Mode())
	modTime := fi.ModTime().Unix()
	size := fi.Size()
	if c.NoMetadata {
		mode = 0
		modTime = 0
		size = 0
	}
	if c.Mode > 0 {
		mode = uint(os.ModePerm) & c.Mode
	}
	if c.ModTime > 0 {
		modTime = c.ModTime
	}

	var md5checksum string
	if c.MD5Checksum {
		var buf []byte

		buf, err = ioutil.ReadFile(asset.Path)
		if err != nil {
			return
		}

		h := md5.New()
		if _, err = h.Write(buf); err != nil {
			return
		}
		md5checksum = fmt.Sprintf("%x", h.Sum(nil))
	}

	_, err = fmt.Fprintf(w, tmplReleaseCommon, asset.Func, asset.Func,
		asset.Name, size, md5checksum, mode, modTime)

	return
}
