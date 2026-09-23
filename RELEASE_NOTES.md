# Penguins GUI v26.9.23

A compact desktop interface for creating Penguins' Eggs live images and clones,
with live output and direct access to the resulting ISO folder.

This release adds graphical clone authentication and encrypted-clone configuration,
maintenance menus, an action toolbar with contextual explanations, persistent font
zoom, and Clear/Copy log controls. The Exit button and window-close action keep
the interface open until the current operation finishes.

The Debian amd64 package includes a desktop launcher and icon. It is built on
Debian Bookworm; Penguins' Eggs must already be available from its repository.
Clone password dialogs also require sudo, and graphical encrypted clones require
an Eggs version supporting `EGGS_LUKS_PASSPHRASE`.

Download the `.deb` and install it with:

```sh
sudo apt install ./penguins-gui_26.9.23-1_amd64.deb
```

Run Penguins GUI as a normal desktop user. SHA256SUMS is provided alongside the
package for download verification.

Known limitations: progress is textual, ISO discovery uses file timestamps, and
operation cancellation is not yet available. Encrypted-mode secret handling is
unchanged in this release: in the sudo path, the passphrase appears in the command
arguments and GUI log. Do not share encrypted-mode logs containing secrets.
