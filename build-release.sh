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

cp ../npiperelay/npiperelay.exe ${_dist}/.
cd ${_dist}
zip -9 ../wsl-gpg-agent.zip *
cd ..
echo ${BUILD_PSWD} | minisign -S -s ~/.minisign/build.key -c "wsl-gpg-agent release signature" -m wsl-gpg-agent.zip
