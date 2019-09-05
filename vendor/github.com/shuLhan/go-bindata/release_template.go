package bindata

const tmplImport string = `
import (
	"os"
	"time"
)

`

const tmplImportCompressNomemcopy = `
import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data, name string) ([]byte, error) {
	gz, err := gzip.NewReader(strings.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

`

const tmplImportCompressMemcopy = `
import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

`

const tmplImportNocompressNomemcopy = `
import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
	"unsafe"
)

// nolint: deadcode, gas
func bindataRead(data, name string) ([]byte, error) {
	var empty [0]byte
	sx := (*reflect.StringHeader)(unsafe.Pointer(&data))
	b := empty[:]
	bx := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bx.Data = sx.Data
	bx.Len = len(data)
	bx.Cap = bx.Len
	return b, nil
}

`

const tmplImportNocompressMemcopy = `
import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

`

const tmplReleaseHeader = `
type asset struct {
	bytes []byte
	info  fileInfoEx
}

type fileInfoEx interface {
	os.FileInfo
	MD5Checksum() string
}

type bindataFileInfo struct {
	name        string
	size        int64
	mode        os.FileMode
	modTime     time.Time
	md5checksum string
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) MD5Checksum() string {
	return fi.md5checksum
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

`

const tmplFuncCompressNomemcopy string = `"

func %sBytes() ([]byte, error) {
	return bindataRead(
		_%s,
		%q,
	)
}

`

const tmplFuncCompressMemcopy string = `")

func %sBytes() ([]byte, error) {
	return bindataRead(
		_%s,
		%q,
	)
}

`

const tmplFuncNocompressNomemcopy string = `"

func %sBytes() ([]byte, error) {
	return bindataRead(
		_%s,
		%q,
	)
}

`

const tmplFuncNocompressMemcopy string = `)

func %sBytes() ([]byte, error) {
	return _%s, nil
}

`

const tmplReleaseCommon string = `

func %s() (*asset, error) {
	bytes, err := %sBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name: %q,
		size: %d,
		md5checksum: %q,
		mode: os.FileMode(%d),
		modTime: time.Unix(%d, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

`
