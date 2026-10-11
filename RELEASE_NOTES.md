# Penguins GUI v26.10.11

A modern desktop interface for creating Penguins' Eggs live images and clones,
cooking system recipes with Penguins Chef, managing system installers and setup tools,
and booting live ISOs directly from disk without a USB drive.

## What's New in v26.10.11

### 🍳 Penguins Chef Integration (Cook)
* **Interactive Recipe Picker**: New "Cook…" action in the toolbar and Edit menu to browse and apply Penguins Chef recipes.
* **Full-Window Expanded View**: Recipe picker expands to the full window canvas with classic minimize, maximize/restore, and close controls.
* **Live YAML Preview & Details**: Inspect recipe definitions, distribution profile support, and descriptions before running.
* **Apply & Dry-Run**: Apply recipes directly or execute a dry run with live logging.
* **Automatic Recipe Repository Updates**: Download and update `~/.chef` recipes repository directly from the GUI.
* **Streamlined Categories**: Low-level recipes (`base`, `de`, `dm`, `costumes`) are filtered for a clean desktop experience.

### 🚀 Boot ISO Clarification
* **Renamed Action**: Renamed the bootloader integration action to "Boot ISO…" in the toolbar and menus for clearer user intent.

### 📦 Packaging Updates
* **Arch Linux / Manjaro**: Added `xorg-xrandr` dependency in packaging templates for full resolution and display handling.

## Installation

### AppImage
```sh
chmod +x penguins-gui-26.10.11-1-x86_64.AppImage
./penguins-gui-26.10.11-1-x86_64.AppImage
```

### Debian / Ubuntu / Devuan (.deb)
```sh
sudo apt install ./penguins-gui_26.10.11-1_amd64.deb
```

### Fedora (.rpm)
```sh
sudo dnf install ./penguins-gui-26.10.11-1.fc44.x86_64.rpm
```

### openSUSE (.rpm)
```sh
sudo zypper install ./penguins-gui-26.10.11-1.opensuse.x86_64.rpm
```

### Arch Linux / Manjaro (.pkg.tar.zst)
```sh
sudo pacman -U ./penguins-gui-26.10.11-1-x86_64.pkg.tar.zst
```

Run Penguins GUI as a normal desktop user. SHA256SUMS is provided alongside the
packages for download verification.
