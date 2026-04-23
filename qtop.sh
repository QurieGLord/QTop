#!/bin/bash
# QTop - Professional Build & Installation Utility
# This script handles binary compilation, multi-distro packaging, and system installation.

set -e

VERSION="1.0.0"
PROJECT_ROOT=$(pwd)
DIST_DIR="$PROJECT_ROOT/dist"

# Professional Color Palette
BOLD='\033[1m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Decorative icons
ICON_BUILD="🔨"
ICON_PKG="📦"
ICON_INSTALL="🚀"
ICON_SUCCESS="✅"
ICON_ERROR="❌"

mkdir -p "$DIST_DIR"

# Built state trackers
BUILT_ARCH=false
BUILT_DEB=false
BUILT_RPM=false
INSTALL_AFTER_BUILD=false

print_status() { echo -e "${BLUE}${BOLD}==>${NC} ${BOLD}$1${NC}"; }
print_success() { echo -e "${GREEN}${ICON_SUCCESS} $1${NC}"; }
print_error() { echo -e "${RED}${ICON_ERROR} $1${NC}"; }
print_warning() { echo -e "${YELLOW}! $1${NC}"; }

check_requirements() {
    if ! command -v go &> /dev/null; then
        print_error "Go compiler not found. Please install Go 1.21+ to proceed."
        exit 1
    fi
}

build_binary() {
    print_status "${ICON_BUILD} Compiling QTop binary..."
    go build -o qtop ./cmd/qtop
    print_success "Binary compiled successfully."
}

build_arch() {
    print_status "${ICON_PKG} Creating Arch Linux package..."
    cd "$PROJECT_ROOT/packaging/arch"
    rm -rf src pkg qtop-*.pkg.tar.zst
    makepkg -f
    mv qtop-*.pkg.tar.zst "$DIST_DIR/"
    cd "$PROJECT_ROOT"
    BUILT_ARCH=true
    print_success "Arch Linux package generated in dist/"
}

build_debian() {
    if ! command -v dpkg-deb &> /dev/null; then
        print_error "dpkg-deb utility not found. Cannot build Debian package."
        return
    fi
    print_status "${ICON_PKG} Creating Debian package..."
    BUILD_DIR="/tmp/qtop-deb-build"
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR/usr/bin" "$BUILD_DIR/DEBIAN"
    
    cp qtop "$BUILD_DIR/usr/bin/"
    cp packaging/debian/control "$BUILD_DIR/DEBIAN/"
    
    dpkg-deb --build "$BUILD_DIR" "$DIST_DIR/qtop_${VERSION}_amd64.deb" > /dev/null
    rm -rf "$BUILD_DIR"
    BUILT_DEB=true
    print_success "Debian package generated in dist/"
}

build_rpm() {
    if ! command -v rpmbuild &> /dev/null; then
        print_error "rpmbuild utility not found. Cannot build RPM package."
        return
    fi
    print_status "${ICON_PKG} Creating RPM package..."
    BUILD_DIR="/tmp/qtop-rpm-build"
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
    
    cp qtop "$BUILD_DIR/BUILD/"
    cp packaging/rpm/qtop.spec "$BUILD_DIR/SPECS/"
    
    rpmbuild --define "_topdir $BUILD_DIR" -ba "$BUILD_DIR/SPECS/qtop.spec" > /dev/null 2>&1
    find "$BUILD_DIR/RPMS" -name "*.rpm" -exec cp {} "$DIST_DIR/" \;
    rm -rf "$BUILD_DIR"
    BUILT_RPM=true
    print_success "RPM package generated in dist/"
}

install_packages() {
    echo ""
    print_status "${ICON_INSTALL} Proceeding to system installation (sudo may be required)..."
    
    if [ "$BUILT_ARCH" = true ]; then
        sudo pacman -U --noconfirm "$DIST_DIR"/qtop-"$VERSION"-*.pkg.tar.zst
    fi

    if [ "$BUILT_DEB" = true ]; then
        sudo dpkg -i "$DIST_DIR/qtop_${VERSION}_amd64.deb"
    fi

    if [ "$BUILT_RPM" = true ]; then
        if command -v dnf &> /dev/null; then
            sudo dnf install -y "$DIST_DIR"/qtop-"$VERSION"-*.rpm
        else
            sudo rpm -i "$DIST_DIR"/qtop-"$VERSION"-*.rpm
        fi
    fi
}

usage() {
    echo -e "${BOLD}QTop Management Utility v${VERSION}${NC}"
    echo ""
    echo -e "${BOLD}Usage:${NC}"
    echo "  $0 [options]"
    echo ""
    echo -e "${BOLD}Options:${NC}"
    echo "  -a    Build Arch Linux package (.pkg.tar.zst)"
    echo "  -d    Build Debian package (.deb)"
    echo "  -r    Build RPM package (.rpm)"
    echo "  -s    Build from source (single binary output)"
    echo "  -i    Install the built package(s) into the system"
    echo "  -h    Show this help message"
    echo ""
    echo -e "${BOLD}Examples:${NC}"
    echo "  $0 -a -i    # Build and install for Arch Linux"
    echo "  $0 -s       # Just compile the binary"
    exit 1
}

if [ $# -eq 0 ]; then usage; fi

check_requirements

# First pass: identify intent
while getopts "adrshi" opt; do
    case "$opt" in
        i) INSTALL_AFTER_BUILD=true ;;
        a|d|r|s) ;; 
        h) usage ;;
        *) usage ;;
    esac
done

# Second pass: execute builds
OPTIND=1
while getopts "adrshi" opt; do
    case "$opt" in
        a) build_binary; build_arch ;;
        d) build_binary; build_debian ;;
        r) build_binary; build_rpm ;;
        s) build_binary; print_success "Stand-alone binary is ready at ./qtop" ;;
    esac
done

# Perform installation if requested
if [ "$INSTALL_AFTER_BUILD" = true ]; then
    if [ "$BUILT_ARCH" = false ] && [ "$BUILT_DEB" = false ] && [ "$BUILT_RPM" = false ]; then
        print_warning "No package was built. Please specify a platform (e.g., -a, -d, or -r) with -i."
        exit 1
    fi
    install_packages
    echo ""
    print_success "Deployment complete! You can now launch QTop by typing 'qtop'."
fi
