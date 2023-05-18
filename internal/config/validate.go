package config

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	validate.RegisterValidation("regexp", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		_, err := regexp.Compile(value)
		return err == nil
	})

	validate.RegisterValidation("dnsLabel", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return ValidateDNSLabel(value)
	})
}

var dnsLabel = regexp.MustCompile("^[a-z]([-a-z0-9]*[a-z0-9])?$")

const dnsLabelMaxLength = 63

func ValidateDNSLabel(value string) bool {
	if len(value) > dnsLabelMaxLength {
		return false
	}
	if !dnsLabel.MatchString(value) {
		return false
	}
	return true
}
