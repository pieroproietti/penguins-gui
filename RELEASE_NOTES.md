# Penguins GUI v26.10.8

A modern desktop interface for creating Penguins' Eggs live images and clones,
managing system installers and setup tools, and booting live ISOs directly from
disk without a USB drive.

## What's New in v26.10.8

### 📦 Multi-Distribution Packaging & AppImage
* **Standalone AppImage**: Direct portable AppImage binary (`.AppImage`) for modern 64-bit Linux distributions.
* **openSUSE RPM Support**: Native RPM packages for openSUSE Leap, Tumbleweed, and Slowroll, with zypper integration in `./m` and openSUSE Tumbleweed CI builds.
* **Manjaro & Arch Packages**: Native Arch Linux (`.pkg.tar.zst`) and dedicated Manjaro release packages built in CI.
* **Fedora RPM Packages**: Native RPM packaging and CI workflow for Fedora 44.
* **Unified Release Assets**: All packages (`.deb`, `.pkg.tar.zst`, `.rpm`, `.AppImage`) and SHA256SUMS are published automatically.

### 🐧 System Setup, Tools & Calamares
* **Native Repository Management**: Configure or remove native Penguins' Eggs repositories directly from the GUI without requiring the CLI.
* **Automatic Eggs Setup**: Easily install the `penguins-eggs` CLI on first run.
* **Calamares Installation**: Direct action to install Calamares and matching Qt 5 / Qt 6 slideshow dependencies across supported distributions.
* **Streamlined Menus**: Easy access to configuration tools (Eggs CLI, Calamares, /etc/skel, grub40).

### 🚀 Boot Without USB & GRUB Integration
* **Boot Without USB**: Configure GRUB to boot created ISOs directly from disk, eliminating the need to write to a physical USB drive.
* **Safe GRUB Management**: Cleanly replaces Eggs boot entries and creates non-executable backups to prevent stale duplicate entries.
* **Bootloader Availability**: Automatically detects GRUB availability on startup and provides guidance to update GRUB (`sudo update-grub`).

## Installation

### AppImage
```sh
chmod +x penguins-gui-26.10.8-1-x86_64.AppImage
./penguins-gui-26.10.8-1-x86_64.AppImage
```

### Debian / Ubuntu / Devuan (.deb)
```sh
sudo apt install ./penguins-gui_26.10.8-1_amd64.deb
```

### Fedora (.rpm)
```sh
sudo dnf install ./penguins-gui-26.10.8-1.fc44.x86_64.rpm
```

### openSUSE (.rpm)
```sh
sudo zypper install ./penguins-gui-26.10.8-1.opensuse.x86_64.rpm
```

### Arch Linux / Manjaro (.pkg.tar.zst)
```sh
sudo pacman -U ./penguins-gui-26.10.8-1-x86_64.pkg.tar.zst
```

Run Penguins GUI as a normal desktop user. SHA256SUMS is provided alongside the
packages for download verification.
