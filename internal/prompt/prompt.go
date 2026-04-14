package prompt

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"
)

// IsTTY returns true if stdin is a terminal.
func IsTTY() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

// RequireArg checks if a value is provided, prompts for it if TTY, or returns an error.
func RequireArg(name, value string) (string, error) {
	if value != "" {
		return value, nil
	}
	if !IsTTY() {
		return "", fmt.Errorf("required: --%s (or provide interactively in a terminal)", name)
	}
	return AskString(fmt.Sprintf("Enter %s:", name))
}

// RequireArgSensitive is like RequireArg but uses a password input.
func RequireArgSensitive(name, value string) (string, error) {
	if value != "" {
		return value, nil
	}
	if !IsTTY() {
		return "", fmt.Errorf("required: --%s (or provide interactively in a terminal)", name)
	}
	return AskPassword(fmt.Sprintf("Enter %s:", name))
}

// AskString prompts the user for a string input.
func AskString(prompt string) (string, error) {
	var result string
	err := huh.NewInput().
		Title(prompt).
		Value(&result).
		Run()
	return result, err
}

// AskPassword prompts the user for a password (masked input).
func AskPassword(prompt string) (string, error) {
	var result string
	err := huh.NewInput().
		Title(prompt).
		EchoMode(huh.EchoModePassword).
		Value(&result).
		Run()
	return result, err
}

// AskConfirm prompts the user for a yes/no confirmation.
func AskConfirm(prompt string) (bool, error) {
	var result bool
	err := huh.NewConfirm().
		Title(prompt).
		Value(&result).
		Run()
	return result, err
}

// AskSelect prompts the user to select one option from a list.
func AskSelect(title string, options []huh.Option[string]) (string, error) {
	var result string
	err := huh.NewSelect[string]().
		Title(title).
		Options(options...).
		Value(&result).
		Run()
	return result, err
}
