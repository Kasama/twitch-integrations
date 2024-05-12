#!/bin/bash

# Specifying the icon(s) in the script
# This allows us to change its appearance conditionally
music_icon=""
muted_icon=""

ignoredPlayers="firefox,chromium"

#loop forever
while true; do
  status=$(playerctl -i "$ignoredPlayers" status 2> /dev/null)
  if [[ "$status" = "Playing" ]]; then
    metadata="$music_icon $(playerctl -i "$ignoredPlayers" metadata artist 2> /dev/null) - $(playerctl -i "$ignoredPlayers" metadata title 2> /dev/null)        "
  else
    metadata=""
  fi

  echo "$metadata" > /tmp/music-playing.txt

  default_mic=$(pactl get-default-source)
  mic_muted=$(pactl get-source-mute "$default_mic" | awk '{print $2}')

  if [ "$mic_muted" = "yes" ]; then
    muted_status="$muted_icon"
  else
    muted_status=""
  fi
  echo "$muted_status" > /tmp/mic-muted.txt

  ######################### Microphone status

  default_mic=$(pactl get-default-source)
  mic_muted=$(pactl get-source-mute "$default_mic" | awk '{print $2}')

  if [ "$mic_muted" = "yes" ]; then
    echo "${muted_icon}" > /tmp/mic-muted.txt
  else
    echo "" > /tmp/mic-muted.txt
  fi

  sleep 1
done
