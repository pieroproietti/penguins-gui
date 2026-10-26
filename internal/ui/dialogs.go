package ui

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func needsSudoPassword() bool {
	if os.Geteuid() == 0 {
		return false
	}
	return exec.Command("sudo", "-n", "true").Run() != nil
}

func openDirectory(path string) error {
	launcher, err := exec.LookPath("xdg-open")
	if err != nil {
		return errors.New("xdg-open not found")
	}
	return exec.Command(launcher, path).Start()
}

func showCloneAuthDialog(parent fyne.Window, onConfirm func(adminPassword string)) {
	pwEntry := widget.NewPasswordEntry()
	pwEntry.PlaceHolder = "Administrator password"
	items := []*widget.FormItem{
		widget.NewFormItem("", widget.NewIcon(theme.LoginIcon())),
		widget.NewFormItem("Password", pwEntry),
	}
	d := dialog.NewForm("Authentication Required", "Authenticate", "Cancel", items, func(ok bool) {
		if !ok {
			return
		}
		if strings.TrimSpace(pwEntry.Text) == "" {
			dialog.ShowError(errors.New("password cannot be empty"), parent)
			return
		}
		onConfirm(pwEntry.Text)
	}, parent)
	d.Resize(fyne.NewSize(380, 160))
	d.Show()
}

func showEncryptedCloneDialog(parent fyne.Window, requireAdminPw bool, onConfirm func(adminPassword, luksPassphrase string)) {
	luksEntry := widget.NewPasswordEntry()
	luksEntry.PlaceHolder = "Leave empty for default ('evolution')"
	luksConfirm := widget.NewPasswordEntry()
	luksConfirm.PlaceHolder = "Confirm LUKS passphrase"

	items := []*widget.FormItem{
		widget.NewFormItem("", widget.NewIcon(theme.AccountIcon())),
		widget.NewFormItem("LUKS Passphrase", luksEntry),
		widget.NewFormItem("Confirm Passphrase", luksConfirm),
	}

	var adminEntry *widget.Entry
	if requireAdminPw {
		adminEntry = widget.NewPasswordEntry()
		adminEntry.PlaceHolder = "Administrator password"
		items = append(items, widget.NewFormItem("Admin Password", adminEntry))
	}

	d := dialog.NewForm("Encrypted Clone Configuration", "Create ISO", "Cancel", items, func(ok bool) {
		if !ok {
			return
		}
		if luksEntry.Text != luksConfirm.Text {
			dialog.ShowError(errors.New("LUKS passphrases do not match"), parent)
			return
		}
		adminPw := ""
		if adminEntry != nil {
			if strings.TrimSpace(adminEntry.Text) == "" {
				dialog.ShowError(errors.New("administrator password cannot be empty"), parent)
				return
			}
			adminPw = adminEntry.Text
		}
		pass := luksEntry.Text
		if pass == "" {
			pass = "evolution"
		}
		onConfirm(adminPw, pass)
	}, parent)
	d.Resize(fyne.NewSize(420, 220))
	d.Show()
}

func showBootGuideDialog(parent fyne.Window) {
	markdown := `### How to Boot and Test Your ISO

#### 1. Virtual Machines (Quickest & Safest)
Test your ISO immediately without rebooting your physical computer:
* **GNOME Boxes** or **virt-manager**: create a new virtual machine and select the ISO file.
* **VirtualBox**: create an OS machine and mount the ISO in the virtual optical drive.
* **QEMU CLI**:
` + "```bash" + `
qemu-system-x86_64 -enable-kvm -m 4G -cdrom /path/to/egg.iso
` + "```" + `

#### 2. Ventoy USB (Recommended for Hardware)
The most convenient method for physical hardware:
* Install **Ventoy** on your USB flash drive once (https://www.ventoy.net).
* Simply copy and paste the ` + "`.iso`" + ` file onto the Ventoy USB partition.
* No need to reformat when generating new ISOs!

#### 3. Direct USB Flashing (Dedicated Live USB)
Write the ISO directly to a dedicated USB drive:
* **Graphical tools**: Balena Etcher, Popsicle, Raspberry Pi Imager, or your desktop's image writer.
* **Terminal (dd)**:
` + "```bash" + `
sudo dd if=/path/to/egg.iso of=/dev/sdX bs=4M status=progress oflag=sync
` + "```" + `
*(Replace /dev/sdX with your actual USB drive; take extra care not to overwrite system drives!)*
`
	rich := widget.NewRichTextFromMarkdown(markdown)
	rich.Wrapping = fyne.TextWrapWord
	scroll := container.NewScroll(rich)
	scroll.SetMinSize(fyne.NewSize(580, 420))

	d := dialog.NewCustom("Testing & Booting the ISO", "Close", scroll, parent)
	d.SetIcon(theme.HelpIcon())
	d.Resize(fyne.NewSize(620, 480))
	d.Show()
}

func showGRUBISOPickerDialog(parent fyne.Window, nest string, names []string, onSelect func(isoName string)) {
	selection := widget.NewSelect(names, nil)
	selection.SetSelected(names[0])
	picker := dialog.NewForm(fmt.Sprintf("Select ISO from %s", nest), "Continue", "Cancel",
		[]*widget.FormItem{widget.NewFormItem("ISO", selection)}, func(ok bool) {
			if !ok || selection.Selected == "" {
				return
			}
			onSelect(selection.Selected)
		}, parent)
	picker.Resize(fyne.NewSize(720, 180))
	picker.Show()
}
