#!/bin/bash
for file in $@; do
    if [ ! -f $file ]; then
        exit 1
    fi
done
exit 0
