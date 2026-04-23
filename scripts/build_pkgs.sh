#!/bin/bash
# QTop Package Builder Script
set -e

VERSION="1.0.0"
PROJECT_ROOT=$(pwd)
DIST_DIR="$PROJECT_ROOT/dist"

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

mkdir -p "$DIST_DIR"

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
    # Очистка старых билдов
    rm -rf src pkg qtop-*.pkg.tar.zst
    makepkg -f
    mv qtop-*.pkg.tar.zst "$DIST_DIR/"
    echo -e "${GREEN}==> Arch package saved to dist/${NC}"
    cd "$PROJECT_ROOT"
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
}

usage() {
    echo "Usage: $0 [-a] [-d] [-r] [-s]"
    echo "  -a: Build Arch Linux package (using makepkg)"
    echo "  -d: Build Debian package (using dpkg-deb)"
    echo "  -r: Build RPM package (using rpmbuild)"
    echo "  -s: Build from source (just binary)"
    echo ""
    echo "Example: $0 -a -d"
    exit 1
}

if [ $# -eq 0 ]; then usage; fi

check_go

# Обработка флагов
while getopts "adrs" opt; do
    case "$opt" in
        a) build_binary; build_arch ;;
        d) build_binary; build_debian ;;
        r) build_binary; build_rpm ;;
        s) build_binary; echo -e "${GREEN}==> Binary build complete: ./qtop${NC}" ;;
        *) usage ;;
    esac
done
