#!/usr/bin/env sh
# Paterna installer
# uso: curl -fsSL https://raw.githubusercontent.com/hugaojanuario/Paterna/main/install.sh | sh
#      curl -fsSL https://.../install.sh | sh -s -- v0.2.1   (versão específica)
set -e

REPO="hugaojanuario/Paterna"
BIN_NAME="paterna"
VERSION="${1:-latest}"

red()   { printf '\033[31m%s\033[0m\n' "$1"; }
green() { printf '\033[32m%s\033[0m\n' "$1"; }
info()  { printf '\033[36m==>\033[0m %s\n' "$1"; }

# detecta OS
case "$(uname -s)" in
    Darwin) OS="darwin" ;;
    Linux)  OS="linux" ;;
    *)
        red "OS não suportado: $(uname -s)"
        exit 1
        ;;
esac

# detecta arch
case "$(uname -m)" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)
        red "Arquitetura não suportada: $(uname -m)"
        exit 1
        ;;
esac

# decide install dir
if [ -w "/usr/local/bin" ] 2>/dev/null; then
    INSTALL_DIR="/usr/local/bin"
    SUDO=""
elif [ "$(id -u)" = "0" ]; then
    INSTALL_DIR="/usr/local/bin"
    SUDO=""
elif command -v sudo >/dev/null 2>&1; then
    INSTALL_DIR="/usr/local/bin"
    SUDO="sudo"
else
    INSTALL_DIR="$HOME/.local/bin"
    SUDO=""
    mkdir -p "$INSTALL_DIR"
fi

info "OS:      $OS"
info "Arch:    $ARCH"
info "Versão:  $VERSION"
info "Destino: $INSTALL_DIR"

# monta URL — usa redirect direto, sem chamar a API do GitHub
TARBALL="paterna_${OS}_${ARCH}.tar.gz"
if [ "$VERSION" = "latest" ]; then
    URL="https://github.com/$REPO/releases/latest/download/$TARBALL"
else
    URL="https://github.com/$REPO/releases/download/$VERSION/$TARBALL"
fi

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

info "Baixando $URL..."
if ! curl -fsSL --connect-timeout 10 --max-time 120 "$URL" -o "$TMP_DIR/$TARBALL"; then
    red "Falha ao baixar $URL"
    red "Confira se a release existe em: https://github.com/$REPO/releases"
    exit 1
fi

info "Extraindo..."
tar -xzf "$TMP_DIR/$TARBALL" -C "$TMP_DIR"

if [ ! -f "$TMP_DIR/$BIN_NAME" ]; then
    red "Binário $BIN_NAME não encontrado dentro do tarball."
    exit 1
fi

info "Instalando em $INSTALL_DIR/$BIN_NAME..."
$SUDO install -m 0755 "$TMP_DIR/$BIN_NAME" "$INSTALL_DIR/$BIN_NAME"

# macOS: remove flag de quarentena para evitar bloqueio Gatekeeper
if [ "$OS" = "darwin" ]; then
    $SUDO xattr -d com.apple.quarantine "$INSTALL_DIR/$BIN_NAME" 2>/dev/null || true
fi

green "Paterna instalado com sucesso!"

# avisa se INSTALL_DIR não está no PATH
case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *)
        printf '\n'
        red "Atenção: $INSTALL_DIR não está no seu PATH."
        echo "Adicione esta linha ao seu ~/.zshrc ou ~/.bashrc:"
        echo "    export PATH=\"$INSTALL_DIR:\$PATH\""
        ;;
esac

printf '\n'
"$INSTALL_DIR/$BIN_NAME" version || true
printf '\nRode: %s\n' "$BIN_NAME"
