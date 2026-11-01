# Penguins GUI v26.11.1

A modern desktop interface for creating Penguins' Eggs live images and clones,
cooking system recipes with Penguins Chef, participating in the Penguins Eggs Community,
managing system installers and setup tools, and booting live ISOs directly from disk without a USB drive.

## What's New in v26.11.1

### 🌍 Simplified Community Workflow & In-Memory Token Security
* **Gravatar & SHA-256 Identity**: Simplified community identification using the SHA-256 hash of the artisan's email address, displaying the user's Gravatar avatar preview while ensuring raw email is not used as the public identifier.
* **In-Memory Token (`egg-pat`) Security**: The private access token is generated on first connection (Trust-On-First-Use) and managed strictly in volatile memory. It is never persisted to disk or configuration files.
* **Modal Token Backup**: A new dialog displays the generated token with one-click copy options for easy storage in password managers.
* **Authenticated Machine Deactivation**: Disconnecting and removing a machine from the community server now securely requires the author's `egg-pat` token.
* **Remaster Integration**: Live remastering operations automatically confirm and log community participation status on ISO creation.

### 👤 Robust User Environment Resolution
* **Target User Resolution**: Added comprehensive user and home directory resolution in `internal/system` (`SUDO_USER`, `PKEXEC_UID`, `logname`, `/etc/passwd`) to ensure application preferences and tools target the real invoking user when running elevated.
* **File Ownership Management**: Automatically enforces human-user ownership on configuration directories (`~/.config/penguins-eggs`, `~/.config/fyne`, and `~/.chef`).

### 🌐 Localization & UI Cleanup
* **English Localization**: Standardized all UI prompts, status labels, confirmation dialogs, and error messages in English.
* **Focused Remaster View**: Cleaned up the Remaster mode selection row by removing redundant community shortcuts.

## Installation

### AppImage
```sh
chmod +x penguins-gui-26.11.1-1-x86_64.AppImage
./penguins-gui-26.11.1-1-x86_64.AppImage
```

### Debian / Ubuntu / Devuan (.deb)
```sh
sudo apt install ./penguins-gui_26.11.1-1_amd64.deb
```

### Fedora (.rpm)
```sh
sudo dnf install ./penguins-gui-26.11.1-1.fc44.x86_64.rpm
```

### openSUSE (.rpm)
```sh
sudo zypper install ./penguins-gui-26.11.1-1.opensuse.x86_64.rpm
```

### Arch Linux / Manjaro (.pkg.tar.zst)
```sh
sudo pacman -U ./penguins-gui-26.11.1-1-x86_64.pkg.tar.zst
```

Run Penguins GUI as a normal desktop user. SHA256SUMS is provided alongside the
packages for download verification.
