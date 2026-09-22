# Roadmap

This roadmap turns the project vision into small, independently verifiable
milestones. It deliberately avoids release dates. Priorities may change after
real use, but a milestone is complete only when its acceptance criteria pass.

## M0 — Remaster proof of concept

**Status:** completed.

The GUI can detect Eggs, start standard and clone remasters, open encrypted
mode in a terminal, display live textual output and locate a newly created ISO.

Acceptance criteria already demonstrated:

- the application runs as a normal desktop user;
- only the external command is elevated;
- Eggs remains unchanged and independently usable;
- a real remaster completes from the Fyne window;
- the resulting ISO can be located and its directory opened.

## M1 — Stabilize the current GUI

Goal: make the existing textual integration dependable before expanding scope.

- extract process execution and Eggs command construction from Fyne widgets;
- model a remaster operation with explicit states;
- retain a bounded log and make technical details copyable;
- handle authorization denial and missing dependencies clearly;
- add graceful cancellation where the current Eggs process permits it;
- improve result and error presentation;
- keep unit tests for mode mapping, process selection and artifact fallback.

**Complete when:** the current three modes behave consistently across success,
failure, denied authorization and user cancellation, without freezing the UI.

## M2 — Eggs protocol foundation

Goal: replace interpretation of human output with a stable machine interface.

- agree on `penguins/v1` request and event envelopes;
- add protocol JSON Schemas and shared fixtures;
- add Eggs capability discovery;
- add a non-destructive remaster check/plan;
- emit NDJSON events from `eggs remaster`;
- emit one terminal result containing the exact ISO artifact;
- emit structured terminal errors;
- preserve the normal human CLI as the default.

**Complete when:** the GUI can run a standard remaster without parsing log text
or scanning a directory to identify the ISO.

## M3 — Protocol-native remaster experience

Goal: use Eggs data to present a trustworthy remaster workflow.

- show detected system, architecture and readiness;
- build controls from Eggs capabilities;
- show normalized options, space estimates and warnings before execution;
- render named steps and real progress when available;
- support protocol cancellation and non-interruptible step indications;
- retain the legacy adapter for older Eggs versions;
- design a protected flow for encrypted-mode secrets.

**Complete when:** standard, clone and crypted modes use structured data on a
supporting Eggs version and degrade clearly on older versions.

## Note on Penguins Tailor

Tailor integration is explicitly out of scope for `penguins-gui`. Tailor is
designed to dress and configure naked systems from the console (CLI/TUI) before
any display server or graphical interface is installed. A desktop GUI cannot run
in that environment, and once a desktop is present, the tailoring step has
already passed.

## M4 — Workflow composition and artifact management

Goal: coordinate tools and artifacts without turning them into one coupled application.

- persist only GUI preferences and resumable workflow metadata;
- represent remastering and future installation as separate operations and results;
- detect changes that invalidate an earlier check or plan;
- display artifacts, ISOs, checksums and reports together;
- never assume that Eggs and Krill share installation or versions.

**Complete when:** workflows can combine independently versioned tools and manage
resulting artifacts entirely through their public contracts.

## M5 — Independent Krill integration

Goal: integrate installation only after Krill has a stable independent CLI.

- discover capabilities and supported installation modes;
- retrieve safe disk inventory;
- create and validate an immutable installation plan;
- present destructive changes prominently;
- require immediate explicit confirmation before execution;
- render NDJSON progress and the final installation report;
- communicate cancellation limits and reboot requirements.

**Complete when:** the GUI can install from a validated plan without importing
Krill internals or weakening its own safeguards.

## M6 — Optional adviser

Goal: add assistance without changing the default GUI-first experience.

- derive deterministic recommendations from capabilities and checks first;
- let an optional AI adviser explain choices, warnings and failures;
- allow it to prepare a draft request;
- validate every draft through the owning tool;
- show every material choice before execution;
- prohibit AI confirmation of destructive or privileged actions.

**Complete when:** advice improves understanding but the same workflows remain
fully usable without AI or network access.

## Continuous work

These concerns apply to every milestone:

- native packages and desktop integration;
- accessibility and readable themes;
- Italian and English localization without using translated messages as logic;
- protocol compatibility tests;
- clear release notes and supported-version information;
- strict separation between GUI, Eggs and Krill.

## Immediate next steps

The next development sessions should remain small:

1. introduce the operation state type and tests in `penguins-gui`;
2. extract the existing Eggs legacy adapter without changing behaviour;
3. add schema validation to protocol fixtures;
4. implement the common event writer in Eggs behind an explicit structured
   output option;
5. return the final ISO path from Eggs in a terminal result event.

