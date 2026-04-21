#!/bin/sh
# scripts/test-install.sh — Happy-path test for install.sh.
#
# Builds a goreleaser snapshot, serves dist/ on localhost, points install.sh
# at it via SPOT_RELEASE_API_URL / SPOT_RELEASE_ASSET_BASE, and confirms the
# installed binary reports the expected snapshot version.
#
# Not wired into CI — the local HTTP server is awkward on hosted runners.
# Run manually: `make test-install`.

set -eu

cd "$(dirname "$0")/.."

require() {
    if ! command -v "$1" >/dev/null 2>&1; then
        echo "test-install.sh: required command '$1' not found" >&2
        exit 1
    fi
}
require curl
require tar
require python3

# 1. Build snapshot artifacts.
echo "[1/4] Running goreleaser snapshot..."
TAP_GITHUB_TOKEN=fake-for-snapshot goreleaser release --snapshot --clean >/dev/null

# 2. Figure out which archive matches this machine.
os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in x86_64|amd64) arch=amd64 ;; aarch64|arm64) arch=arm64 ;; esac

archive=$(ls dist/spot_*_"${os}"_"${arch}".tar.gz 2>/dev/null | head -n 1)
if [ -z "${archive:-}" ]; then
    echo "test-install.sh: no snapshot archive for ${os}_${arch}" >&2
    exit 1
fi
# dist/spot_<version>_<os>_<arch>.tar.gz → <version>
version=$(basename "$archive" | sed -E "s/^spot_(.+)_${os}_${arch}\.tar\.gz\$/\1/")
tag="v${version}"

# 3. Write a fake releases/latest response pointing at the snapshot tag.
cat > dist/releases-latest.json <<EOF
{"tag_name":"${tag}"}
EOF

# 4. Serve dist/ on a free port.
PORT=${SPOT_TEST_PORT:-8765}
echo "[2/4] Serving dist/ on http://127.0.0.1:${PORT}"
python3 -m http.server "$PORT" --directory dist >/tmp/spot-test-install-server.log 2>&1 &
server_pid=$!
# shellcheck disable=SC2064
trap "kill $server_pid 2>/dev/null || true" EXIT INT TERM

# Give the server a moment to bind.
sleep 1

# 5. Run install.sh.
install_dir=$(mktemp -d)
echo "[3/4] Running install.sh into ${install_dir}"
SPOT_RELEASE_API_URL="http://127.0.0.1:${PORT}/releases-latest.json" \
    SPOT_RELEASE_ASSET_BASE="http://127.0.0.1:${PORT}" \
    SPOT_INSTALL_DIR="$install_dir" \
    sh install.sh

# 6. Assert the binary installed and reports the snapshot version.
echo "[4/4] Verifying installed binary..."
if [ ! -x "$install_dir/spot" ]; then
    echo "FAIL: spot binary not installed at $install_dir/spot"
    exit 1
fi
actual=$("$install_dir/spot" --version 2>&1)
if ! printf '%s' "$actual" | grep -qF "$version"; then
    echo "FAIL: expected --version to mention ${version}, got: $actual"
    exit 1
fi

echo "PASS: install.sh installed spot ${tag} successfully."
