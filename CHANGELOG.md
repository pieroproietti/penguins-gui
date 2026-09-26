# Changelog

## Unreleased

- Add native openSUSE RPM packaging for Leap, Tumbleweed and Slowroll, zypper installation through `./m`, and a Tumbleweed CI artifact included in releases.

## Unreleased

* **GRUB duplicate entries**: Save the custom-menu backup without execute permissions so GRUB does not load its old ISO entries. Replacing an existing backup also removes its execute permissions.

* **Single ISO boot entry**: Boot without USB replaces previous marked Eggs blocks with the selected ISO, preserving unrelated custom entries and saving a backup. Generate the new entry before writing and reject incomplete markers. ISO cleanup remains a manual Kill action.

* **GRUB detection**: Check kernel/EFI boot information at startup and disable Boot without USB in the toolbar and menu when GRUB cannot be identified, with an explanation on hover. Keep this availability check after other operations finish.

* **GRUB reminder**: Boot without USB now reminds users to update GRUB before restarting, both in the confirmation and after a successful operation, with `sudo update-grub` guidance for Debian/Ubuntu.

* **Boot without USB**: Renamed the grub40 button and menu entry to “Boot without USB…” and clarified that the action adds the ISO to this computer's startup menu for a later restart.

* **Edit menu**: Replaced its previous entries with the four setup actions and removed the Tools menu. Updated first-run guidance to point to Edit.

* **grub40 button**: Renamed GRUB to grub40, with hover and click explanations describing ISO booting from disk without a USB drive.

* **Tools menu**: Reduced to four concise entries: Install penguins-egg CLI, Install calamares, Update /etc/skel, Configure grub40.

* **CLI setup entry**: Renamed the repository-and-install action to “Install penguins-eggs CLI” and added a startup warning directing users there when `eggs` is not found.

* **Native repositories**: Added independent add/remove actions, using the native repository locations from Eggs without requiring the Eggs CLI.
* **First-run setup**: Made Eggs optional in Debian, Arch and Fedora packages; automatically install Eggs after adding its native repository, with missing-Eggs guidance and automatic detection after setup.
* **Calamares installation**: Added a Tools action to install Calamares and matching Qt 5/Qt 6 slideshow dependencies with native package managers, administrative authorization, live logs and error handling. Added simulated installation tests without modifying host packages.
* **Fedora packaging**: Added native RPM generation, Fedora detection, installation through `./m`, and a Fedora 44 CI job with release artifacts and checksums.

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
