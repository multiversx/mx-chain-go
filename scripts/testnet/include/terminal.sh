# Determine which terminal emulator is available 
# (currently, one of "konsole", "gnome-terminal" or "none").
# TMUX support is in development.
TERMWRAPPER="none"

if [ -n "$(command -v "konsole")" ]
then
  export TERMWRAPPER="konsole"
elif [ -n "$(command -v "gnome-terminal")" ]
then
  export TERMWRAPPER="gnome-terminal"
fi

export CURRENT_TMUX_SESSION=""
export CURRENT_TMUX_LAYOUT="tiled"
declare -A TMUX_SESSION_PANES

setTerminalSession() {
  if [ $USETMUX -eq 1 ]
  then
    local session_name=$1
    if [ -z "$session_name" ]
    then
      session_name=$CURRENT_TMUX_SESSION
    fi

    if [ "$(tmuxSessionAlreadyExists $session_name)" -eq 0 ]
    then
      CURRENT_TMUX_SESSION=$session_name
      tmux new-session -d -s $session_name
      TMUX_SESSION_PANES[$session_name]=1
    else
      CURRENT_TMUX_SESSION=$session_name
    fi
  fi
}

setTerminalLayout() {
  CURRENT_TMUX_LAYOUT=$1
}

showTerminalSession() {
  local session_name=$1
  local keepopen=$2
  if [ $USETMUX -eq 1 ]
  then
    if [ $TERMWRAPPER == "none" ]
    then
      echo "No terminal emulator found. The tmux session will continue to run in background, and can be attached to manually."
    else
      executeCommandInTerminalEmulator "tmux attach-session -t $session_name" $keepopen
    fi
  fi
}

setWorkdirForNextCommands() {
  CURRENT_COMMAND_WORKDIR=$1
}

runCommandInTerminal() {
  local command_to_run=$1
  local keepopen=$2

  if [ $USETMUX -eq 1 ]
  then
    local pane_to_use=${TMUX_SESSION_PANES[$CURRENT_TMUX_SESSION]}
    if [ $pane_to_use -gt 1 ]
    then
      tmux split-window -t $CURRENT_TMUX_SESSION:0
      tmux select-layout -t $CURRENT_TMUX_SESSION:0 $CURRENT_TMUX_LAYOUT
      let TMUX_SESSION_PANES[$CURRENT_TMUX_SESSION]=${TMUX_SESSION_PANES[$CURRENT_TMUX_SESSION]}+1
    fi
    
    let pane_id=$pane_to_use-1
    tmux send-keys -t $CURRENT_TMUX_SESSION:0.$pane_id "cd $CURRENT_COMMAND_WORKDIR" C-m
    tmux send-keys -t $CURRENT_TMUX_SESSION:0.$pane_id "$command_to_run" C-m

    if [ $pane_to_use -eq 1 ]
    then
      let TMUX_SESSION_PANES[$CURRENT_TMUX_SESSION]=${TMUX_SESSION_PANES[$CURRENT_TMUX_SESSION]}+1
    fi
  else
    executeCommandInTerminalEmulator "$command_to_run" $keepopen
  fi
}

executeCommandInTerminalEmulator() {
  local command_to_run=$1
  local keepopen=$2

  cd $CURRENT_COMMAND_WORKDIR

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

  if [ $TERMWRAPPER == "none" ]
  then
    echo "No terminal emulator found, command could not be run."
  fi

}

stopProcessByPort() {
  sendSignalToProcessByPort $1 2
}

pauseProcessByPort() {
  sendSignalToProcessByPort $1 19
}

resumeProcessByPort() {
  sendSignalToProcessByPort $1 18
}

sendSignalToProcessByPort() {
  local port=$1
  local pid=$(lsof -t -i:$port)
  local signal=$2

  if [ -n "$pid" ]
  then
    echo "Sending signal $signal to process $pid..."
    kill -n $signal $pid
  fi
}

tmuxSessionAlreadyExists() {
  local session_name=$1

  tmux has-session -t $session_name 2>/dev/null

  if [ $? != 0 ]
  then
    echo 0
  else
    echo 1
  fi
}
