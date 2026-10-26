#!/bin/sh
# Fixed, embedded script: no user-supplied shell fragments.
set -eu
family=$1
install_packages() {
    case "$family" in
        debian|ubuntu) DEBIAN_FRONTEND=noninteractive apt-get install --yes "$@" ;;
        arch) pacman -S --needed --noconfirm "$@" ;;
        fedora) dnf install -y "$@" ;;
        suse) zypper --non-interactive install "$@" ;;
        alpine) apk add "$@" ;;
        *) echo "Unsupported distribution: $family" >&2; exit 1 ;;
    esac
}
case "$family" in
    debian|ubuntu) apt-get update ;;
    alpine) apk update ;;
esac
install_packages calamares

# Inspect the installed executable, not its version: Calamares 3.3 can use Qt5 or Qt6.
calamares_path=$(command -v calamares)
libraries=$(ldd "$calamares_path")
case "$libraries" in
    *libQt6Core*) qt=6 ;;
    *libQt5Core*) qt=5 ;;
    *) echo "Cannot determine Calamares Qt version; slideshow dependencies were not installed." >&2; exit 1 ;;
esac
echo "Installing slideshow dependencies for Qt $qt..."
case "$family:$qt" in
    debian:5|ubuntu:5)
        install_packages qml-module-qtquick2 qml-module-qtquick-controls qml-module-qtquick-controls2 qml-module-qtquick-layouts qml-module-qtquick-window2 ;;
    debian:6|ubuntu:6)
        install_packages qml6-module-qtquick qml6-module-qtquick-controls qml6-module-qtquick-layouts qml6-module-qtquick-window qml6-module-qtquick-templates qml6-module-qtqml-workerscript ;;
    arch:5) install_packages qt5-declarative qt5-quickcontrols qt5-quickcontrols2 ;;
    arch:6) install_packages qt6-declarative ;;
    fedora:5) install_packages qt5-qtdeclarative qt5-qtquickcontrols qt5-qtquickcontrols2 ;;
    fedora:6) install_packages qt6-qtdeclarative ;;
    suse:5) install_packages libqt5-qtdeclarative-imports libqt5-qtquickcontrols libqt5-qtquickcontrols2 ;;
    suse:6) install_packages qt6-declarative-imports ;;
    alpine:5) install_packages qt5-qtdeclarative qt5-qtquickcontrols qt5-qtquickcontrols2 ;;
    alpine:6) install_packages qt6-qtdeclarative ;;
esac
if [ "$family" = ubuntu ]; then
    install_packages language-selector-common
fi
if [ "$family" = alpine ]; then
    # Alpine splits installer modules into individual packages.
    install_packages calamares-mod-bootloader calamares-mod-displaymanager \
        calamares-mod-finished calamares-mod-fstab calamares-mod-grubcfg \
        calamares-mod-hwclock calamares-mod-keyboard calamares-mod-locale \
        calamares-mod-machineid calamares-mod-mkinitfs calamares-mod-mount \
        calamares-mod-packages calamares-mod-partition calamares-mod-removeuser \
        calamares-mod-services-openrc calamares-mod-shellprocess calamares-mod-summary \
        calamares-mod-umount calamares-mod-unpackfs calamares-mod-users calamares-mod-welcome
fi
echo "Calamares and slideshow dependencies installed successfully."
