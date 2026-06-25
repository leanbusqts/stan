#!/usr/bin/env sh
set -eu

usage() {
  cat <<'USAGE'
Usage: ./install.sh [--bin-dir DIR] [--prefix DIR] [--dry-run]

Builds stan and installs it as a global command.

Options:
  --bin-dir DIR  Install directly into DIR
  --prefix DIR   Install into DIR/bin
  --dry-run      Print the planned install path without writing
  -h, --help     Show this help
USAGE
}

die() {
  printf '%s\n' "error: $*" >&2
  exit 1
}

info() {
  printf '%s\n' "$*"
}

resolve_repo_dir() {
  script=$0
  case $script in
    */*) script_dir=${script%/*} ;;
    *) script_dir=. ;;
  esac
  cd "$script_dir" && pwd
}

bin_dir=${STAN_INSTALL_BIN_DIR:-}
prefix=${STAN_INSTALL_PREFIX:-}
dry_run=0

while [ "$#" -gt 0 ]; do
  case "$1" in
    --bin-dir)
      [ "$#" -ge 2 ] || die "--bin-dir requires a value"
      bin_dir=$2
      shift 2
      ;;
    --prefix)
      [ "$#" -ge 2 ] || die "--prefix requires a value"
      prefix=$2
      shift 2
      ;;
    --dry-run)
      dry_run=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown argument: $1"
      ;;
  esac
done

if [ -n "$bin_dir" ] && [ -n "$prefix" ]; then
  die "use either --bin-dir or --prefix, not both"
fi

if [ -z "$bin_dir" ]; then
  if [ -n "$prefix" ]; then
    bin_dir=$prefix/bin
  else
    bin_dir=$HOME/bin
  fi
fi

repo_dir=$(resolve_repo_dir)
install_path=$bin_dir/stan

info "Stan installer"
info "Repo: $repo_dir"
info "Install path: $install_path"

if [ "$dry_run" -eq 1 ]; then
  exit 0
fi

command -v go >/dev/null 2>&1 || die "go is required but was not found in PATH"

mkdir -p "$bin_dir"
tmp_dir=${TMPDIR:-/tmp}
tmp_bin=$tmp_dir/stan-install-$$
trap 'rm -f "$tmp_bin"' EXIT HUP INT TERM

(cd "$repo_dir" && go build -o "$tmp_bin" .)
mv "$tmp_bin" "$install_path"
chmod 755 "$install_path"

info "Installed: $install_path"

case ":$PATH:" in
  *":$bin_dir:"*) ;;
  *)
    info ""
    info "Warning: $bin_dir is not in PATH."
    info "Add this to your shell profile:"
    info "  export PATH=\"$bin_dir:\$PATH\""
    ;;
esac

info ""
"$install_path" version

credentials=$HOME/.config/stan/client_secret.json
if [ ! -f "$credentials" ] && [ ! -f "$repo_dir/client_secret.json" ]; then
  info ""
  info "Google OAuth credentials are still required."
  info "1. Open Google Cloud Console."
  info "2. Enable Google Calendar API and Google Tasks API."
  info "3. Create OAuth client credentials of type Desktop app."
  info "4. Download the JSON and install it with:"
  info "   mkdir -p ~/.config/stan"
  info "   mv ~/Downloads/client_secret*.json ~/.config/stan/client_secret.json"
fi

info ""
info "Next:"
info "  stan doctor"
info "  stan auth login"
