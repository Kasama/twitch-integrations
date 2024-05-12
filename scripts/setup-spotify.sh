#!/bin/bash

DEFAULT_OUTPUT=$(pactl info|sed -n -e 's/^.*Default Sink: //p')
sink_name=spotify
pactl load-module module-null-sink sink_name="$sink_name"
pactl load-module module-loopback source="$sink_name.monitor" sink="$DEFAULT_OUTPUT"
pactl load-module module-remap-source master="$sink_name.monitor" source_name=discord-shared source_properties=device.description="$sink_name"
