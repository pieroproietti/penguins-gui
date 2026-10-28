# Penguins GUI v26.10.28

A modern desktop interface for creating Penguins' Eggs live images and clones,
cooking system recipes with Penguins Chef, participating in the Penguins Eggs Community,
managing system installers and setup tools, and booting live ISOs directly from disk without a USB drive.

## What's New in v26.10.28

### 🖥️ Toolbar & Top Bar Layout
* **Right-Aligned Top Bar Actions**: Moved `Docs`, `Stats`, and `Exit` to the right side of the primary top bar alongside the `Remaster` and `Community` tabs.
* **Streamlined Action Toolbar**: Removed the duplicate `Community` button from the secondary toolbar, keeping it focused on `Create ISO`, `Kill`, `Clean`, `Cook`, and `Boot ISO`.

### 🌍 Penguins Eggs Community Refinements
* **ON / OFF Toggle Button**: Replaced the community checkbox and status text with a clean `Join the Penguins Eggs Community: ON / OFF` toggle button.
* **Visible Read-Only State when OFF**: When `OFF`, saved profile fields remain visible but disabled while public data is removed from the Community server. When `ON`, fields are enabled with automatic local saving and server synchronization.

### 🍳 Penguins Chef CLI Update
* **Updated Chef Commands**: Replaced `chef get` and `chef apply` with `chef cook --pull` and `chef cook`.

## Installation

### AppImage
```sh
chmod +x penguins-gui-26.10.28-1-x86_64.AppImage
./penguins-gui-26.10.28-1-x86_64.AppImage
```

### Debian / Ubuntu / Devuan (.deb)
```sh
sudo apt install ./penguins-gui_26.10.28-1_amd64.deb
```

### Fedora (.rpm)
```sh
sudo dnf install ./penguins-gui-26.10.28-1.fc44.x86_64.rpm
```

### openSUSE (.rpm)
```sh
sudo zypper install ./penguins-gui-26.10.28-1.opensuse.x86_64.rpm
```

### Arch Linux / Manjaro (.pkg.tar.zst)
```sh
sudo pacman -U ./penguins-gui-26.10.28-1-x86_64.pkg.tar.zst
```

Run Penguins GUI as a normal desktop user. SHA256SUMS is provided alongside the
packages for download verification.
