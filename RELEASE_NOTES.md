# Penguins GUI v26.10.26

A modern desktop interface for creating Penguins' Eggs live images and clones,
cooking system recipes with Penguins Chef, participating in the Penguins Eggs Community,
managing system installers and setup tools, and booting live ISOs directly from disk without a USB drive.

## What's New in v26.10.26

### 🌍 Penguins Eggs Community Integration
* **Dedicated Community Tab**: Join the global Penguins Eggs Community directly from the GUI to share your public artisan profile (Nickname, Location, Repository/Website URL, Bio up to 200 characters, and optional Social / Fediverse handles).
* **Single-Checkbox Workflow**:
  * **Check `Join the Penguins Eggs Community`** to validate, save (`community.json` & `community.yaml`), and publish your profile to the Community server (or automatically import your existing profile if only the Community Token field is filled).
  * **Uncheck `Join the Penguins Eggs Community`** to remove your public data from the Community server while keeping all your local fields preserved and editable on your machine.
* **Secure Private Token Storage**: Stores your private Community Token (`egg_pat_...`) exclusively in `~/.ssh/penguins-community.secret` (`0600` permissions), never inside `community.json` or `community.yaml`.
* **Privacy & Anti-Spam Protection**: Email addresses are strictly forbidden across all public fields, while Fediverse handles (`@user@instance.social`) are fully supported.

### 🏗️ Modular Architecture Refactoring
* **Clean Package Structure**: Reorganized the codebase into focused internal packages (`internal/ui`, `internal/community`, `internal/eggs`, `internal/chef`, `internal/system`, and `internal/builder`) with comprehensive unit test coverage.

## Installation

### AppImage
```sh
chmod +x penguins-gui-26.10.26-1-x86_64.AppImage
./penguins-gui-26.10.26-1-x86_64.AppImage
```

### Debian / Ubuntu / Devuan (.deb)
```sh
sudo apt install ./penguins-gui_26.10.26-1_amd64.deb
```

### Fedora (.rpm)
```sh
sudo dnf install ./penguins-gui-26.10.26-1.fc44.x86_64.rpm
```

### openSUSE (.rpm)
```sh
sudo zypper install ./penguins-gui-26.10.26-1.opensuse.x86_64.rpm
```

### Arch Linux / Manjaro (.pkg.tar.zst)
```sh
sudo pacman -U ./penguins-gui-26.10.26-1-x86_64.pkg.tar.zst
```

Run Penguins GUI as a normal desktop user. SHA256SUMS is provided alongside the
packages for download verification.
