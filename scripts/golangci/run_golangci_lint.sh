#!/bin/sh

if ! [ -x "$(command -v golangci-lint)" ]; then
  echo Error: golangci-lint is not installed.
  exit 1
fi

input="linters.txt"

dir_output="Output"
if [ -d "$dir_output" ]; then
	rm -r "$dir_output"
fi

mkdir "$dir_output"

while IFS= read -r linter
do
  echo -n Running "$linter" "linter "
	golangci-lint run ./../../... > "$dir_output/${linter}_output.txt" --max-issues-per-linter 0 --max-same-issues 0 --disable-all --enable="$linter"
	echo -- Done
done < "$input"
