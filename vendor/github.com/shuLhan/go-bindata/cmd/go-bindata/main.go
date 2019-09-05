// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/shuLhan/go-bindata"
)

const (
	appName         = "go-bindata"
	appVersionMajor = 3
	appVersionMinor = 3
)

var (
	//
	// AppVersionRev part of the program version.
	//
	// This will be set automatically at build time like so:
	//
	//     go build -ldflags "-X main.AppVersionRev `date -u +%s`" (go version < 1.5)
	//     go build -ldflags "-X main.AppVersionRev=`date -u +%s`" (go version >= 1.5)
	AppVersionRev string

	lerr = log.New(os.Stderr, "", 0)
	lout = log.New(os.Stdout, "", 0)
)

// List of error messages.
var (
	ErrInvalidIgnoreRegex  = errors.New("Invalid -ignore regex pattern")
	ErrInvalidIncludeRegex = errors.New("Invalid -include regex pattern")
	ErrInvalidPrefixRegex  = errors.New("Invalid -prefix regex pattern")
	ErrNoInput             = errors.New("Missing <input directories>")
)

// List of local variables.
var (
	argIgnore  []string
	argInclude []string
	argVersion bool
	argPrefix  string
	cfg        *bindata.Config
)

func main() {
	initArgs()

	err := parseArgs()
	if err != nil {
		lerr.Println(err)

		if err == ErrNoInput {
			lerr.Println()
			usage()
		}

		os.Exit(2)
	}

	err = bindata.Translate(cfg)
	if err != nil {
		lerr.Println("bindata: ", err)
		os.Exit(1)
	}
}

func usage() {
	lerr.Println("Usage: " + appName + " [options] <input directories>\n")

	flag.PrintDefaults()
}

func version() {
	if len(AppVersionRev) == 0 {
		AppVersionRev = "0"
	}

	lout.Printf("%s %d.%d.%s (Go runtime %s).\n", appName, appVersionMajor,
		appVersionMinor, AppVersionRev, runtime.Version())
	lout.Println("Copyright (c) 2010-2015, Jim Teeuwen.")

	os.Exit(0)
}

//
// initArgs will initialize all command line arguments.
//
func initArgs() {
	cfg = bindata.NewConfig()

	flag.Usage = usage

	flag.BoolVar(&argVersion, "version", false, "Displays version information.")
	flag.BoolVar(&cfg.Debug, "debug", cfg.Debug, "Do not embed the assets, but provide the embedding API. Contents will still be loaded from disk.")
	flag.BoolVar(&cfg.Dev, "dev", cfg.Dev, "Similar to debug, but does not emit absolute paths. Expects a rootDir variable to already exist in the generated code's package.")
	flag.BoolVar(&cfg.MD5Checksum, "md5checksum", cfg.MD5Checksum, "MD5 checksums will be calculated for assets.")
	flag.BoolVar(&cfg.NoCompress, "nocompress", cfg.NoCompress, "Assets will *not* be GZIP compressed when this flag is specified.")
	flag.BoolVar(&cfg.NoMemCopy, "nomemcopy", cfg.NoMemCopy, "Use a .rodata hack to get rid of unnecessary memcopies. Refer to the documentation to see what implications this carries.")
	flag.BoolVar(&cfg.NoMetadata, "nometadata", cfg.NoMetadata, "Assets will not preserve size, mode, and modtime info.")
	flag.BoolVar(&cfg.Split, "split", cfg.NoMetadata, "Split output into several files, avoiding to have a big output file.")
	flag.Int64Var(&cfg.ModTime, "modtime", cfg.ModTime, "Optional modification unix timestamp override for all files.")
	flag.StringVar(&argPrefix, "prefix", "", "Optional path prefix to strip off asset names.")
	flag.StringVar(&cfg.Output, "o", cfg.Output, "Optional name of the output file to be generated.")
	flag.StringVar(&cfg.Package, "pkg", cfg.Package, "Package name to use in the generated code.")
	flag.StringVar(&cfg.Tags, "tags", cfg.Tags, "Optional set of build tags to include.")
	flag.UintVar(&cfg.Mode, "mode", cfg.Mode, "Optional file mode override for all files.")
	flag.Var((*AppendSliceValue)(&argIgnore), "ignore", "Regex pattern to ignore")
	flag.Var((*AppendSliceValue)(&argInclude), "include", "Regex pattern to include")
}

//
// parseArgs creates a new, filled configuration instance by reading and parsing
// command line options.
//
// The order of parsing is important to minimize unneeded processing, i.e.,
//
// (1) checking for version argument must be first,
// (2) followed by checking input directory argument, and then everything else.
//
// If no input directory or one of the command line options are incorrect, it
// will return error.
//
func parseArgs() (err error) {
	flag.Parse()

	// (1)
	if argVersion {
		version()
	}

	// (2)
	if flag.NArg() == 0 {
		return ErrNoInput
	}

	err = parsePrefix()
	if err != nil {
		return
	}

	err = parseIgnore()
	if err != nil {
		return
	}

	err = parseInclude()
	if err != nil {
		return
	}

	parseOutputPkg()

	// Create input configurations.
	cfg.Input = make([]bindata.InputConfig, flag.NArg())
	for i := range cfg.Input {
		cfg.Input[i] = parseInput(flag.Arg(i))
	}

	return
}

func parsePrefix() (err error) {
	if len(argPrefix) == 0 {
		return
	}

	cfg.Prefix, err = regexp.Compile(argPrefix)
	if err != nil {
		return ErrInvalidPrefixRegex
	}

	return
}

func parseIgnore() (err error) {
	var ignoreVal *regexp.Regexp

	for _, pattern := range argIgnore {
		ignoreVal, err = regexp.Compile(pattern)
		if err != nil {
			return ErrInvalidIgnoreRegex
		}

		cfg.Ignore = append(cfg.Ignore, ignoreVal)
	}

	return
}

func parseInclude() (err error) {
	var includeVal *regexp.Regexp

	for _, pattern := range argInclude {
		includeVal, err = regexp.Compile(pattern)
		if err != nil {
			return ErrInvalidIncludeRegex
		}

		cfg.Include = append(cfg.Include, includeVal)
	}

	return
}

//
// parseOutputPkg will change package name to directory of output, only if
// output flag is set and package flag is not set.
//
func parseOutputPkg() {
	var isPkgSet, isOutputSet bool

	flag.Visit(func(f *flag.Flag) {
		if f.Name == "pkg" {
			isPkgSet = true
			return
		}
		if f.Name == "o" {
			isOutputSet = true
		}
	})

	if isOutputSet && !isPkgSet {
		pkg := filepath.Base(filepath.Dir(cfg.Output))
		if pkg != "." && pkg != "/" {
			cfg.Package = pkg
		}
	}
}

//
// parseInput determines whether the given path has a recursive indicator
// ("/...") and returns a new path with the recursive indicator chopped off if
// it does.
//
//  ex:
//      /path/to/foo/...    -> (/path/to/foo, true)
//      /path/to/bar        -> (/path/to/bar, false)
//
func parseInput(path string) (inputConfig bindata.InputConfig) {
	if strings.HasSuffix(path, "/...") {
		inputConfig.Path = filepath.Clean(path[:len(path)-4])
		inputConfig.Recursive = true
	} else {
		inputConfig.Path = filepath.Clean(path)
	}

	return
}
