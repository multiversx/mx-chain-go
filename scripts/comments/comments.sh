# This script will add a dummy comment to all unexported methods / structs in a directory
# The gocmt tool has to be installed and placed in the root of the project near this file.
# Example:
# -> directory for comments.sh and gocmt binary: github.com/multiversx/mx-chain-go
# -> usage:  ./comments.sh process/mock  (will update if necessarily all the files from that directory)
packagePath=$1
for file in $packagePath/*
do
    if [[ -f $file ]]; then
       echo $file
       res=`./gocmt -t "-" $file`
       if ! [[ -z $res ]]; then
          echo "$res" > $file
       fi
    fi
done
