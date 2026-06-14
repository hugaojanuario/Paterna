#!/usr/bin/env sh
set -e

REPO="hugaojanuario/Paterna"
BINARY="paterna"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Arquitetura nao suportada: $ARCH"; exit 1 ;;
esac

case "$OS" in
    linux|darwin) ;;
    *) echo "Sistema nao suportado: $OS"; exit 1 ;;
esac

echo "Buscando versao mais recente do Paterna..."

LATEST_TAG=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep '"tag_name":' \
    | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_TAG" ]; then
    echo "Nao foi possivel obter a versao mais recente. Verifique se o repositorio tem releases publicadas."
    exit 1
fi

VERSION="${LATEST_TAG#v}"
ARCHIVE="${BINARY}_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$LATEST_TAG/$ARCHIVE"

echo "Baixando Paterna $LATEST_TAG para $OS/$ARCH..."

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

if ! curl -fsSL "$URL" -o "$TMP/$ARCHIVE"; then
    echo "Falha ao baixar $URL"
    exit 1
fi

tar -xzf "$TMP/$ARCHIVE" -C "$TMP"

if [ ! -f "$TMP/$BINARY" ]; then
    echo "Binario nao encontrado no arquivo baixado."
    exit 1
fi

if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
    chmod +x "$INSTALL_DIR/$BINARY"
else
    echo "Instalando em $INSTALL_DIR (sudo necessario)..."
    sudo mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
    sudo chmod +x "$INSTALL_DIR/$BINARY"
fi

echo ""
echo "Paterna $LATEST_TAG instalado em $INSTALL_DIR/$BINARY"
echo ""
echo "Proximos passos:"
echo "  paterna init    cria conta de administrador"
echo "  paterna         abre o painel TUI"
echo "  paterna serve   sobe a API HTTP"
