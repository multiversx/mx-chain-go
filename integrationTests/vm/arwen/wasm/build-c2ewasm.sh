#!/usr/bin/env bash

HELP="USAGE: Some help here."
CLANG="clang"
LLC="llc"
WASMLD="wasm-ld"
PYTHON3="python3"
TOOLS="$CLANG $LLD $WASMLD $PYTHON3 $EWASMIFY $WASMDIS $WAT2WASM"

# Make sure we have the first argument, i.e. the name of the file to be
# compiled.
if [ -z $1 ]
then
  echo "Argument required."
  echo ""
  echo $HELP
  exit 1
fi
SOURCE_C=$1

# Take the filename without extension and set it to the SOURCE variable -
# needed to create intermediary files, which have various extensions.
SOURCE="${SOURCE_C%%.*}"

# If a second argument was provided, use it as SOURCE_SYMS. Otherwise, use the
# default "main.syms".
SOURCE_SYMS="main.syms"
if [ ! -z $2 ]
then
    $SOURCE_SYMS = $2
fi

# We need a C file and a .syms file in order to compile and link successfully.
echo "Looking for files $SOURCE_C and $SOURCE_SYMS..."
if [ ! -f $SOURCE_C ]
then
  echo "File $SOURCE_C not found."
  exit 1
fi

if [ ! -f $SOURCE_SYMS ]
then
  echo "File $SOURCE_SYMS not found."
  exit 1
fi

# Verify if the required tools are available as commands.
for TOOL in $TOOLS; do
  if [ ! -f $TOOL ]
  then
    if [ ! -x `command -v $TOOL` ]
    then
      echo "$TOOL is required, but was not found."
      exit 1
    fi
  fi
done

# Assemble exports
SOURCE_EXPORT=$SOURCE".export"
EXPORTS=""
while read exp; do
	EXPORTS=$EXPORTS" -export=$exp"
done < $SOURCE_EXPORT

# Now perform the actual building.
$CLANG -cc1 -Ofast -emit-llvm -triple=wasm32-unknown-unknown-wasm $SOURCE_C

SOURCE_LL=$SOURCE".ll"
SOURCE_O=$SOURCE".o"
SOURCE_WASM=$SOURCE".wasm"
$LLC -O3 -filetype=obj $SOURCE_LL -o $SOURCE_O
$WASMLD --no-entry $SOURCE_O -o $SOURCE_WASM --strip-all -allow-undefined-file=$SOURCE_SYMS $EXPORTS
