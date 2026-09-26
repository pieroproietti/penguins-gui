#!/bin/sh
set -eu
cd "$(dirname "$0")"
make clean
make package "$@"
. /etc/os-release
for family in "$ID" ${ID_LIKE:-}; do
    case "$family" in
        opensuse|opensuse-leap|opensuse-tumbleweed|opensuse-slowroll|suse)
            sudo zypper install ./dist/penguins-gui-*.rpm
            exit
            ;;
        fedora)
            sudo dnf install dist/penguins-gui-*.rpm
            exit
            ;;
        arch|manjaro)
            sudo pacman -U dist/penguins-gui-*.pkg.tar.zst
            exit
            ;;
        debian|ubuntu|devuan)
            sudo dpkg -i dist/penguins-gui_*.deb
            exit
            ;;
    esac
done
echo "Unsupported distribution: $ID" >&2
exit 1
