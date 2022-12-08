#!/usr/bin/env bash

# List of server IDs + paths
declare -A SERVERS=(
  [GS1]="/home/username/GSA/01"
  [GS2]="/home/username/GSA/02"
  [GST]="/home/username/GSA/GST"
)
export SERVERS

# GSA args
export GSAARGS="-h Foo_Bar -v 3"

# Firejail arguments
export JAILARGS=""

# Name of the jail for access with `firejail` cli
#  firejail --shutdown=JAIL
export JAIL=GSA

# Update URL (binary of GeneShiftAutoServer)
export UPDATE=https://geneshiftauto.com/game/GeneShiftAuto/GeneShiftAutoServer