#!/bin/sh
# install.sh — Install the `spot` CLI on macOS or Linux.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/spot-nyc/spot/main/install.sh | sh
#
# Environment overrides:
#   SPOT_INSTALL_DIR      Target install directory (default /usr/local/bin, or
#                         $HOME/.local/bin if the former is not writable).
#   SPOT_RELEASE_API_URL  Override the GitHub releases/latest endpoint (tests).
#   SPOT_RELEASE_ASSET_BASE
#                         Override the base URL for release asset downloads.

set -eu

REPO="spot-nyc/spot"
API_URL_DEFAULT="https://api.github.com/repos/${REPO}/releases/latest"

err() {
    printf 'install.sh: %s\n' "$*" >&2
}

detect_os() {
    case "$(uname -s)" in
        Darwin) echo darwin ;;
        Linux) echo linux ;;
        *)
            err "Unsupported OS $(uname -s)."
            err "For Windows, use Scoop:"
            err "  scoop bucket add spot-nyc https://github.com/spot-nyc/scoop-bucket"
            err "  scoop install spot"
            exit 1
            ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo amd64 ;;
        aarch64|arm64) echo arm64 ;;
        *)
            err "Unsupported architecture $(uname -m)."
            exit 1
            ;;
    esac
}

pick_sha_cmd() {
    if command -v sha256sum >/dev/null 2>&1; then
        echo "sha256sum"
    elif command -v shasum >/dev/null 2>&1; then
        echo "shasum -a 256"
    else
        err "Neither sha256sum nor shasum is installed."
        exit 1
    fi
}

pick_install_dir() {
    if [ -n "${SPOT_INSTALL_DIR:-}" ]; then
        mkdir -p "$SPOT_INSTALL_DIR"
        echo "$SPOT_INSTALL_DIR"
        return
    fi
    if [ -w /usr/local/bin ]; then
        echo /usr/local/bin
        return
    fi
    mkdir -p "$HOME/.local/bin"
    echo "$HOME/.local/bin"
}

main() {
    os=$(detect_os)
    arch=$(detect_arch)
    sha_cmd=$(pick_sha_cmd)

    api_url="${SPOT_RELEASE_API_URL:-$API_URL_DEFAULT}"
    printf 'Fetching latest release from %s\n' "$api_url"
    tag=$(curl -fsSL "$api_url" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n 1)
    if [ -z "${tag:-}" ]; then
        err "Could not determine latest release tag."
        exit 1
    fi
    version="${tag#v}"
    archive_name="spot_${version}_${os}_${arch}.tar.gz"

    asset_base="${SPOT_RELEASE_ASSET_BASE:-https://github.com/${REPO}/releases/download/${tag}}"
    tmpdir=$(mktemp -d)
    # shellcheck disable=SC2064
    trap "rm -rf '$tmpdir'" EXIT INT TERM

    printf 'Downloading %s\n' "$archive_name"
    curl -fsSL -o "$tmpdir/$archive_name" "$asset_base/$archive_name"
    curl -fsSL -o "$tmpdir/checksums.txt" "$asset_base/checksums.txt"

    printf 'Verifying checksum\n'
    expected=$(grep " $archive_name\$" "$tmpdir/checksums.txt" | awk '{print $1}')
    if [ -z "${expected:-}" ]; then
        err "Could not find $archive_name in checksums.txt."
        exit 1
    fi
    actual=$($sha_cmd "$tmpdir/$archive_name" | awk '{print $1}')
    if [ "$expected" != "$actual" ]; then
        err "Checksum mismatch! expected=$expected actual=$actual"
        exit 1
    fi

    tar -xzf "$tmpdir/$archive_name" -C "$tmpdir"

    install_dir=$(pick_install_dir)
    mv "$tmpdir/spot" "$install_dir/spot"
    chmod +x "$install_dir/spot"

    printf '\nInstalled spot %s to %s/spot\n' "$tag" "$install_dir"
    case ":${PATH}:" in
        *":$install_dir:"*) ;;
        *) printf 'NOTE: %s is not on your PATH. Add it to your shell profile.\n' "$install_dir" ;;
    esac
    printf 'Run `spot --version` to verify.\n'
}

main "$@"
