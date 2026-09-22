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
- Packaging and desktop integration are not included yet.

## Architectural rule

```text
penguins-gui -> eggs -> coa -> oa
```

Eggs remains fully usable without the GUI, and the GUI only consumes the public
command-line interface.
