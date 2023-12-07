package models_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/oursky/pageship/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestDomainVerification(t *testing.T) {
	v := models.NewDomainVerification(time.Now(), "example.com", "appId")
	expectedDommain := fmt.Sprintf("%s._pageship.example.com", v.DomainPrefix)
	domain, _ := v.GetTxtRecord()
	assert.Equal(t, expectedDommain, domain)
}
