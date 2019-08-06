#!/bin/sh

_red=$(printf "\033[1;91m")
_green=$(printf "\033[1;92m")
_yellow=$(printf "\033[1;93m")
_normal=$(printf "\033[0m")

run () {
    echo "> $_green$@$_normal"
    if $@; then
        echo "${_green}ok$_normal"
    else
        _ret=$?
        echo "${_red}failed$_normal"
        exit ${_ret}
    fi
}


echo "Expecting protoc version >= 3.6.1:"
protoc=$(which protoc)
$protoc --version

CURDIR=$(cd `dirname $0`; pwd)

if ! type protoc-gen-gogoslick >/dev/null 2>&1; then
    echo 'protoc-gen-gogoslick not found, run `go get github.com/gogo/protobuf/protoc-gen-gogoslick` to install' >&2
    exit 1
fi

gogoarg="plugins=grpc"

run protoc --gogoslick_out=${gogoarg}:${CURDIR}/bustsurvivor --proto_path=${CURDIR}/bustsurvivor bustsurvivor.proto