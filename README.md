# wsl-ssh-agent
--------------

## Replacement for [ssh-agent-wsl](https://github.com/rupor-github/ssh-agent-wsl).

Windows 10 has very convenient `ssh-agent` service (with support for persistence and Windows security). Unfortunately it is
not accessible from WSL. This project aims to correct this situation by enabling access to SSH keys held by Windows own
`ssh-agent` service from inside the [Windows Subsystem for Linux](https://msdn.microsoft.com/en-us/commandline/wsl/about).

My first attempt - [ssh-agent-wsl](https://github.com/rupor-github/ssh-agent-wsl) was successful, but due to Windows interop restrictions
it required elaborate life-time management on the WSL side. Starting with build 17063 (which was many updates ago) Windows implemented AF_UNIX sockets.
This makes it possible to remove all trickery from WSL side greatly simplifying everything.

`wsl-ssh-agent-gui.exe` is a simple "notification tray" applet which maintains AF_UNIX ssh-agent compatible socket on Windows
end. It proxes all requests from this socket to ssh-agent.exe via named pipe. The only thing required on WSL end for it to work
is to make sure that WSL `SSH_AGENT_SOCK` points to proper socket path. The same socket could be shared by any/all WSL sessions.

As an additional bonus `wsl-ssh-agent-gui.exe` could work as [lemonade](https://github.com/lemonade-command/lemonade) server so you could send your
clipboard from tmux or neovim remote session back to your windows box over SSH secured connection easily. Running `lemonade.exe` console
application in the background on Windows was always a bit tedious. Please, read lemonade documentation for details on how this works and parameters description.

**SECURITY NOTICE:** All the usual security caveats applicable to WSL apply.
Most importantly, all interaction with the Win32 world happens with the credentials of
the user who started the WSL environment. In practice, *if you allow someone else to
log in to your WSL environment remotely, they may be able to access the SSH keys stored in
your ssh-agent.* This is a fundamental feature of WSL; if you are not sure of what you're doing, do not allow remote access to
your WSL environment (i.e. by starting an SSH server).

**COMPATIBILITY NOTICE:** `wsl-ssh-agent-gui` was tested on Windows 10 1903 with multiple distributions and should work on anything
starting with 1809 - beginning with insider build 17063 and would not work on older versions of Windows 10, because it requires
[AF_UNIX socket support](https://devblogs.microsoft.com/commandline/af_unix-comes-to-windows/) feature.

## Installation

### From binaries

Download from the [releases page](https://github.com/rupor-github/wsl-ssh-agent/releases) and unpack it in a convenient location.

### From source

It is possible to build sources on Windows - after all it is written in _go_, however I am building under WSL, so you will need at least
go 1.13 installed under WSL along with regular build essentials and recent cmake:
	`sudo apt install build-essential cmake`
After that just execute
	`./build-release.sh`

## Usage

1. Ensure that on Windows side `ssh-agent` service (OpenSSH Authentication Agent) is started - you may want to switch its startup mode to "automatic". Using powershell with elevated privileges (admin mode):
```
	Start-Service ssh-agent
	Set-Service -StartupType Automatic ssh-agent
```
2. Run `wsl-ssh-agent-gui.exe`. Basically there are several possible scenarios:

	* Using `--socket` option specify "well known" path on Windows side and then properly specify the same path in every WSL session:

		Windows:
		    ```
			wsl-ssh-agent-gui.exe -socket c:\wsl-ssh-agent\ssh-agent.sock
		    ```

		WSL:
		    ```
		    export SSH_AUTH_SOCK=/mnt/c/wsl-ssh-agent/ssh-agent.sock
		    ```

    * You could avoid any actions on WSL side by manually setting `SSH_AUTH_SOCK` and `WSLENV=SSH_AUTH_SOCK/up` on Windows side.

	* Using `--setenv` option allows application automatically modify user environment, so every WSL session started while
      `wsl-ssh-agent-gui.exe` is running will have proper `SSH_AUTH_SOCKET` available to it (using `WSLENV`). By default socket
      path points to user temporary directory. Usual Windows user environment modification rules are applicable here.

## Options

Run `wsl-ssh-agent-gui.exe -help`

	Helper to interface with Windows ssh-agent.exe service from WSL

	Version:
		1.1 (go1.13)
		<git sha hash>

	Usage:
		wsl-ssh-agent-gui [options]

	Options:

		-debug
			Enable verbose debug logging
		-help
			Show help
		-lemonade list
			semicolon separated list of lemonade "server" options (TCP port, Allow IP Range, Line Endings)
		-pipe name
			Pipe name used by Windows ssh-agent.exe
		-setenv
			Export SSH_AUTH_SOCK and modify WSLENV
		-socket path
			Auth socket path (max 108 characters)

## Example

Putting it all together nicely - `remote` here refers to your wsl shell or some other box or virtual machine you could `ssh` to. Assuming that [lemonade](https://github.com/lemonade-command/lemonade) is in your path on `remote` and you installed [win32yank](https://github.com/equalsraf/win32yank) somewhere in `drvfs` location (and link it in your path `/usr/local/bin/win32yank ⇒ /mnt/c/wslhome/win32yank.exe`).

I auto-start `wsl-ssh-agent-gui.exe` on logon on my Windows box using following command line:
```
	wsl-ssh-agent-gui.exe --setenv --lemonade=2489;127.0.0.1/24
```
and my `.ssh/config` entries used to `ssh` to `remote` have port forwarding enabled:
```
	RemoteForward 2489 127.0.0.1:2489
```
On `remote` my `tmux.conf` includes following lines:
```
	set -g set-clipboard off
	if-shell 'if [ $(uname -a | grep -c Microsoft) = 1 ]; then true; else false; fi' \
		'bind-key -T copy-mode-vi Enter send-keys -X copy-pipe-and-cancel "win32yank -i --crlf" ; bind-key -T copy-mode-vi MouseDragEnd1Pane send-keys -X copy-pipe-and-cancel "win32yank -i --crlf"' \
		'bind-key -T copy-mode-vi Enter send-keys -X copy-pipe-and-cancel "~/.local/bin/lemonade copy" ; bind-key -T copy-mode-vi MouseDragEnd1Pane send-keys -X copy-pipe-and-cancel "~/.local/bin/lemonade copy"'
```
And my `neovim` configuration `init.vim` has following:
```
    let s:u_wsl = 1
    let _ = system("uname -a | grep -cq Microsoft")
    if v:shell_error
        let s:u_wsl = 0
    endif

	" ----- Clipboard
	set clipboard+=unnamedplus
	if s:u_wsl
		" ----- on WSL try to use win32yank if available
		" ----- we are relying on WSLENV being set to "USERPROFILE/up" outside of WSL
		let s:win32yank = $USERPROFILE . '/.wsl/win32yank.exe'
		if filereadable(s:win32yank)
			let g:clipboard = {
				\   'name': 'win32yank',
				\   'copy': {
				\      '+': s:win32yank . ' -i --crlf',
				\      '*': s:win32yank . ' -i --crlf',
				\    },
				\   'paste': {
				\      '+': s:win32yank . ' -o --lf',
				\      '*': s:win32yank . ' -o --lf',
				\   },
				\   'cache_enabled': 0,
				\ }
		endif
	elseif has("unix")
		" ----- on UNIX ask lemonade to translate line-endings
		if executable('lemonade')
			let g:clipboard = {
				\   'name': 'lemonade',
				\   'copy': {
				\      '+': 'lemonade copy',
				\      '*': 'lemonade copy',
				\    },
				\   'paste': {
				\      '+': 'lemonade paste --line-ending lf',
				\      '*': 'lemonade paste --line-ending lf',
				\   },
				\   'cache_enabled': 0,
				\ }
		endif
	endif
```

## Credit

* Thanks to [Ben Pye](https://github.com/benpye) with his [wsl-ssh-pageant](https://github.com/benpye/wsl-ssh-pageant) for inspiration.
* Thanks to [Masataka Pocke Kuwabara](https://github.com/pocke) for [lemonade](https://github.com/lemonade-command/lemonade) - a remote utility tool. (copy, paste and open browser) over TCP.

------------------------------------------------------------------------------
Licensed under the GNU GPL version 3 or later, http://gnu.org/licenses/gpl.html

This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.

See the `COPYING` file for license details.
