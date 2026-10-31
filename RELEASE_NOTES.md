# Penguins GUI v26.10.31

A modern desktop interface for creating Penguins' Eggs live images and clones,
cooking system recipes with Penguins Chef, participating in the Penguins Eggs Community,
managing system installers and setup tools, and booting live ISOs directly from disk without a USB drive.

## What's New in v26.10.31

### 🛠️ Unified Penguins Tools Installation
* **Install Penguins Tools**: Replaced separate menu items for Penguins' Eggs and Chef with a single unified action under **Edit → Install Penguins Tools**. It configures the official Penguins repository and installs both `penguins-eggs` and `penguins-chef` in one streamlined operation.
* **Streamlined Post-Install Workflow**: Automatically detects newly installed CLI tools and enables remastering capabilities without requiring an application restart.

### 🧹 Streamlined Menus & System Cleanup
* **Removed Legacy Calamares Action**: Cleaned up the Edit menu and system scripts by removing the standalone Calamares installer action, keeping the interface focused on essential Penguins tooling.

## Installation

### AppImage
```sh
chmod +x penguins-gui-26.10.31-1-x86_64.AppImage
./penguins-gui-26.10.31-1-x86_64.AppImage
```

### Debian / Ubuntu / Devuan (.deb)
```sh
sudo apt install ./penguins-gui_26.10.31-1_amd64.deb
```

### Fedora (.rpm)
```sh
sudo dnf install ./penguins-gui-26.10.31-1.fc44.x86_64.rpm
```

### openSUSE (.rpm)
```sh
sudo zypper install ./penguins-gui-26.10.31-1.opensuse.x86_64.rpm
```

### Arch Linux / Manjaro (.pkg.tar.zst)
```sh
sudo pacman -U ./penguins-gui-26.10.31-1-x86_64.pkg.tar.zst
```

Run Penguins GUI as a normal desktop user. SHA256SUMS is provided alongside the
packages for download verification.
