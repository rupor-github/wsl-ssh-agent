#!/bin/bash
set -e
passage show build | minisign -S -s ${HOME}/.minisign/build.key -c "wsl-ssh-agent release signature" -m wsl-ssh-agent.zip
sed -i "s/__CURRENT_HASH__/$(sha256sum -z wsl-ssh-agent.zip | awk '{ print $1; }')/g" wsl-ssh-agent.json
