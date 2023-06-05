package config

import (
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	validate.SetTagName("pageship")

	validate.RegisterValidation("regexp", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		_, err := regexp.Compile(value)
		return err == nil
	})

	validate.RegisterValidation("dnsLabel", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return ValidateDNSLabel(value)
	})

	validate.RegisterValidation("duration", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return ValidateDuration(value)
	})

	validate.RegisterValidation("accessLevel", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return AccessLevel(value).IsValid()
	})
}

// ref: RFC1123
var dnsLabel = regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")

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

func ValidateDuration(value string) bool {
	d, err := time.ParseDuration(value)
	if err != nil {
		return false
	}
	if d <= 0 {
		return false
	}
	return true
}

func ValidateAppConfig(conf *AppConfig) error {
	return validate.Struct(conf)
}

func ValidateSiteConfig(conf *SiteConfig) error {
	return validate.Struct(conf)
}
