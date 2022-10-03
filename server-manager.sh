#!/usr/bin/env bash
# Manage Geneshift servers using firejail.
# This enables the ability to choose a specific IP address,
# view process trees, see network and other stats via firejail.
#
#  Example: firejail --netstats

# TODO:
#   Allow starting/stopping/restarting a specific server (./server-manager.sh --start ID)
#   Figure out more firejail related features like disabling root.
#   Many other things I can't even think about right now!?

declare -A SERVERS=()
HOST="UNSET"
JAILARGS=""
# XXX: No need to change this unless managing
#      multiple split servers for some reason?
JAIL=Geneshift

# Read/orverwrite settings?
if [ -f server-config.sh ]; then
    # shellcheck source=/dev/null
    . ./server-config.sh
fi

# Make sure the following vars aren't overwritten in server-config.sh
POSITIONAL_ARGS=()
QUIETMODE=false

exec 3>&1
function log() {
    if [ "$QUIETMODE" = false ]; then
        if $2; then
            echo -e ">>> $1" 1>&3
        else
            echo -e "$1" 1>&3
        fi
    fi
}

function ValidateServers() {
    if [ ${#SERVERS[@]} == 0 ]; then
        log "No servers provided; please check the server-config.sh file!"
        exit 1
    fi
}

function Jailer() {
    if firejail --list | grep "join-or-start=$JAIL" >/dev/null 2>&1; then
        log "Jailer ($JAIL) exists!"
    else
        log "Starting jailer ($JAIL)..."
        (exec nohup firejail --name=jailer "$JAILARGS" --join-or-start=$JAIL sleep inf) </dev/null &>/dev/null &
    fi
}

function JailerProcs() {
    log "Jailer ($JAIL) processes:"
    log "$(firejail --list | grep "join-or-start=$JAIL")" false
}

function StartServers() {
    ValidateServers
    log "Starting Geneshift servers; please wait..."
    for KEY in "${!SERVERS[@]}"; do
        VAL=${SERVERS[$KEY]}
        if firejail --list | grep "name=$KEY" >/dev/null 2>&1; then
            log "Server for $KEY is already running!" false
        else
            log "Starting $KEY ($VAL) in 5 seconds..." false
            sleep 5
            (exec nohup firejail --private-cwd="$VAL" --name="$KEY" --noprofile "$JAILARGS" --join-or-start=$JAIL -- ./GeneshiftServer -h $HOST) </dev/null &>"$VAL/stdout.txt" &
        fi
    done
}

function StopServers() {
    log "Stopping Geneshift jail; please wait..."
    firejail --shutdown=$JAIL
}

function UpdateGeneshift() {
    ValidateServers
    for KEY in "${!SERVERS[@]}"; do
        if firejail --list | grep "name=$KEY" >/dev/null 2>&1; then
            log "Please stop Geneshift servers first!"
            JailerProcs
            exit 1
        fi
    done
    log "Updating Geneshift binary; please wait..."
    tmpfile=$(mktemp /tmp/Geneshift.XXXXXX)
    if ! wget "$UPDATE" -O "$tmpfile" -q --show-progress; then
        log "Failed to download $UPDATE to $tmpfile!" false
        exit 1
    fi
    for KEY in "${!SERVERS[@]}"; do
        VAL=${SERVERS[$KEY]}
        OUT="$VAL/GeneshiftServer"
        log "Copying $tmpfile to $OUT..." false
        cp -rf "$tmpfile" "$OUT"
        chmod +x "$OUT"
    done
    log "Finished updating binaries!"
}

function Help() {
    log "TODO: Help"
}

# Parse these args FIRST
while [[ $# -gt 0 ]]; do
    case $1 in
    -q | --quiet)
        QUIETMODE=true
        shift
        ;;
    *)
        POSITIONAL_ARGS+=("$1")
        shift
        ;;
    esac
done
set -- "${POSITIONAL_ARGS[@]}" # restore positional parameters

# Parse secondary args
while [[ $# -gt 0 ]]; do
    case $1 in
    --update)
        UpdateGeneshift
        shift
        ;;
    --start)
        Jailer
        StartServers
        shift
        ;;
    --stop)
        StopServers
        shift
        ;;
    --restart)
        StopServers
        Jailer
        StartServers
        shift
        ;;
    -h | --help)
        Help
        shift
        ;;
    --* | -*)
        log "Unknown option $1"
        exit 1
        ;;
    *)
        POSITIONAL_ARGS+=("$1")
        shift
        ;;
    esac
done
set -- "${POSITIONAL_ARGS[@]}" # restore positional parameters
