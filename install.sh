#!/bin/sh
# install.sh — instala o binário `gira` a partir dos GitHub Releases.
#
# Uso:
#   curl -fsSL https://raw.githubusercontent.com/EduardoShibukawa/gira-g/main/install.sh | sh
#   curl -fsSL .../install.sh | sh -s -- --version v0.1.0 --dir ~/.local/bin
#
# O script detecta SO/arquitetura, baixa o tarball do release, confere o
# sha256 contra o `checksums.txt` publicado e instala o binário `gira`.
# Verificação cosign keyless: follow-up após o primeiro release com
# artefatos (nomes exatos de .sig/.pem ainda não confirmados).
set -eu

REPO="EduardoShibukawa/gira-g"
VERSION="latest"
DIR=""

usage() {
  echo "usage: install.sh [--version vX.Y.Z] [--dir DIR]" >&2
}

while [ $# -gt 0 ]; do
  case "$1" in
    --version) VERSION="${2:?--version precisa de valor}"; shift 2 ;;
    --dir) DIR="${2:?--dir precisa de valor}"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "flag desconhecida: $1" >&2; usage; exit 1 ;;
  esac
done

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "install.sh: precisa de '$1' no PATH" >&2; exit 1
  }
}
need curl
need tar

OS="$(uname -s)"
ARCH="$(uname -m)"
case "$OS" in
  Linux) OS="linux" ;;
  Darwin) OS="darwin" ;;
  MINGW*|MSYS*|CYGWIN*|Windows*) OS="windows" ;;
  *) echo "install.sh: SO não suportado: $OS" >&2; exit 1 ;;
esac
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "install.sh: arquitetura não suportada: $ARCH" >&2; exit 1 ;;
esac
if [ "$OS" = "windows" ] && [ "$ARCH" = "arm64" ]; then
  echo "install.sh: windows/arm64 não é publicado (ver .goreleaser.yaml)" >&2
  exit 1
fi

# GoReleaser publica o tag (v0.1.0) no path, mas o asset usa a versão
# sem o prefixo `v`: gira_0.1.0_linux_amd64.tar.gz
TAG="$VERSION"
if [ "$VERSION" = "latest" ]; then
  TAG="$(curl -fsSL -o /dev/null -w '%{url_effective}' \
    "https://github.com/${REPO}/releases/latest" | sed 's#.*/##')"
  [ -n "$TAG" ] || { echo "install.sh: não resolveu a versão latest" >&2; exit 1; }
fi
ASSET_VER="$(printf '%s' "$TAG" | sed 's/^v//')"
EXT="tar.gz"
[ "$OS" = "windows" ] && EXT="zip"
ASSET="gira_${ASSET_VER}_${OS}_${ARCH}.${EXT}"
BASE="https://github.com/${REPO}/releases/download/${TAG}"

if [ -z "$DIR" ]; then
  if [ -w "/usr/local/bin" ]; then
    DIR="/usr/local/bin"
  else
    DIR="$HOME/.local/bin"
  fi
fi

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT INT TERM
cd "$TMP"

echo "baixando ${ASSET} (${TAG})..." >&2
curl -fsSL -o "$ASSET" "${BASE}/${ASSET}"
curl -fsSL -o checksums.txt "${BASE}/checksums.txt"

echo "conferindo sha256..." >&2
LINE="$(grep "  ${ASSET}\$" checksums.txt || true)"
[ -n "$LINE" ] || { echo "install.sh: asset sem entrada em checksums.txt" >&2; exit 1; }
WANT="$(printf '%s' "$LINE" | awk '{print $1}')"
if command -v sha256sum >/dev/null 2>&1; then
  GOT="$(sha256sum "$ASSET" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  GOT="$(shasum -a 256 "$ASSET" | awk '{print $1}')"
else
  echo "install.sh: precisa de 'sha256sum' ou 'shasum' no PATH" >&2; exit 1
fi
[ "$GOT" = "$WANT" ] || {
  echo "install.sh: checksum divergente (esperado $WANT, obtido $GOT)" >&2
  exit 1
}

echo "extraindo..." >&2
if [ "$EXT" = "zip" ]; then
  need unzip
  unzip -q -o "$ASSET" gira.exe
  BIN="gira.exe"
else
  tar -xzf "$ASSET" gira
  BIN="gira"
fi

mkdir -p "$DIR"
mv "$BIN" "$DIR/"
chmod +x "$DIR/$BIN"

echo "gira instalado em $DIR/$BIN" >&2
"$DIR/$BIN" --version 2>/dev/null || "$DIR/$BIN" version 2>/dev/null || true
case ":$PATH:" in
  *":$DIR:"*) ;;
  *) echo "dica: adicione $DIR ao PATH" >&2 ;;
esac
