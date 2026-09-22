# Changelog

## [Unreleased]

### 🐧 Desktop Interface & Menus

* **File Menu**: Added "Create ISO" to the File menu, replicating the main interface button and synchronizing its availability and busy state.
* **Application Menus**: Added Edit (delete previous ISOs via `eggs kill`), Tools (`clean`, `grub40`, `repo`, `skel`), and Help (booting/testing guide, documentation link, and about dialog).
* **PolicyKit Rules & Window Icon**: Added `49-penguins-gui.rules` for privileged execution and customized window branding.

### 📦 Packaging & CI

* **Unified Debian Package**: Simplified the Hammers CI workflow to build in a Debian Bookworm container, creating a single `penguins-gui-debian-amd64` artifact compatible with Debian (Bookworm and Trixie), Devuan, and Ubuntu.

## Release Notes: penguins-gui v26.9.22 - 2026-09-22

This first release introduces an independent Fyne desktop interface for Penguins'
Eggs, with remaster controls, live logs, ISO discovery and native Debian packaging.

### 🐧 Eggs Remaster Interface

* **Eggs Detection**: Locates `eggs` in `PATH`, displays its version and disables remastering when the executable is unavailable.
* **Three Remaster Modes**: Supports standard live images, system clones (`--clone`) and encrypted clones (`--crypted`), with a selectable working directory defaulting to `/home/eggs`.
* **Per-Process Authorization**: Uses `pkexec` to elevate the Eggs command when the GUI runs as a normal user.
* **Interactive Encrypted Mode**: Opens Eggs' encrypted wizard in a supported graphical terminal, leaving passphrase entry to Eggs.

### 🖥️ Logs and Completion Feedback

* **Live Command Output**: Displays stdout and stderr during standard and clone remasters, stripping ANSI escape sequences for readable logs.
* **Completion Dialog**: Shows a popup after the command exits without error, including the ISO result or a warning when no new ISO is found.
* **ISO Discovery Fix**: Searches only directly inside the selected working directory, avoiding permission failures in Eggs' working subdirectories. Selects the newest qualifying ISO with a two-second timestamp tolerance.
* **Artifact Access**: Displays the ISO path and size and provides an action to open its containing folder.

### 🧩 CLI Adapter and Tests

* **Isolated Eggs Integration**: Moves executable detection, argument construction, privileged execution, terminal launching and ISO discovery into `internal/tools/eggs.CLIAdapter`, preserving the existing GUI flow.
* **Regression Coverage**: Retains tests for remaster arguments, ISO discovery, size formatting and ANSI removal. Updates ISO discovery coverage to verify that nested images are ignored.

### 📦 Debian Packaging and Hammers CI

* **Independent Package Builder**: Adds `make package` and a standalone Go builder that writes native `.deb` archives to `dist/` without administrative privileges.
* **Desktop Integration**: Packages the GUI executable, desktop launcher and SVG icon, with shared-library dependencies calculated through `dpkg-shlibdeps`.
* **Git-Based Versions**: Derives package versions from Git tags and commit history, with tests for version calculation.
* **Hammers Workflow**: Adds GitHub Actions jobs for Debian Bookworm and Trixie on amd64, triggered by pushes and pull requests to `main` or manually. Runs tests with a virtual display, builds as an unprivileged user, inspects archives and uploads artifacts with seven-day retention.

### 📝 Architecture and Protocol Documentation

* **Project Direction**: Adds vision, architecture and roadmap documents describing the separation between the GUI and external Penguins tools.
* **Protocol Proposal**: Adds the machine-readable protocol specification, JSON Schemas and example fixtures for future integration. This release continues to use Eggs' existing textual CLI.
