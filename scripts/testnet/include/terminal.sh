source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"

# Determine which terminal emulator is available 
# (currently, one of "konsole" or "gnome-terminal").
# TMUX support is in development.
if [ -n $(command -v "konsole") ]
then
  export TERMWRAPPER="konsole"
elif [ -n $(command -v "gnome-terminal") ]
then
  export TERMWRAPPER="gnome-terminal"
fi


TERMINAL_COMMAND_BUFFER=""

startTerminal() {
  TERMINAL_COMMAND_BUFFER=""
}

endTerminal() {
  echo -n ""
}

runCommandInTerminal() {
  local command_to_run=$1
  local keepopen=$2

  if [ $TERMWRAPPER == "konsole" ]
  then
    if [ -z "$keepopen" ]
    then
      konsole -e $command_to_run &
    else
      konsole --noclose -e $command_to_run &
    fi
  fi

  if [ $TERMWRAPPER == "gnome-terminal" ]
  then
    gnome-terminal -- $command_to_run &
  fi
}

stopProcessByPort() {
  local port=$1
  local pid=$(lsof -t -i:$port)
  local signal=2  # SIGINT, identical to CTRL+C

  if [ -n "$pid" ]
  then
    kill -n $signal $pid
  fi
}

