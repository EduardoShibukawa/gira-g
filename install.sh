#!/bin/sh
# install.sh — instala o binário `gira` a partir dos GitHub Releases.
#
# Uso:
#   export GITHUB_TOKEN="$(gh auth token)"   # repo privado exige token
#   curl -fsSL -H "Authorization: Bearer $GITHUB_TOKEN" \
#     https://raw.githubusercontent.com/EduardoShibukawa/gira-g/main/install.sh | sh
#   curl .../install.sh | sh -s -- --version v0.1.0 --dir ~/.local/bin --cosign
#
# O script detecta SO/arquitetura, baixa o tarball do release, confere o
# sha256 contra o `checksums.txt` publicado e instala o binário `gira`.
# Com --cosign, verifica também o bundle Sigstore keyless
# (`checksums.txt.sigstore.json`) contra a identidade do workflow release.
set -eu

REPO="EduardoShibukawa/gira-g"
VERSION="latest"
DIR=""
COSIGN="0"

# O repositório é privado: downloads exigem token com acesso de leitura.
# Exporte GITHUB_TOKEN (ex: export GITHUB_TOKEN="$(gh auth token)") e o
# script envia como Authorization em todas as chamadas.
gh_curl() {
  if [ -n "${GITHUB_TOKEN:-}" ]; then
    curl -fsSL -H "Authorization: Bearer $GITHUB_TOKEN" "$@"
  else
    curl -fsSL "$@"
  fi
}

usage() {
  echo "usage: install.sh [--version vX.Y.Z] [--dir DIR] [--cosign]" >&2
}

while [ $# -gt 0 ]; do
  case "$1" in
    --version) VERSION="${2:?--version precisa de valor}"; shift 2 ;;
    --dir) DIR="${2:?--dir precisa de valor}"; shift 2 ;;
    --cosign) COSIGN="1"; shift ;;
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
  if command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then
    TAG="$(gh release view --json tagName -q .tagName -R "$REPO")"
  elif [ -n "${GITHUB_TOKEN:-}" ]; then
    TAG="$(gh_curl "https://api.github.com/repos/${REPO}/releases/latest" \
      | grep '"tag_name"' | head -n 1 | sed 's/.*: *"//;s/".*//')"
  else
    echo "install.sh: latest exige gh autenticado ou GITHUB_TOKEN" >&2
    exit 1
  fi
  [ -n "$TAG" ] || { echo "install.sh: não resolveu a versão latest" >&2; exit 1; }
fi
ASSET_VER="$(printf '%s' "$TAG" | sed 's/^v//')"
EXT="tar.gz"
[ "$OS" = "windows" ] && EXT="zip"
ASSET="gira_${ASSET_VER}_${OS}_${ARCH}.${EXT}"
WANT_FILES="$ASSET checksums.txt"
[ "$COSIGN" = "1" ] && WANT_FILES="$WANT_FILES checksums.txt.sigstore.json"

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

# Downloads do repo privado via API do GitHub (browser URL não aceita
# header Authorization). Prefere gh autenticado; senão, GITHUB_TOKEN.
echo "baixando ${ASSET} (${TAG})..." >&2
if command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then
  # shellcheck disable=SC2086
  gh release download "$TAG" -R "$REPO" --dir "$TMP" --clobber $(
    for f in $WANT_FILES; do printf ' --pattern %s' "$f"; done
  )
elif [ -n "${GITHUB_TOKEN:-}" ] && command -v python3 >/dev/null 2>&1; then
  GIRA_REPO="$REPO" GIRA_TAG="$TAG" GIRA_FILES="$WANT_FILES" python3 - <<'EOF'
import json, os, urllib.request
repo = os.environ["GIRA_REPO"]
tag = os.environ["GIRA_TAG"]
want = os.environ["GIRA_FILES"].split()
tok = os.environ["GITHUB_TOKEN"]
def api(url, accept="application/vnd.github+json"):
    req = urllib.request.Request(url, headers={
        "Authorization": f"Bearer {tok}", "Accept": accept})
    return urllib.request.urlopen(req)
rel = json.load(api(
    f"https://api.github.com/repos/{repo}/releases/tags/{tag}"))
by_name = {a["name"]: a["id"] for a in rel["assets"]}
for name in want:
    if name not in by_name:
        raise SystemExit(f"asset ausente no release: {name}")
    data = api(
        f"https://api.github.com/repos/{repo}/releases/assets/{by_name[name]}",
        accept="application/octet-stream").read()
    with open(name, "wb") as f:
        f.write(data)
EOF
else
  echo "install.sh: precisa de gh autenticado, ou GITHUB_TOKEN + python3" >&2
  exit 1
fi
for f in $WANT_FILES; do
  [ -s "$f" ] || { echo "install.sh: download falhou: $f" >&2; exit 1; }
done

if [ "$COSIGN" = "1" ]; then
  need cosign
  echo "conferindo assinatura keyless..." >&2
  cosign verify-blob \
    --certificate-identity "https://github.com/${REPO}/.github/workflows/release.yml@refs/tags/${TAG}" \
    --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
    --bundle checksums.txt.sigstore.json checksums.txt
fi

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
