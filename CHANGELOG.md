# Changelog

## Release Notes: penguins-gui v26.10.31 - 2026-10-31

### 🛠️ Unified Penguins Tools Installation

* **Unified Repository & Tool Setup**: Replaced separate setup actions with **Edit → Install Penguins Tools**, configuring the official Penguins repository and installing both `penguins-eggs` and `penguins-chef` together.
* **Auto-Activation**: Automatically detects newly installed tools and unlocks remaster actions without requiring an application restart.

### 🧹 Streamlined Menus & Actions

* **Removed Calamares Installer Action**: Removed the standalone Calamares installer script and its Edit menu entry to keep the interface focused on core Penguins tooling.

## Release Notes: penguins-gui v26.10.28 - 2026-10-28

### 🖥️ Toolbar & Top Bar Layout

* **Right-Aligned Top Bar Actions**: Moved `Docs`, `Stats`, and `Exit` to the right side of the primary top bar alongside the `Remaster` and `Community` tabs, making them accessible across the entire application.
* **Streamlined Action Toolbar**: Removed the duplicate `Community` button and relocated global links from the secondary action toolbar, keeping it focused on core operations (`Create ISO`, `Kill`, `Clean`, `Cook`, and `Boot ISO`).

### 🌍 Penguins Eggs Community Refinements

* **ON / OFF Toggle Button**: Replaced the community checkbox and status line with a clean `Join the Penguins Eggs Community: ON / OFF` toggle button.
* **Visible Read-Only State when OFF**: When set to `OFF`, saved profile fields remain populated and visible but disabled, while public data is removed from the Community server. When switched to `ON`, fields are unlocked for editing with automatic local persistence and server synchronization.
* **Table-Driven Socials & Unified HTTP Client**: Refactored social channel management and unified HTTP request/fallback handling in `internal/community`.

### 🍳 Penguins Chef CLI Update

* **Updated Chef Commands**: Replaced `chef get` and `chef apply` with `chef cook --pull` and `chef cook` to match the latest `penguins-chef` CLI.

## Release Notes: penguins-gui v26.10.26 - 2026-10-26

### 🌍 Penguins Eggs Community Integration

* **Dedicated Community Tab**: Join the global Penguins Eggs Community directly from the GUI to share your public artisan profile (Nickname, Location, Repository/Website URL, Bio up to 200 characters, and optional Social / Fediverse handles).
* **Single-Checkbox Workflow**: Checking `Join the Penguins Eggs Community` validates, saves (`community.json` & `community.yaml`), and publishes your profile to the Community server (or imports an existing profile if only the Community Token is provided). Unchecking removes your public data from the server while preserving your local fields for editing.
* **Secure Private Token Storage**: Stores your private Community Token (`egg_pat_...`) exclusively in `~/.ssh/penguins-community.secret` (`0600` permissions), never inside `community.json` or `community.yaml`.
* **Privacy & Anti-Spam Protection**: Strict validation prevents email addresses in public fields while supporting Fediverse handles (`@user@instance.social`).

### 🏗️ Modular Architecture Refactoring

* **Clean Package Structure**: Reorganized the codebase into focused internal packages (`internal/ui`, `internal/community`, `internal/eggs`, `internal/chef`, `internal/system`, and `internal/builder`).

## Release Notes: penguins-gui v26.10.21 - 2026-10-21

### 🖥️ Layout & Interface Improvements

* **Refactored Mode Selector**: Converted remaster mode options to a sleek horizontal layout with subtle separators and compact ellipsis descriptions.
* **Maximized Operation Log Area**: Reorganized header layout to maximize the visible operation log area, placing status bar and mode row neatly above a dedicated log control bar.

### 🍳 Penguins Chef Management

* **Install Penguins Chef**: Added an action in the Edit menu to directly install `penguins-chef` using the appropriate system package manager across supported distributions (APT, Pacman, DNF, Zypper).

### ⚡ Performance & Streaming Improvements

* **Throttled Log UI Updates**: Batched UI log updates with an 80ms throttle window and buffered up to 1000 lines to ensure a fluid desktop experience during heavy output.
* **Carriage Return (`\r`) Handling**: Added specialized stream scanner supporting both `\n` and `\r` (including `\r\n`), ensuring live CLI progress bars and terminal updates stream smoothly.

## Release Notes: penguins-gui v26.10.11 - 2026-10-11

### 🍳 Penguins Chef Integration (Cook)

* **Interactive Recipe Picker**: Added the "Cook…" action in the action toolbar and Edit menu, opening an interactive browser for Penguins Chef recipes.
* **Full-Window Expanded View**: The recipe picker automatically expands to the edges of the parent window canvas, featuring window-style controls (minimize to compact size, maximize/restore, and close).
* **Live Recipe Preview & Metadata**: Real-time YAML definition preview in a scrollable TextGrid, complete with recipe descriptions, file paths, and supported distribution profiles.
* **Apply & Dry-Run Support**: Directly apply recipes or run `--dry-run` previews with live terminal output and execution logging.
* **Automatic Recipes Update**: Update or initialize the local `~/.chef` recipe repository with a single click.
* **Filtered Categories**: Excluded internal low-level recipe categories (`base`, `de`, `dm`, `costumes`) from the GUI for an intuitive user experience.

### 🚀 Boot ISO Clarification

* **Boot ISO Action**: Renamed the bootloader configuration action from "Boot without USB…" to "Boot ISO…" across toolbar buttons and menus for clearer user intent.

### 📦 Packaging Updates

* **Arch Linux / Manjaro**: Added `xorg-xrandr` dependency in packaging templates for full resolution and display handling.

## Release Notes: penguins-gui v26.10.8 - 2026-10-08

### 📦 Multi-Distribution Packaging & AppImage

* **AppImage Packaging**: Added standalone AppImage build support via `make appimage` and integration in the Hammers packaging CI, publishing direct `.AppImage` release binaries.
* **openSUSE RPM Packaging**: Added native RPM packaging for openSUSE Leap, Tumbleweed, and Slowroll, package installation via `./m`, and a dedicated openSUSE Tumbleweed CI job with release artifacts.
* **Manjaro & Arch Packaging**: Added native Arch Linux packaging (`.pkg.tar.zst`), Manjaro packaging, and CI builds with release artifacts.
* **Fedora RPM Packaging**: Added native RPM generation for Fedora, `./m` installer integration, and a Fedora 44 CI workflow producing release artifacts.
* **Release Artifacts**: All distribution packages (`.deb`, `.pkg.tar.zst`, `.rpm`, `.AppImage`) and SHA256SUMS are now automatically built and attached to GitHub releases.

### 🐧 System Setup, Tools & Calamares

* **Native Repositories & Eggs Setup**: Added independent actions to configure or remove native Penguins' Eggs repositories without requiring the CLI, plus automatic Eggs installation.
* **First-Run Experience**: Made Eggs optional in desktop packages, adding guidance and detection when Eggs is not yet installed.
* **Calamares Installation**: Added a dedicated tool action to install Calamares and matching Qt 5 / Qt 6 slideshow dependencies across supported distributions with administrative elevation and live logs.
* **Menu Organization**: Streamlined menu actions to provide direct access to setup tools (Eggs CLI, Calamares, /etc/skel, grub40).

### 🚀 Boot Without USB & GRUB Integration

* **Boot Without USB**: Renamed grub40 integration to "Boot without USB…" to easily configure GRUB to boot created ISOs directly from disk.
* **Safe GRUB Management**: Marked Eggs blocks are cleanly replaced with the selected ISO, generating non-executable backups to prevent stale duplicate entries.
* **GRUB Detection & Reminders**: Automatically identifies GRUB availability on startup, disables the action with an explanation if unavailable, and reminds users to run `sudo update-grub` after configuration.

## Release Notes: penguins-gui v26.9.23 - 2026-09-23

### 🐧 Desktop Interface & Menus

* **Log Actions**: Added compact "Clear" and "Copy" buttons beside the log heading and corresponding Edit menu actions, with inline feedback. Clearing the log leaves the running operation active and continues displaying new output.
* **Dialog Icons**: Added icons to authentication, encryption, boot guide and About dialogs, complementing the native information, error and confirmation icons.
* **Exit & Layout Polish**: Added an Exit button at the right of the toolbar, highlighted Create ISO and added consistent window padding. Toolbar, menu and window-close actions keep the GUI open while an operation is running.
* **File Menu**: Added "Create ISO" to the File menu, replicating the main interface button and synchronizing its availability and busy state.
* **Application Menus**: Added Edit (compact "Kill…" command for `eggs kill`), Tools (`clean`, `grub40`, `repo`, `skel`), and Help (booting/testing guide, documentation link, and about dialog).
* **PolicyKit Rules & Window Icon**: Added `49-penguins-gui.rules` for privileged execution and customized window branding.
* **GUI Authentication & Encryption Dialogs**: Replaced terminal prompts with native GUI dialogs for administrator authentication during system clones and LUKS passphrase configuration for encrypted clones.
* **View Menu & Font Zoom**: Added a View menu and keyboard shortcuts (`Ctrl +`, `Ctrl -`, `Ctrl 0`) with presets (85% to 200%) to scale UI fonts dynamically and persist the choice in user preferences. The initial window size now scales proportionally.
* **Action Toolbar & Interactive Explanations**: Added a quick action toolbar (Create ISO, Kill, Clean, GRUB, Docs) with an interactive explanation banner that displays detailed command information on hover and upon selection from menus or toolbar buttons.
* **About Dialog Metadata**: Updated Help -> About to display the penguins-gui application version and credit author Piero Proietti <piero.proietti@gmail.com>.
* **Streamlined Main Window Layout**: Cleaned up the main window by removing redundant header text, Eggs version string, working directory selector (defaulting to /home/eggs), and duplicate inline Create ISO button, delegating primary actions to the toolbar and menus.

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
