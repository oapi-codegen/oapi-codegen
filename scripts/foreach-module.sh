#!/usr/bin/env bash
# Run a make target in each child go module whose go directive is compatible
# with the current Go toolchain. Modules requiring a newer Go are skipped.
#
# Usage: foreach-module.sh <make-target>
set -euo pipefail

target="${1:?usage: foreach-module.sh <make-target>}"
cur_go="$(go env GOVERSION | sed 's/^go//')"

git ls-files '**/*go.mod' -z | while IFS= read -r -d '' modfile; do
	mod_go="$(sed -n 's/^go *//p' "$modfile")"
	moddir="$(dirname "$modfile")"

	if [ "$(printf '%s\n%s' "$mod_go" "$cur_go" | sort -V | head -1)" = "$mod_go" ]; then
		(set -x; cd "$moddir" && env GOBIN="${GOBIN:-}" make "$target")
	else
		echo "Skipping $moddir: requires go $mod_go, have go $cur_go"
	fi
done
