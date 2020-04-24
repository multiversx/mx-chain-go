#!/usr/bin/env bash

export ELRONDTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

$ELRONDTESTNETSCRIPTSDIR/stop.sh
$ELRONDTESTNETSCRIPTSDIR/clean.sh
$ELRONDTESTNETSCRIPTSDIR/config.sh
$ELRONDTESTNETSCRIPTSDIR/start.sh $1
