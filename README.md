# penguins-gui

Minimal, independent desktop GUI for Penguins' Eggs.

The prototype deliberately does one thing: it starts an Eggs remaster, displays
the live output and locates the ISO created at the end. It does not import Eggs
internals and does not modify Penguins' Eggs.

## Current scope

- Detect `eggs` in `PATH` and show its version.
- Select one of the three existing remaster modes:
  - standard live;
  - system clone (`--clone`);
  - encrypted clone (`--crypted`).
- Select the Eggs working directory (`--path`).
- Request administrative authorization through `pkexec`/polkit.
- Display stdout and stderr while Eggs runs in standard and clone modes.
- Open Eggs' existing interactive TUI in a terminal for encrypted mode.
- Find the newest ISO created beneath the selected working directory.
- Open the ISO directory with `xdg-open`.

Tailor, Krill, AI assistance, configuration editing and artifact management are
intentionally outside this first prototype.

## Requirements

- Linux desktop
- Go 1.25 or later
- Penguins' Eggs installed and available as `eggs`
- polkit with `pkexec` when the GUI is not already running as root
- Fyne build dependencies for the distribution

See the official Fyne documentation for the development packages required by
your distribution: <https://docs.fyne.io/started/>.

## Run during development

```bash
go mod tidy
go run .
```

## Test and build

```bash
go test ./...
go build -o penguins-gui .
```

The same operations are available as `make test`, `make build` and `make run`.

Do not launch the whole GUI with `sudo` for ordinary use. The prototype asks
polkit to authorize only the `eggs` process.

## Known prototype limitations

- Eggs currently emits textual output, so the GUI shows a live log but not a
  reliable percentage progress bar.
- Encrypted remastering still requires Eggs' interactive TUI for the passphrase
  and crypto parameters. The prototype opens that flow in a separate terminal;
  its detailed log therefore remains in that terminal.
- The final artifact is discovered by scanning for a recently modified `.iso`
  below the selected working directory.
- Closing the window does not yet provide a supervised cancellation workflow.
- The command shown in the log is informational and does not shell-escape paths;
  execution itself uses Go's argument-safe `exec.Command`, not a shell.
- Packaging and desktop integration are currently provided only for Debian.

## Architectural rule

```text
penguins-gui -> eggs -> coa -> oa
```

Eggs remains fully usable without the GUI, and the GUI only consumes the public
command-line interface.

## Debian package

On Colibri/Debian, run as a normal user (no sudo):

```bash
make build
make package
```

Packaging requires Git, `dpkg-dev` (including `dpkg-shlibdeps`) and the Fyne
build dependencies. `make package` rebuilds the GUI and writes a native `.deb`
to `dist/`. Cross-compilation is not supported by this packaging workflow.
The standalone Go builder uses only the standard library and has no dependency
on Penguins Tailor. Its staging/metadata/archive design and Git version rules
are adapted from Tailor's `pkg/builder`.

The package installs `/usr/bin/penguins-gui`, its desktop launcher and a scalable
SVG icon. Shared-library dependencies are calculated from the compiled binary
with `dpkg-shlibdeps`; `penguins-eggs`, `pkexec` and `xdg-utils` are also required
for the commands the GUI invokes. Encrypted mode additionally needs one of the
supported graphical terminals listed in `runInteractiveTerminal`.

Versioning follows Tailor: the nearest Git tag loses its leading `v`, and `-`
and `_` become dots. The Debian revision is the number of commits reachable
from HEAD but not from any tag, with zero replaced by one. Without tags, the
base version is `0.1.0` and the revision is the total HEAD commit count (or one
if Git cannot supply it). Uncommitted changes do not alter the version.

Inspect the archive before installing it:

```bash
dpkg-deb --info dist/*.deb
dpkg-deb --contents dist/*.deb
```

`make clean` removes the built binary and `dist/`. There is no `make install`;
installation is left to Debian's package manager, separately from building.
