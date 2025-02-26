#!/usr/bin/env sh

##
## Input parameters
##
BINARY=/pelldvs/${BINARY:-pelldvs}
ID=${ID:-0}
LOG=${LOG:-pelldvs.log}

##
## Assert linux binary
##
if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found. Please add the binary to the shared folder. Please use the BINARY environment variable if the name of the binary is not 'pelldvs' E.g.: -e BINARY=my_test_binary"

	exit 1
fi
BINARY_CHECK="$(file "$BINARY" | grep 'ELF 64-bit LSB executable, x86-64')"
if [ -z "${BINARY_CHECK}" ]; then
	echo "Binary needs to be OS linux, ARCH amd64 (build with 'make build-linux')"
	exit 1
fi

##
## Run binary with all parameters
##
export PELLDVSHOME="/pelldvs/node${ID}"

if [ -d "`dirname ${PELLDVSHOME}/${LOG}`" ]; then
  "$BINARY" "$@" | tee "${PELLDVSHOME}/${LOG}"
else
  "$BINARY" "$@"
fi

chmod 777 -R /pelldvs

