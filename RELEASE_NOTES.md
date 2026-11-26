# Penguins GUI v26.11.26-1

A modern desktop interface for creating Penguins' Eggs live images and clones,
cooking system recipes with Penguins Chef, participating in the Penguins Eggs Community,
managing system installers and setup tools, and booting live ISOs directly from disk without a USB drive.

## What's New in v26.11.26-1

### 🛠️ UI & Menu Refinements
* **Clearer Repository & Suite Setup Action**: Renamed **Install Penguins Tools** to **Install Repo, Eggs & Chef…** in the Edit menu and updated missing CLI prompts to clearly describe repository configuration and suite installation.
* **Updated Screenshots**: Refreshed `remaster.png` and `community.png`.

### 📦 Packaging & AppImage Fixes
* **AppImage `.DirIcon` Validation**: Supplied a 256x256 PNG icon for AppImage `.DirIcon` validation.
* **RPM Spec Packaging**: Packaged the PNG icon and added a `%changelog` section in the RPM spec template.

## Installation

### AppImage
```sh
chmod +x penguins-gui-26.11.26.1-1-x86_64.AppImage
./penguins-gui-26.11.26.1-1-x86_64.AppImage
```

### Debian / Ubuntu / Devuan (.deb)
```sh
sudo apt install ./penguins-gui_26.11.26.1-1_amd64.deb
```

### Fedora (.rpm)
```sh
sudo dnf install ./penguins-gui-26.11.26.1-1.fc44.x86_64.rpm
```

### openSUSE (.rpm)
```sh
sudo zypper install ./penguins-gui-26.11.26.1-1.opensuse.x86_64.rpm
```

### Arch Linux / Manjaro (.pkg.tar.zst)
```sh
sudo pacman -U ./penguins-gui-26.11.26.1-1-x86_64.pkg.tar.zst
```

Run Penguins GUI as a normal desktop user. SHA256SUMS is provided alongside the
packages for download verification.
