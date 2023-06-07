package app

import (
	"errors"

	"github.com/manifoldco/promptui"
)

func Prompt(label string, validate func(value string) error) (string, error) {
	prompt := promptui.Prompt{
		Label: label,
		Validate: func(s string) error {
			if s == "" {
				return errors.New("must not be empty")

			}
			if validate != nil {
				return validate(s)
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		Info("Cancelled.")
		return "", ErrCancelled
	}
	return result, nil
}

func Confirm(label string) error {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
	}
	_, err := prompt.Run()
	if err != nil {
		Info("Cancelled.")
		return ErrCancelled
	}
	return nil
}
