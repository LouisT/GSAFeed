#!/usr/bin/env bash

# List of server IDs + paths
declare -A SERVERS=(
  [GS1]="/home/username/geneshift/Geneshift01"
  [GS2]="/home/username/geneshift/Geneshift02"
  [GST]="/home/username/geneshift/GSTesting"
)
export SERVERS

# Geneshift username
export HOST=Foo_Bar

# Firejail arguments
export JAILARGS="--dns=8.8.8.8" # Use default network settings (set DNS server)
# export JAILARGS="--dns=8.8.8.8 --net=eth0 --ip=127.0.0.1" # Set sandbox IP + DNS server

# Name of the jail for access with `firejail` cli
#  firejail --shutdown=$JAIL
export JAIL=Geneshift

# Update URL (binary of GeneshiftServer)
export UPDATE=https://geneshift.net/game/Geneshift/GeneshiftServer