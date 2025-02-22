#!/bin/sh

inputdatad=/usr/share/dict
inputdataf=words
inputdata="${inputdatad}/${inputdataf}"

inputidxd=./sample.d/idx.d

indexfile=./sample.d/idx.d/words.dat

genindex(){
	echo generating index...

	export ENV_INPUT_DIR_NAME="${inputdatad}"
	export ENV_OUTPUT_DIR_NAME="${inputidxd}"

	mkdir -p sample.d/idx.d

	printf '%s\n' \
		words |
		./text2suffix
}

test -f "${indexfile}" || genindex

export ENV_INDEX_FILE_NAME="${indexfile}"

printf \
	'%s\n' \
	hello \
	world |
	\time -l ./suffix2needles2count

\time fgrep --count hello "${inputdata}"
\time fgrep --count world "${inputdata}"
