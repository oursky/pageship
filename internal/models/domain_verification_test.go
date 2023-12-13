package models_test

import (
	"testing"
	"time"

	"github.com/oursky/pageship/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestDomainVerification(t *testing.T) {
	v := models.NewDomainVerification(time.Now(), "example.com", "appId")
	domain, value := v.GetTxtRecord()
	assert.Regexp(t, "\\d{6}\\._pageship\\.example\\.com", domain)
	assert.Regexp(t, "\\d{6}", value)
}
