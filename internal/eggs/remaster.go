package eggs

// RemasterMode identifies one of the existing remaster choices.
type RemasterMode string

// Remaster modes retain the labels used by the current GUI.
const (
	ModeStandard  RemasterMode = "Standard live"
	ModeClone     RemasterMode = "System clone"
	ModeEncrypted RemasterMode = "Encrypted clone"
)

// RemasterArgs produces the arguments for a remaster operation.
func (CLIAdapter) RemasterArgs(mode RemasterMode, workingDirectory string) []string {
	args := []string{"remaster"}
	switch mode {
	case ModeClone:
		args = append(args, "--clone")
	case ModeEncrypted:
		args = append(args, "--crypted")
	}
	if workingDirectory != "" {
		args = append(args, "--path", workingDirectory)
	}
	return args
}

// ModeDescription returns a human-readable explanation of the remaster mode.
func ModeDescription(mode RemasterMode) string {
	switch mode {
	case ModeClone:
		return "Include users, configurations and data in an unencrypted clone."
	case ModeEncrypted:
		return "Create an encrypted clone; passphrase and parameters are requested by Eggs."
	default:
		return "Create a clean live system without user accounts or personal data."
	}
}
