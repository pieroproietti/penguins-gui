#!/bin/sh
# Repository locations and signing key identity follow coa/pkg/repo in Penguins' Eggs.
set -eu
family=$1
action=$2
base=https://penguins-eggs.net/repos
key_url=$base/KEY.asc
key_id=F6773EA7D2F309BA3E5DE08A45B10F271525403F
case "$family" in debian|ubuntu|arch|manjaro|fedora|el9|suse|alpine) ;; *) exit 1 ;; esac
case "$action" in add|rm) ;; *) exit 1 ;; esac

# Download into a private directory; never truncate a configured key on failure.
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT HUP INT TERM
fetch() {
    curl --fail --silent --show-error --location --proto '=https' --proto-redir '=https' "$1" -o "$2"
}
case "$family" in
    debian|ubuntu)
        key=/usr/share/keyrings/penguins-repos.gpg
        list=/etc/apt/sources.list.d/penguins-repos.list
        sources=/etc/apt/sources.list.d/penguins-repos.sources
        if [ "$action" = add ]; then
            fetch "$key_url" "$work/key.asc"
            gpg --batch --yes --dearmor --output "$work/key.gpg" "$work/key.asc"
            install -D -m 644 "$work/key.gpg" "$key"
            mkdir -p /etc/apt/sources.list.d
            printf 'Types: deb\nURIs: %s/deb\nSuites: stable\nComponents: main\nSigned-By: %s\n' "$base" "$key" > "$sources"
            rm -f "$list"
        else
            rm -f "$list" "$sources" "$key"
        fi
        ;;
    arch|manjaro)
        conf=/etc/pacman.conf
        if [ "$action" = add ]; then
            pacman-key --recv-key "$key_id" --keyserver keyserver.ubuntu.com
            pacman-key --lsign-key "$key_id"
            if ! grep -q '^\[penguins-eggs\]' "$conf"; then
                printf '\n# penguins-repos\n[penguins-eggs]\nSigLevel = Required DatabaseOptional\nServer = %s/%s\n' "$base" "$family" >> "$conf"
            fi
        else
            # Remove only this section, preserving every unrelated repository.
            awk '/^# penguins-repos$/ {next} /^\[penguins-eggs\]/ {skip=1; next} /^\[/ {skip=0} !skip {print}' "$conf" > "$work/pacman.conf"
            cat "$work/pacman.conf" > "$conf"
        fi
        ;;
    fedora|el9|suse)
        repo=/etc/yum.repos.d/penguins-eggs.repo
        url=$base/rpm/fedora/42
        if [ "$family" = el9 ]; then url=$base/rpm/el9; fi
        if [ "$family" = suse ]; then
            repo=/etc/zypp/repos.d/penguins-eggs.repo
            url=$base/rpm/opensuse/leap
        fi
        if [ "$action" = add ]; then
            fetch "$key_url" "$work/key.asc"
            rpm --import "$work/key.asc"
            printf '[penguins-eggs]\nname=penguins-eggs.net repos\nbaseurl=%s\nenabled=1\ngpgcheck=1\ngpgkey=%s\n' "$url" "$key_url" > "$work/repo"
            install -D -m 644 "$work/repo" "$repo"
        else
            rm -f "$repo"
        fi
        ;;
    alpine)
        key_name=piero.proietti@gmail.com-662b958c.rsa.pub
        key=/etc/apk/keys/$key_name
        conf=/etc/apk/repositories
        url=$base/alpine/
        if [ "$action" = add ]; then
            fetch "$url$key_name" "$work/key"
            install -D -m 644 "$work/key" "$key"
            if ! grep -Fqx "$url" "$conf"; then printf '\n%s\n' "$url" >> "$conf"; fi
        else
            awk -v repo="$url" '$0 != repo {print}' "$conf" > "$work/repositories"
            cat "$work/repositories" > "$conf"
            rm -f "$key"
        fi
        ;;
esac
# RPM and Pacman keys may be shared by other repositories: retain their trust entries.
echo "Native repository operation completed: $action"
