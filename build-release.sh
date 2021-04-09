#!/bin/bash

_dist=bin
[ -d ${_dist} ] && rm -rf ${_dist}
(
	[ -d release ] && rm -rf release
	mkdir release
	cd release

    cmake -DCMAKE_BUILD_TYPE=Release ..
	make install
)

cp ../../jstarks/npiperelay/npiperelay.exe ${_dist}/.
cd ${_dist}
zip -9 ../wsl-ssh-agent.zip *
cd ..
echo ${BUILD_PSWD} | minisign -S -s ~/.minisign/build.key -c "wsl-ssh-agent release signature" -m wsl-ssh-agent.zip
