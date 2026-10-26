#!/bin/sh
set -eu
case "$1" in
    debian|ubuntu)
        apt-get update
        DEBIAN_FRONTEND=noninteractive apt-get install --yes penguins-chef
        ;;
    arch|manjaro)
        # Refresh and upgrade together to avoid an unsupported partial upgrade.
        pacman -Syu --needed --noconfirm penguins-chef
        ;;
    fedora|el9) dnf --refresh install -y penguins-chef ;;
    suse) zypper --non-interactive refresh; zypper --non-interactive install penguins-chef ;;
    alpine) apk update; apk add penguins-chef ;;
    *) echo "Unsupported distribution: $1" >&2; exit 1 ;;
esac
