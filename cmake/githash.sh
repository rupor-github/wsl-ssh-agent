#!/bin/bash

if [ "${1}" = "" ]; then
    git="git"
else
    git=${1}
fi

sha=`${git} rev-list -1 HEAD | tr -d '\n'`
mod=`test -n "$(${git} diff --shortstat 2> /dev/null | tail -n1)" && echo -n "*" || true`

echo -n "${sha}${mod}"

