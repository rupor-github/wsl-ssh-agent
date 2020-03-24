package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"

	"github.com/Microsoft/go-winio"
	si "github.com/allan-simon/go-singleinstance"

	"github.com/rupor-github/wsl-ssh-agent/citrus"
	"github.com/rupor-github/wsl-ssh-agent/misc"
	"github.com/rupor-github/wsl-ssh-agent/static"
	"github.com/rupor-github/wsl-ssh-agent/systray"
	"github.com/rupor-github/wsl-ssh-agent/util"
)

const (
	title   = "wsl-ssh-agent-gui"
	tooltip = "Helper to interface with Windows ssh-agent.exe service from WSL"
)

var (
	// LastGitCommit is used during build to inject git sha
	LastGitCommit string
	// Program arguments
	debug      bool
	help       bool
	ignorelock bool
	socketName string
	pipeName   string
	setenv     bool
	envName    = "SSH_AUTH_SOCK"
	lemon      = citrus.ParamsValue{
		Port:  2489,
		Allow: "0.0.0.0/0,::/0",
	}
	usage  string
	locked int32
	cli    = flag.NewFlagSet(title, flag.ContinueOnError)
)

func onReady() {

	systray.SetTitle(title)
	systray.SetTooltip(tooltip)

	icon, err := static.Asset("icon.ico")
	if err == nil {
		systray.SetIcon(icon)
	}

	help := systray.AddMenuItem("About", "Shows application help")
	systray.AddSeparator()
	quit := systray.AddMenuItem("Exit", "Exits application")

	go func() {
		for {
			select {
			case <-help.ClickedCh:
				cli.Usage()
			case <-quit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onSession(e systray.SessionEvent) {
	if debug {
		log.Printf("Session event %s", e)
	}
	switch e {
	case systray.SesLock:
		atomic.StoreInt32(&locked, 1)
	case systray.SesUnlock:
		atomic.StoreInt32(&locked, 0)
	}
}

func onExit() {
	log.Print("Exiting systray")
}

func makeSocketName() (string, error) {
	f, err := ioutil.TempFile("", "ssh-*.sock")
	if err != nil {
		return "", err
	}
	defer f.Close()
	return f.Name(), nil
}

func serve(ln net.Listener, pipeName string, query func(name string, req []byte) (resp []byte, err error)) {

	var badResponse = [...]byte{0, 0, 0, 1, 5}

	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Listener Accept error on %s - '%s'", pipeName, err)
			return
		}
		go func(conn net.Conn, debug bool) {
			defer conn.Close()

			handle := fmt.Sprintf("%v", conn)

			if debug {
				log.Printf("[%s] Incoming: %s", handle, conn.LocalAddr())
			}

			reader := bufio.NewReader(conn)
			for {
				if debug {
					log.Printf("[%s] Reading loop", handle)
				}

				lenBuf := make([]byte, 4)
				_, err := io.ReadFull(reader, lenBuf)
				if err != nil {
					if debug {
						log.Printf("[%s] ReadFull error '%s'", handle, err)
					}
					return
				} else if debug {
					log.Printf("[%s] Got msg length for request to ssh-agent.exe: %s)", handle, fmt.Sprintf("%+v", lenBuf))
				}

				l := binary.BigEndian.Uint32(lenBuf)
				buf := make([]byte, l)
				_, err = io.ReadFull(reader, buf)
				if err != nil {
					if debug {
						log.Printf("[%s] ReadFull error '%s'", handle, err)
					}
					return
				} else if debug {
					log.Printf("[%s] Got request for query: %d)", handle, len(buf))
				}

				var res []byte
				if !ignorelock && atomic.LoadInt32(&locked) == 1 {
					log.Print("Session is locked")
					res = badResponse[:]
				} else {
					res, err = query(pipeName, append(lenBuf, buf...))
					if err != nil {
						// If for some reason talking to ssh-agent.exe failed send back error
						if debug {
							log.Printf("[%s] query error '%s'", handle, err)
						}
						res = badResponse[:]
					} else if debug {
						log.Printf("[%s] Got query response: %d bytes", handle, len(res))
					}
				}

				_, err = conn.Write(res)
				if err != nil {
					if debug {
						log.Printf("[%s] Conn.Write error '%s'", handle, err)
					}
					return
				} else if debug {
					log.Printf("[%s] Sent query response back", handle)
				}
			}
		}(conn, debug)
	}
}

func queryAgent(pipeName string, buf []byte) (result []byte, err error) {

	conn, err := winio.DialPipe(pipeName, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to pipe %s: %w", pipeName, err)
	} else if debug {
		log.Printf("Connected to %s: %d", pipeName, len(buf))
	}
	defer conn.Close()

	l, err := conn.Write(buf)
	if err != nil {
		return nil, fmt.Errorf("cannot write to pipe %s: %w", pipeName, err)
	} else if debug {
		log.Printf("Sent to %s: %d", pipeName, l)
	}

	reader := bufio.NewReader(conn)
	res := make([]byte, util.MaxAgentMsgLen)

	l, err = reader.Read(res)
	if err != nil {
		return nil, fmt.Errorf("cannot read from pipe %s: %w", pipeName, err)
	} else if debug {
		log.Printf("Received from %s: %d", pipeName, l)
	}
	return res[0:l], nil
}

func run() (err error) {

	if len(socketName) == 0 {
		socketName, err = makeSocketName()
		if err != nil {
			return fmt.Errorf("unable to create socket name: %w", err)
		}
	}
	if len(socketName) > util.MaxNameLen {
		return fmt.Errorf("socket name is too long: %d, max allowed: %d", len(socketName), util.MaxNameLen)
	}
	if !filepath.IsAbs(socketName) {
		return errors.New("socket name must be absolute path")
	}

	if setenv {
		if err := util.PrepareUserEnvironment(envName, socketName, debug); err != nil {
			return fmt.Errorf("unable to prepare user environment: %w", err)
		}
		defer func() {
			if err := util.CleanUserEnvironment(envName, debug); err != nil {
				log.Printf("Unable to clean user environment: %s", err.Error())
			}
		}()
	}

	if len(pipeName) == 0 {
		pipeName = util.AgentPipeName
	}

	_, err = os.Stat(socketName)
	if err == nil || !os.IsNotExist(err) {
		err = syscall.Unlink(socketName)
		if err != nil {
			return fmt.Errorf("failed to unlink socket %s: %w", socketName, err)
		}
	}

	sock, err := net.Listen("unix", socketName)
	if err != nil {
		return fmt.Errorf("could not open socket %s: %w", socketName, err)
	}
	defer func() {
		sock.Close()
		// Just in case - should not be needed
		_ = os.Remove(socketName)
	}()
	log.Printf("Listening on Unix socket: %s", socketName)

	go func() {
		serve(sock, pipeName, queryAgent)
		// If for some reason process breaks - exit
		log.Printf("Quiting - serve on %s ended", socketName)
		systray.Quit()
	}()

	systray.Run(onReady, onExit, onSession)
	return nil
}

func main() {

	// Redirect all logging to OutputDebugString()
	dlog := util.NewDebugWriter()
	log.SetOutput(dlog)
	log.SetPrefix("[" + title + "] ")
	log.SetFlags(0)

	// Prepare help and parse arguments

	cli.StringVar(&socketName, "socket", "", fmt.Sprintf("Auth socket `path` (max %d characters)", util.MaxNameLen))
	cli.StringVar(&pipeName, "pipe", "", "Pipe `name` used by Windows ssh-agent.exe")
	cli.StringVar(&envName, "envname", "SSH_AUTH_SOCK", "Environment variable `name` to hold socket path")
	cli.BoolVar(&setenv, "setenv", false, "Export environment variable with 'envname' and modify WSLENV")
	cli.BoolVar(&ignorelock, "nolock", false, "Provide access to ss-agent.exe even when user session is locked")
	cli.BoolVar(&help, "help", false, "Show help")
	cli.BoolVar(&debug, "debug", false, "Enable verbose debug logging")
	cli.Var(&lemon, "lemonade", "Semicolon separated `list` of lemonade \"server\" options (TCP port, Allow IP Range, Line Endings)")

	// Build usage string
	var buf strings.Builder
	cli.SetOutput(&buf)
	fmt.Fprintf(&buf, "\n%s\n\nVersion:\n\t%s (%s)\n\t%s\n\n", tooltip, misc.GetVersion(), runtime.Version(), LastGitCommit)
	fmt.Fprintf(&buf, "Usage:\n\t%s [options]\n\nOptions:\n\n", title)
	cli.PrintDefaults()
	usage = buf.String()
	cli.SetOutput(dlog)

	// do not show usage while parsing arguments
	cli.Usage = func() {}
	if err := cli.Parse(os.Args[1:]); err != nil {
		util.ShowOKMessage(util.MsgError, title, err.Error())
		os.Exit(1)
	}

	cli.Usage = func() {
		text := usage
		if len(socketName) > 0 {
			text += fmt.Sprintf("\nSocket path:\n  %s", socketName)
		}
		if len(pipeName) > 0 {
			text += fmt.Sprintf("\nPipe name:\n  %s", pipeName)
		}
		if lemon.IsSet() {
			text += fmt.Sprintf("\nLemonade stand:\n  %s", lemon.String())
		}
		util.ShowOKMessage(util.MsgInformation, title, text)
	}

	if help {
		cli.Usage()
		os.Exit(0)
	}

	// Check if Windows supports AF_UNIX sockets
	if ok, err := util.IsProperWindowsVer(); err != nil {
		util.ShowOKMessage(util.MsgError, title, err.Error())
		os.Exit(1)
	} else if !ok {
		util.ShowOKMessage(util.MsgError, title, "This Windows version does not support AF_UNIX sockets")
		os.Exit(1)
	}

	// Only allow single instance to run
	lockName := filepath.Join(os.TempDir(), title+".lock")
	inst, err := si.CreateLockFile(lockName)
	if err != nil {
		log.Print("Application already running")
		os.Exit(0)
	}
	defer func() {
		// Not necessary at all
		inst.Close()
		os.Remove(lockName)
	}()

	// Start lemonade backend if requested
	var lemonade *citrus.Citrus
	if lemon.IsSet() {
		log.Printf("Starting lemonade server on '%s'", lemon.String())
		if lemonade, err = citrus.NewCitrus(lemon, debug); err != nil {
			util.ShowOKMessage(util.MsgError, title, err.Error())
			os.Exit(1)
		} else {
			log.Print("Serving lemonade")
			go func() {
				lemonade.Serve(debug)
				// If for some reason process breaks - exit
				log.Print("Quiting - lemonade server ended")
				systray.Quit()
			}()
		}
	}

	// enter main processing loop
	if err := run(); err != nil {
		util.ShowOKMessage(util.MsgError, title, err.Error())
	}
}
