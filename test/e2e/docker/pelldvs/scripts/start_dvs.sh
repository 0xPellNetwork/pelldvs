set -e

logt() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') $1"
}

function emulator_healthcheck {
  set +e
  while true; do
    ssh emulator "test -f /root/emulator_initialized"
    if [ $? -eq 0 ]; then
      echo "Emulator initialized, proceeding to the next step..."
      break
    fi
    echo "Emulator not initialized, retrying in 2 second..."
    sleep 2
  done
  ## Wait for emulator to be ready
  ## TODO: remove this once we have a better healthcheck
  sleep 8
  set -e
}

emulator_healthcheck

# start sshd
/usr/sbin/sshd

if [ ! -f /root/dvs_initialized ]; then
  source "$(dirname "$0")/init_dvs.sh"
  touch /root/dvs_initialized
fi

source "$(dirname "$0")/start_aggregator.sh"

