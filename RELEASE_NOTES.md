# Penguins GUI v26.10.21

A modern desktop interface for creating Penguins' Eggs live images and clones,
cooking system recipes with Penguins Chef, managing system installers and setup tools,
and booting live ISOs directly from disk without a USB drive.

## What's New in v26.10.21

### 🖥️ Layout & Interface Improvements
* **Refactored Mode Selector**: Converted remaster mode options to a sleek horizontal layout with subtle separators and compact ellipsis descriptions.
* **Maximized Operation Log Area**: Reorganized header layout to maximize the visible operation log area, placing status bar and mode row neatly above a dedicated log control bar.

### 🍳 Penguins Chef Management
* **Install Penguins Chef**: Added an action in the Edit menu to directly install `penguins-chef` using the appropriate system package manager across supported distributions (APT, Pacman, DNF, Zypper).

### ⚡ Performance & Streaming Improvements
* **Throttled Log UI Updates**: Batched UI log updates with an 80ms throttle window and buffered up to 1000 lines to ensure a fluid desktop experience during heavy output.
* **Carriage Return (`\r`) Handling**: Added specialized stream scanner supporting both `\n` and `\r` (including `\r\n`), ensuring live CLI progress bars and terminal updates stream smoothly.

## Installation

### AppImage
```sh
chmod +x penguins-gui-26.10.21-1-x86_64.AppImage
./penguins-gui-26.10.21-1-x86_64.AppImage
```

### Debian / Ubuntu / Devuan (.deb)
```sh
sudo apt install ./penguins-gui_26.10.21-1_amd64.deb
```

### Fedora (.rpm)
```sh
sudo dnf install ./penguins-gui-26.10.21-1.fc44.x86_64.rpm
```

### openSUSE (.rpm)
```sh
sudo zypper install ./penguins-gui-26.10.21-1.opensuse.x86_64.rpm
```

### Arch Linux / Manjaro (.pkg.tar.zst)
```sh
sudo pacman -U ./penguins-gui-26.10.21-1-x86_64.pkg.tar.zst
```

Run Penguins GUI as a normal desktop user. SHA256SUMS is provided alongside the
packages for download verification.
