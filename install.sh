#!/bin/sh
# scout installer - detects os/arch, downloads the matching release binary,
# verifies its checksum, and installs it. re-run to upgrade to the latest release.
#
# usage:
#   curl -fsSL https://raw.githubusercontent.com/mirageglobe/scout/main/install.sh | sh
#
# environment overrides:
#   SCOUT_VERSION      install a specific version (e.g. 0.8.0); default: latest release
#   SCOUT_BIN_DIR      install directory; default: ~/.local/bin (falls back to /usr/local/bin)
#   SCOUT_SKIP_VERIFY  set to 1 to install without verifying the checksum; off by default,
#                      and the only way to proceed when the checksum cannot be confirmed

set -eu

REPO="mirageglobe/scout"
BINARY="scout"

# --- helpers ---------------------------------------------------------------

info() { printf '[ scout ] %s\n' "$1"; }
fail() { printf '[ fail  ] %s\n' "$1" >&2; exit 1; }

need() { command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"; }

# --- detect download tool --------------------------------------------------

if command -v curl >/dev/null 2>&1; then
  dl() { curl -fsSL "$1" -o "$2"; }
  dl_stdout() { curl -fsSL "$1"; }
elif command -v wget >/dev/null 2>&1; then
  dl() { wget -qO "$2" "$1"; }
  dl_stdout() { wget -qO - "$1"; }
else
  fail "need curl or wget to download"
fi

need tar
need uname

# --- detect os/arch --------------------------------------------------------

os=$(uname -s)
case "$os" in
  Darwin) os="darwin" ;;
  Linux)  os="linux" ;;
  *) fail "unsupported os: $os (scout ships darwin and linux builds)" ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64 | amd64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  *) fail "unsupported arch: $arch" ;;
esac

# linux/arm64 is not built (see .goreleaser.yaml)
if [ "$os" = "linux" ] && [ "$arch" = "arm64" ]; then
  fail "no linux/arm64 build available; build from source: https://github.com/$REPO"
fi

# --- resolve version -------------------------------------------------------

version="${SCOUT_VERSION:-}"
if [ -z "$version" ]; then
  info "resolving latest release..."
  tag=$(dl_stdout "https://api.github.com/repos/$REPO/releases/latest" \
    | grep '"tag_name"' | head -n1 | cut -d'"' -f4)
  [ -n "$tag" ] || fail "could not resolve latest release tag"
else
  # accept both "0.8.0" and "v0.8.0"
  case "$version" in v*) tag="$version" ;; *) tag="v$version" ;; esac
fi

# goreleaser strips the leading v for artifact names
ver_no_v=$(printf '%s' "$tag" | sed 's/^v//')

archive="${BINARY}_${ver_no_v}_${os}_${arch}.tar.gz"
checksums="${BINARY}_${ver_no_v}_checksums.txt"
base="https://github.com/$REPO/releases/download/$tag"

# --- download & verify -----------------------------------------------------

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

info "downloading $archive ($tag)..."
dl "$base/$archive" "$tmp/$archive" || fail "download failed: $base/$archive"

# verification is fail-closed: anything that stops us confirming the checksum aborts
# the install. this script is meant to be piped into sh, where an advisory "skipping
# verification" line scrolls past unread, so skipping has to be opted into by hand.
if [ "${SCOUT_SKIP_VERIFY:-0}" = "1" ]; then
  info "SCOUT_SKIP_VERIFY=1 set; installing WITHOUT checksum verification"
else
  info "verifying checksum..."

  # pick a sha256 tool up front, same shape as the dl() selection above
  if command -v sha256sum >/dev/null 2>&1; then
    sha256() { sha256sum "$1"; }
  elif command -v shasum >/dev/null 2>&1; then
    sha256() { shasum -a 256 "$1"; }
  else
    fail "no sha256 tool found (need sha256sum or shasum); install one, or re-run with SCOUT_SKIP_VERIFY=1 to bypass"
  fi

  dl "$base/$checksums" "$tmp/$checksums" ||
    fail "could not download checksums: $base/$checksums; re-run with SCOUT_SKIP_VERIFY=1 to bypass"

  expected=$(grep " $archive\$" "$tmp/$checksums" | awk '{print $1}')
  [ -n "$expected" ] ||
    fail "$archive is not listed in $checksums; re-run with SCOUT_SKIP_VERIFY=1 to bypass"

  actual=$(sha256 "$tmp/$archive" | awk '{print $1}')
  [ "$actual" = "$expected" ] ||
    fail "checksum mismatch for $archive: expected $expected, got $actual"

  info "checksum ok"
fi

# --- extract & install -----------------------------------------------------

tar -xzf "$tmp/$archive" -C "$tmp" || fail "failed to extract $archive"
[ -f "$tmp/$BINARY" ] || fail "binary '$BINARY' not found in archive"
chmod +x "$tmp/$BINARY"

bin_dir="${SCOUT_BIN_DIR:-$HOME/.local/bin}"
if ! mkdir -p "$bin_dir" 2>/dev/null || [ ! -w "$bin_dir" ]; then
  bin_dir="/usr/local/bin"
  info "falling back to $bin_dir"
fi

dest="$bin_dir/$BINARY"
if [ -w "$bin_dir" ]; then
  mv "$tmp/$BINARY" "$dest"
else
  info "elevated permissions needed to write $bin_dir"
  sudo mv "$tmp/$BINARY" "$dest"
fi

info "installed $BINARY $tag -> $dest"

case ":$PATH:" in
  *":$bin_dir:"*) ;;
  *) info "note: $bin_dir is not on your PATH; add it to your shell profile" ;;
esac

info "run 'scout' to get started"
