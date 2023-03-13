#!/bin/bash

if [ "${1}" = "win32" ] || [ "${1}" = "win64" ]; then

    hash=`sha256sum -z wsl-ssh-agent.zip | awk '{ print $1; }'`
    sed -i -e "s/__CURRENT_HASH_${1}__/${hash}/g" wsl-ssh-agent.json

fi
