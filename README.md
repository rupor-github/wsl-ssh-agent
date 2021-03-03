<p align="center">
    <h1 align="center">wsl-ssh-agent</h1>
    <p align="center">
		Helper to interface with Windows ssh-agent.exe service from WSL, replacement for ssh-agent-wsl.
    </p>
    <p align="center">
        <a href="https://pkg.go.dev/mod/github.com/rupor-github/wsl-ssh-agent/?tab=packages"><img alt="GoDoc" src="https://img.shields.io/badge/godoc-reference-blue.svg" /></a>
        <a href="https://goreportcard.com/report/github.com/rupor-github/wsl-ssh-agent"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/rupor-github/wsl-ssh-agent" /></a>
    </p>
    <hr>
</p>

Windows 10 has very convenient `ssh-agent` service (with support for persistence and Windows security). Unfortunately it is not accessible from WSL. This project aims to correct this situation by enabling access to SSH keys held by Windows own `ssh-agent` service from inside the [Windows Subsystem for Linux](https://msdn.microsoft.com/en-us/commandline/wsl/about).

My first attempt - [ssh-agent-wsl](https://github.com/rupor-github/ssh-agent-wsl) was successful, but due to Windows interop restrictions it required elaborate life-time management on the WSL side. Starting with build 17063 (which was many updates ago) Windows implemented AF_UNIX sockets. This makes it possible to remove all trickery from WSL side greatly simplifying everything. 

**NOTE:** If you need access to more functionality (smard cards, identity management) provided by [GnuPG](https://www.gnupg.org/) set of tools on Windows - you may want to take a look at [win-gpg-agent](https://github.com/rupor-github/win-gpg-agent) instead.

`wsl-ssh-agent-gui.exe` is a simple "notification tray" applet which maintains AF_UNIX ssh-agent compatible socket on Windows end. It proxes all requests from this socket to ssh-agent.exe via named pipe. The only thing required on WSL end for it to work is to make sure that WSL `SSH_AGENT_SOCK` points to proper socket path. The same socket could be shared by any/all WSL sessions.

As an additional bonus `wsl-ssh-agent-gui.exe` could work as remote clipboard server so you could send your clipboard from tmux or neovim remote session back to your windows box over SSH secured connection easily. 

**NOTE: BREAKING CHANGE** Version 1.5.0 introduces breaking change. If you were not using `wsl-ssh-agent-gui.exe` as `lemonade` clipboard backend - this should not concern you at the slightest. Otherwise lemonade support no longer - it has been replaced with [gclpr](https://github.com/rupor-github/gclpr) which is more secure.


**SECURITY NOTICE:** All the usual security caveats applicable to WSL apply. Most importantly, all interaction with the Win32 world happens with the credentials of the user who started the WSL environment. In practice, *if you allow someone else to log in to your WSL environment remotely, they may be able to access the SSH keys stored in your ssh-agent.* This is a fundamental feature of WSL; if you are not sure of what you're doing, do not allow remote access to your WSL environment (i.e. by starting an SSH server).

**COMPATIBILITY NOTICE:** `wsl-ssh-agent-gui` was tested on Windows 10 1903 with multiple distributions and should work on anything
starting with 1809 - beginning with insider build 17063 and would not work on older versions of Windows 10, because it requires
[AF_UNIX socket support](https://devblogs.microsoft.com/commandline/af_unix-comes-to-windows/) feature.

## Installation

### From binaries

Download from the [releases page](https://github.com/rupor-github/wsl-ssh-agent/releases) and unpack it in a convenient location.

Starting with v1.5.1 releases are packed with zip and signed with [minisign](https://jedisct1.github.io/minisign/). Here is public key for verification: ![key](docs/build_key.png) RWTNh1aN8DrXq26YRmWO3bPBx4m8jBATGXt4Z96DF4OVSzdCBmoAU+Vq

## Usage

1. Ensure that on Windows side `ssh-agent.exe` service (OpenSSH Authentication Agent) is started and has your keys. (After adding keys to Windows `ssh-agent.exe` you may remove them from your wsl home .ssh directory - just do not forget to adjust `IdentitiesOnly` directive in your ssh config accordingly. Keys are securely persisted in Windows registry, available for your account only). You may also want to switch its startup mode to "automatic". Using powershell with elevated privileges (admin mode):
```
	Start-Service ssh-agent
	Set-Service -StartupType Automatic ssh-agent
```

2. Run `wsl-ssh-agent-gui.exe` with arguments which make sense for your usage. Basically there are several ways:

	* Using `-socket` option specify "well known" path on Windows side and then properly specify the same path in every WSL session:

		Windows:
		    ```
			wsl-ssh-agent-gui.exe -socket c:\wsl-ssh-agent\ssh-agent.sock
		    ```

		WSL:
		    ```
		    export SSH_AUTH_SOCK=/mnt/c/wsl-ssh-agent/ssh-agent.sock
		    ```

    * You could avoid any actions on WSL side by manually setting `SSH_AUTH_SOCK` and `WSLENV=SSH_AUTH_SOCK/up` on Windows side (**see note below**).

	* Using `-setenv` option allows application automatically modify user environment, so every WSL session started while
      `wsl-ssh-agent-gui.exe` is running will have proper `SSH_AUTH_SOCK` available to it (using `WSLENV`). By default socket
      path points to user temporary directory. Usual Windows user environment modification rules are applicable here (**see note below**).

**NOTE:** Setting SSH_AUTH_SOCK environment on Windows side may (and probably will) interfere with some of Windows OpenSSH. As far as I could see presently utilities in `Windows\System32\OpenSSH` expect this environment variable to be either empty or set to proper `ssh-agent.exe` pipe, otherwise they cannot read socket:

```
	if (getenv("SSH_AUTH_SOCK") == NULL)
		_putenv("SSH_AUTH_SOCK=\\\\.\\pipe\\openssh-ssh-agent");
```

To avoid this and still be able to use `-setenv` and automatically generated socket path use `-envname` to specify variable name to set. Later on WSL side you could use:
```
export SSH_AUTH_SOCK=${<<YOUR-NAME-HERE>>}
```

When `wsl-ssh-agent-gui.exe` is running you could see what it is connected to by clicking on its icon in notification tray area and selecting `About`. At the bottom of the message you would see something like:
```
Socket path:
  C:\Users\rupor\AppData\Local\Temp\ssh-273683143.sock
Pipe name:
  \\.\pipe\openssh-ssh-agent
Remote clipboard:
  gclpr is serving 2 key(s) on port 2850
```

For security reasons unless `-nolock` argument is specified program will refuse access to `ssh-agent.exe` pipe when user session is locked, so any long running background jobs in WSL which require ssh may fail.

## Options

Run `wsl-ssh-agent-gui.exe -help`

```
---------------------------
wsl-ssh-agent-gui
---------------------------

Helper to interface with Windows ssh-agent.exe service from WSL

Version:
	1.5.0 (go1.15.6)

Usage:
	wsl-ssh-agent-gui [options]

Options:

  -debug
    	Enable verbose debug logging
  -envname name
    	Environment variable name to hold socket path (default "SSH_AUTH_SOCK")
  -help
    	Show help
  -line-endings string
    	Remote clipboard convert line endings (LF/CRLF)
  -nolock
    	Provide access to ss-agent.exe even when user session is locked
  -pipe name
    	Pipe name used by Windows ssh-agent.exe
  -port int
    	Remote clipboard port (default 2850)
  -setenv
    	Export environment variable with 'envname' and modify WSLENV
  -socket path
    	Auth socket path (max 108 characters)
```

## WSL 2 compatibility

At the moment AF_UNIX interop does not seems to be working with WSL2 VMs. Hopefully this will be sorted out eventually. Meantime there is an easy workaround (proposed by multiple people) which does not use wsl-ssh-agent.exe at all and relies on combination of linux socat tool from your distribution and [npiperelay.exe](https://github.com/jstarks/npiperelay). Put npiperelay.exe somewhere on devfs for interop to work its magic (I have `winhome ⇒ /mnt/c/Users/rupor` in my $HOME directory for that) and add following lines in your .bashrc/.zshrc:

```
export SSH_AUTH_SOCK=$HOME/.ssh/agent.sock
ss -a | grep -q $SSH_AUTH_SOCK
if [ $? -ne 0   ]; then
    rm -f $SSH_AUTH_SOCK
    ( setsid socat UNIX-LISTEN:$SSH_AUTH_SOCK,fork EXEC:"$HOME/winhome/.wsl/npiperelay.exe -ei -s //./pipe/openssh-ssh-agent",nofork & ) >/dev/null 2>&1
fi
```
You *really* have to be on WSL 2 in order for this to work - if you see errors like `Cannot open netlink socket: Protocol not supported` - you probably are under WSL 1 and should not use this workaround. Run `wsl.exe -l -all -v` to check what is going on. When on WSL 2 make sure that socat is installed and npiperelay.exe is on windows partition and path is right. For convinience I will be packing pre-build npiperelay.exe with wsl-ssh-agent.

## Example

Putting it all together nicely - `remote` here refers to your wsl shell or some other box or virtual machine you could `ssh` to.

For my WSL installations I always create `~/winhome` and link it to my Windows home directory (where I have `.wsl` directory with various interoperability tools from Windows side). I am assuming that [gclpr](https://github.com/rupor-github/gclpr) is in your path on `remote` and you installed it's Windows counterpart somewhere in `drvfs` location (~/winhome/.wsl is a good place).

I auto-start `wsl-ssh-agent-gui.exe` on logon on my Windows box using following command line:
```
wsl-ssh-agent-gui.exe -setenv -envname=WSL_AUTH_SOCK
```
In my .bashrc I have:
```
[ -n ${WSL_AUTH_SOCK} ] && export SSH_AUTH_SOCK=${WSL_AUTH_SOCK}
```
and my `.ssh/config` entries used to `ssh` to `remote` have port forwarding enabled:
```
RemoteForward 2850 127.0.0.1:2850
```
On `remote` my `tmux.conf` includes following lines:
```
set -g set-clipboard off
if-shell 'if [ -n ${WSL_DISTRO_NAME} ]; then true; else false; fi' \
  'bind-key -T copy-mode-vi Enter send-keys -X copy-pipe-and-cancel "~/winhome/.wsl/gclpr.exe copy" ; bind-key -T copy-mode-vi MouseDragEnd1Pane send-keys -X copy-pipe-and-cancel "~/winhome/.wsl/gclpr.exe copy"' \
  'bind-key -T copy-mode-vi Enter send-keys -X copy-pipe-and-cancel "gclpr copy" ; bind-key -T copy-mode-vi MouseDragEnd1Pane send-keys -X copy-pipe-and-cancel "gclpr copy"'
```
And my `neovim` configuration file `init.vim` on `remote` has following lines:
```
set clipboard+=unnamedplus
if has("unix")
	" ----- on UNIX ask lemonade to translate line-endings
	if empty($WSL_DISTRO_NAME)
		if executable('gclpr')
			let g:clipboard = {
				\   'name': 'gclpr',
				\   'copy': {
				\      '+': 'gclpr copy',
				\      '*': 'gclpr copy',
				\    },
				\   'paste': {
				\      '+': 'gclpr paste --line-ending lf',
				\      '*': 'gclpr paste --line-ending lf',
				\   },
				\   'cache_enabled': 0,
				\ }
		endif
	else
		" ---- we are inside WSL - reach out to the Windows side
		if executable($HOME . '/winhome/.wsl/gclpr.exe')
			let g:clipboard = {
				\   'name': 'gclpr',
				\   'copy': {
				\      '+': $HOME . '/winhome/.wsl/gclpr.exe copy',
				\      '*': $HOME . '/winhome/.wsl/gclpr.exe copy',
				\    },
				\   'paste': {
				\      '+': $HOME . '/winhome/.wsl/gclpr.exe paste --line-ending lf',
				\      '*': $HOME . '/winhome/.wsl/gclpr.exe paste --line-ending lf',
				\   },
				\   'cache_enabled': 0,
				\ }
		endif
	endif
endif
```

Now you could open your WSL in terminal of your choice - mintty, cmd, Windows terminal, `ssh` to your `remote` using keys stored in Windows `ssh-agent.exe` without entering any additional passwords and have your clipboard content back on Windows transparently.

## Credit

* Thanks to [Ben Pye](https://github.com/benpye) with his [wsl-ssh-pageant](https://github.com/benpye/wsl-ssh-pageant) for inspiration.
* Thanks to [Masataka Pocke Kuwabara](https://github.com/pocke) for [lemonade](https://github.com/lemonade-command/lemonade) - a remote utility tool. (copy, paste and open browser) over TCP.
* Thanks to [John Starks](https://github.com/jstarks) for [npiperelay](https://github.com/jstarks/npiperelay) - access to Windows pipes from WSL.

------------------------------------------------------------------------------
Licensed under the GNU GPL version 3 or later, http://gnu.org/licenses/gpl.html

This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.
