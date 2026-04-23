#!/bin/bash
# QTop Package Builder & Installer Script
set -e

VERSION="1.0.0"
PROJECT_ROOT=$(pwd)
DIST_DIR="$PROJECT_ROOT/dist"

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

mkdir -p "$DIST_DIR"

# Переменные для отслеживания того, что было собрано
BUILT_ARCH=false
BUILT_DEB=false
BUILT_RPM=false
INSTALL_AFTER_BUILD=false

check_go() {
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: Go is not installed.${NC}"
        exit 1
    fi
}

build_binary() {
    echo -e "${GREEN}==> Building qtop binary...${NC}"
    go build -o qtop ./cmd/qtop
}

build_arch() {
    echo -e "${GREEN}==> Building Arch Linux package...${NC}"
    cd "$PROJECT_ROOT/packaging/arch"
    rm -rf src pkg qtop-*.pkg.tar.zst
    makepkg -f
    mv qtop-*.pkg.tar.zst "$DIST_DIR/"
    echo -e "${GREEN}==> Arch package saved to dist/${NC}"
    cd "$PROJECT_ROOT"
    BUILT_ARCH=true
}

build_debian() {
    if ! command -v dpkg-deb &> /dev/null; then
        echo -e "${RED}Error: dpkg-deb not found. Install 'dpkg' (available in AUR for Arch).${NC}"
        return
    fi
    echo -e "${GREEN}==> Building Debian package...${NC}"
    BUILD_DIR="/tmp/qtop-deb-build"
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR/usr/bin"
    mkdir -p "$BUILD_DIR/DEBIAN"
    
    cp qtop "$BUILD_DIR/usr/bin/"
    cp packaging/debian/control "$BUILD_DIR/DEBIAN/"
    
    dpkg-deb --build "$BUILD_DIR" "$DIST_DIR/qtop_${VERSION}_amd64.deb"
    rm -rf "$BUILD_DIR"
    echo -e "${GREEN}==> Debian package saved to dist/${NC}"
    BUILT_DEB=true
}

build_rpm() {
    if ! command -v rpmbuild &> /dev/null; then
        echo -e "${RED}Error: rpmbuild not found. Install 'rpm-tools' (available in AUR for Arch).${NC}"
        return
    fi
    echo -e "${GREEN}==> Building RPM package...${NC}"
    BUILD_DIR="/tmp/qtop-rpm-build"
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
    
    cp qtop "$BUILD_DIR/BUILD/"
    cp packaging/rpm/qtop.spec "$BUILD_DIR/SPECS/"
    
    rpmbuild --define "_topdir $BUILD_DIR" -ba "$BUILD_DIR/SPECS/qtop.spec"
    find "$BUILD_DIR/RPMS" -name "*.rpm" -exec cp {} "$DIST_DIR/" \;
    rm -rf "$BUILD_DIR"
    echo -e "${GREEN}==> RPM package saved to dist/${NC}"
    BUILT_RPM=true
}

install_packages() {
    if [ "$BUILT_ARCH" = true ]; then
        echo -e "${GREEN}==> Installing Arch package...${NC}"
        sudo pacman -U --noconfirm "$DIST_DIR"/qtop-"$VERSION"-*.pkg.tar.zst
    fi

    if [ "$BUILT_DEB" = true ]; then
        echo -e "${GREEN}==> Installing Debian package...${NC}"
        sudo dpkg -i "$DIST_DIR/qtop_${VERSION}_amd64.deb"
    fi

    if [ "$BUILT_RPM" = true ]; then
        echo -e "${GREEN}==> Installing RPM package...${NC}"
        if command -v dnf &> /dev/null; then
            sudo dnf install -y "$DIST_DIR"/qtop-"$VERSION"-*.rpm
        else
            sudo rpm -i "$DIST_DIR"/qtop-"$VERSION"-*.rpm
        fi
    fi
}

usage() {
    echo "Usage: $0 [-a] [-d] [-r] [-s] [-i]"
    echo "  -a: Build Arch Linux package"
    echo "  -d: Build Debian package"
    echo "  -r: Build RPM package"
    echo "  -s: Build from source (binary only)"
    echo "  -i: Install built packages (requires sudo)"
    echo ""
    echo "Example (build and install on Arch): $0 -a -i"
    exit 1
}

if [ $# -eq 0 ]; then usage; fi

check_go

# Сначала собираем аргументы
while getopts "adrsi" opt; do
    case "$opt" in
        i) INSTALL_AFTER_BUILD=true ;;
        a|d|r|s) ;; # Будем обрабатывать во втором проходе
        *) usage ;;
    esac
done

# Сбрасываем getopts для второго прохода сборки
OPTIND=1
while getopts "adrsi" opt; do
    case "$opt" in
        a) build_binary; build_arch ;;
        d) build_binary; build_debian ;;
        r) build_binary; build_rpm ;;
        s) build_binary; echo -e "${GREEN}==> Binary build complete: ./qtop${NC}" ;;
        i) ;; # Уже обработано
    esac
done

# Выполняем установку, если был флаг -i
if [ "$INSTALL_AFTER_BUILD" = true ]; then
    install_packages
    echo -e "${GREEN}==> QTop has been installed! You can now run 'qtop' from your terminal.${NC}"
fi
