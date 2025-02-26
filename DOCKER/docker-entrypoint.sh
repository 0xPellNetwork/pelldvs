#!/bin/bash
set -e

if [ ! -d "$PELLDVSHOME/config" ]; then
	echo "Running pelldvs init to create (default) configuration for docker run."
	pelldvs init

	sed -i \
		-e "s/^proxy_app\s*=.*/proxy_app = \"$PROXY_APP\"/" \
		-e "s/^moniker\s*=.*/moniker = \"$MONIKER\"/" \
		-e 's/^addr_book_strict\s*=.*/addr_book_strict = false/' \
		-e 's/^timeout_commit\s*=.*/timeout_commit = "500ms"/' \
		-e 's/^index_all_tags\s*=.*/index_all_tags = true/' \
		-e 's,^laddr = "tcp://127.0.0.1:26657",laddr = "tcp://0.0.0.0:26657",' \
		-e 's/^prometheus\s*=.*/prometheus = true/' \
		"$PELLDVSHOME/config/config.toml"

fi

exec pelldvs "$@"
