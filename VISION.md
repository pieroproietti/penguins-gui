# Penguins GUI: Vision and Scope

## Purpose

Penguins GUI is an independent graphical application for the Penguins ecosystem.
Its purpose is to make existing command-line tools easier to discover, configure,
and use without replacing them or duplicating their internal logic.

The project begins with a deliberately small goal:

> Start a Penguins' Eggs remaster from a desktop window, follow its output, and
> locate the resulting ISO.

From that foundation, Penguins GUI may grow into a common graphical workspace
for preparing a Linux system, remastering it, inspecting the resulting artifacts,
and launching the appropriate installer.

## Core principles

### Independent tools

Penguins GUI, Penguins' Eggs, and Krill must remain independent programs.
(Penguins Tailor remains a dedicated CLI/TUI tool).

Each tool owns one clear responsibility:

- **Penguins Tailor** prepares and customizes a naked/headless system from the console (CLI/TUI). Because it runs before any desktop or display server is installed, it is intentionally excluded from Penguins GUI.
- **Penguins' Eggs** remasters the running system into a live, bootable image.
- **Krill** installs a live system onto storage.
- **Penguins GUI** discovers, presents, and coordinates the desktop-ready tools (primarily Eggs and later Krill).

The dependency direction is always from the GUI toward the tools:

```text
penguins-gui ---> eggs ---> coa ---> oa
             \
              +--> krill
```

Eggs and Krill must never require Penguins GUI. Removing the GUI must
not affect their command-line operation.

### Public interfaces, not internal coupling

Penguins GUI should use the public command-line interfaces and stable,
versioned data formats exposed by the other tools. It should not import their
internal packages or reproduce their business logic.

When richer integration is needed, it should first improve the corresponding
CLI through generally useful capabilities such as structured status output,
machine-readable events, dry-run support, or versioned manifests. The GUI then
becomes one consumer of those interfaces.

### Graphical interface first

The normal user experience is a conventional desktop GUI with clear controls,
summaries, progress information, and logs. A conversational interface is
optional and must not be required to operate the application.

### Transparent operation

The GUI must not turn remastering or system customization into a black box.
Before an operation starts, the user should be able to understand what will be
executed. During execution, progress and logs should remain visible. On failure,
the application should expose the relevant command, exit status, logs, and
workspace state.

### Safe privilege boundaries

The GUI runs as the normal desktop user. Privileged operations are delegated
only when required, using an explicit authorization mechanism such as polkit.
The complete graphical application should not normally run as root.

Destructive or system-wide actions require a traditional, unambiguous user
confirmation. Neither the GUI nor an AI assistant may silently format storage,
change repositories, install packages, or delete workspaces.

### Incremental development

Penguins GUI should grow through small, useful releases. A feature is added only
after the previous layer works reliably on real systems. Supporting every tool,
distribution, and workflow in the first release is explicitly not a goal.

## Current scope

The initial implementation is a graphical front end for `eggs remaster`.

It currently aims to:

- detect an installed `eggs` executable and display its version;
- offer standard live, system clone, and encrypted clone modes;
- allow selection of the Eggs working directory;
- request authorization for the privileged Eggs process;
- display remaster output in real time;
- report success or failure clearly;
- locate the ISO created during the current operation;
- open the directory containing the resulting ISO.

Encrypted remastering currently retains the interactive Eggs workflow and may
open it in a separate terminal. Penguins GUI should not duplicate encryption
logic or collect secrets unless Eggs first provides an appropriate, stable,
and secure interface.

## Product direction

### Remaster workspace

The remaster view will remain the central feature. It may progressively add:

- preflight checks for compatibility, dependencies, workspace state, and free
  space;
- compression settings already supported by Eggs;
- structured progress for the stages in the Eggs flight plan;
- artifact metadata, checksums, and access to the completed ISO;
- recovery actions after an interrupted or failed remaster;
- clear separation between observed facts, warnings, and suggested actions.

### Tailor (out of scope)

Penguins Tailor is intentionally not integrated into Penguins GUI. Tailor is
designed to dress and configure naked Linux systems (often lacking Xorg/Wayland
and any desktop environment). Since a graphical interface cannot run on a naked
system, Tailor belongs naturally and exclusively to the terminal (CLI/TUI). Once
a desktop is installed and running, the tailoring phase is already complete.

### Installation tools

Penguins GUI may detect and launch available installers such as Calamares or
Krill. It should not reimplement their partitioning and installation logic.

Future Krill integration may present supported modes, including Coexist, only
through capabilities explicitly exposed by Krill. Any eventual separation of
Krill from Penguins' Eggs is an independent project decision and is not a
prerequisite for developing Penguins GUI.

### Artifact management

A later artifact view may list generated ISOs, native packages, checksums,
plans, and logs. It may also invoke existing export or GRUB integration commands
provided by the relevant tools.

Penguins GUI should not invent a second artifact format or silently delete old
outputs.

## AI assistance

AI is an optional advisor and preparer, not the primary interface and not the
execution engine.

When enabled, AI assistance may:

- analyze the detected system and available tool capabilities;
- recommend a compatible costume or remaster configuration;
- pre-fill ordinary GUI controls;
- explain warnings and failures using the current logs and system state;
- prepare a proposed configuration or change for review;
- answer optional conversational questions.

The default workflow remains graphical and does not require a chat.

AI output must be treated as a proposal. It must be converted into a limited,
structured form, validated by deterministic code, displayed to the user, and
explicitly approved before execution.

```text
user intent
    |
    v
AI proposal (optional)
    |
    v
structured GUI configuration
    |
    v
deterministic validation
    |
    v
user approval
    |
    v
Eggs / Krill
```

The AI must not receive an unrestricted root shell or independently perform
destructive actions.

## Native distribution packages

Penguins GUI should be distributed as native packages while keeping the source
project and build process independent from the other Penguins tools.

Packaging may follow the proven builder structure used by Penguins Tailor,
adapted for a graphical Fyne application. Packages should install at least:

```text
/usr/bin/penguins-gui
/usr/share/applications/penguins-gui.desktop
/usr/share/icons/hicolor/scalable/apps/penguins-gui.svg
```

Debian packaging is the first target. Other native formats should be added only
after installation, desktop integration, runtime dependencies, upgrades, and
removal have been verified on the corresponding distribution families.

Build-time Fyne dependencies must be distinguished from runtime dependencies.
Penguins' Eggs may be a package dependency or an explicitly declared required
companion, while Krill and AI support remain optional integrations.

## Non-goals

Penguins GUI is not intended to:

- replace the command-line interfaces of Eggs or Krill;
- support Penguins Tailor, which is designed exclusively for naked console environments;
- merge the Penguins projects into a single monolithic application;
- maintain a second implementation of remastering, customization, encryption,
  partitioning, or installation logic;
- hide commands, logs, errors, or system changes from the user;
- require AI or a chat interface for normal operation;
- grant unrestricted administrative access to an AI model;
- become a build-from-spec distribution pipeline;
- claim support for a distribution before the complete workflow has been
  tested there.

## Development milestones

### Milestone 1: useful remaster GUI

- Reliable standard and clone remaster execution.
- Readable real-time logs.
- Correct success and failure reporting.
- Discovery of the newly generated ISO.

### Milestone 2: desktop distribution

- Native Debian package.
- Desktop launcher and project icon.
- Verified installation, upgrade, and removal.
- Documented build and runtime dependencies.

### Milestone 3: stronger Eggs integration

- Preflight checks.
- Structured status and progress where supported by Eggs.
- Better artifact and failure handling.
- Secure integration of interactive modes.

### Milestone 4: installer integration (Krill)

- Discovery of Krill and supported installation modes.
- Storage and partition plan inspection.
- Supervised execution and progress reporting.

### Milestone 5: advisory AI

- Optional system-aware recommendations.
- Structured proposals that pre-fill the GUI.
- Explanations grounded in actual tool output and system state.
- Optional chat for users who want it.

### Milestone 6: wider ecosystem

- Installer launching and Krill capability discovery.
- Artifact export and related utility integration.
- Native packages and verified behavior on additional distributions.

These milestones describe direction, not a promise of release order or dates.
Each stage should remain useful even if later stages are never implemented.

## Success criteria

Penguins GUI succeeds when it makes the Penguins workflow easier to understand
and operate while preserving the qualities of the underlying tools:

- independence;
- transparency;
- scriptability;
- distribution awareness;
- user control.

The long-term objective can be summarized as:

> Penguins GUI helps the user prepare the system, asks the existing Penguins
> tools to perform their specialized work, and makes every important decision
> and result visible.
