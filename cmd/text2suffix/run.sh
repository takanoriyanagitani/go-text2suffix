#!/bin/sh

idir=/usr/share/dict

export ENV_INPUT_DIR_NAME="${idir}"

mkdir -p ./sample.d/out.d

export ENV_OUTPUT_DIR_NAME=./sample.d/out.d

printf '%s\n' \
	words |
	\time -l ./text2suffix

ls -lShL \
	"${ENV_INPUT_DIR_NAME}/words" \
	"${ENV_OUTPUT_DIR_NAME}/words.dat"
