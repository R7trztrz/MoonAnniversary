#!/bin/sh
# Downloads and runs the latest macOS or Linux release without installing anything.
# Author: Simon Tian
set -eu

repository="${MOON_REPOSITORY:?Set MOON_REPOSITORY to owner/repository}"
system=$(uname -s | tr '[:upper:]' '[:lower:]')
machine=$(uname -m)

case "$system" in
  linux) platform="linux" ;;
  darwin) platform="darwin" ;;
  *) echo "Unsupported operating system: $system" >&2; exit 1 ;;
esac

case "$machine" in
  x86_64|amd64) architecture="amd64" ;;
  arm64|aarch64) architecture="arm64" ;;
  *) echo "Unsupported architecture: $machine" >&2; exit 1 ;;
esac

asset="moon-${platform}-${architecture}"
url="https://github.com/${repository}/releases/latest/download/${asset}"
destination="${TMPDIR:-/tmp}/moon-$$"
checksum_file="${destination}.sha256"
trap 'rm -f "$destination" "$checksum_file"' EXIT HUP INT TERM

curl --fail --location --silent --show-error "$url" --output "$destination"
curl --fail --location --silent --show-error \
  "https://github.com/${repository}/releases/latest/download/SHA256SUMS" \
  --output "$checksum_file"
expected=$(awk -v name="$asset" '$2 == name || $2 == "*" name { print $1 }' "$checksum_file")
if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$destination" | awk '{ print $1 }')
else
  actual=$(shasum -a 256 "$destination" | awk '{ print $1 }')
fi
if [ -z "$expected" ] || [ "$actual" != "$expected" ]; then
  echo "The downloaded executable failed SHA-256 verification." >&2
  exit 1
fi
chmod +x "$destination"
"$destination"
