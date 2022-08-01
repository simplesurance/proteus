package specialflags

var (
	// Help is the special flag, provided using "--help", used to ask
	// the application to show usage details.
	Help = SpecialParam{
		Name:        "help",
		Description: "Display usage instructions",
	}

	// DryMode is the special flag, provided using "--dry-mode" used to
	// ask the application to validate all other parameters. If they are
	// valid, the application will immediately terminates with exit code 0.
	// If the flags are invalid the application will exit with exist code
	// 1 and output details about the cause for the failed validation.
	DryMode = SpecialParam{
		Name:        "dry-mode",
		Description: "Validates parameters and return either status code 0 or status code 1 and details about validation failure",
	}
)

// SpecialParam holds information about a special parameter.
type SpecialParam struct {
	Name        string
	Description string
}
