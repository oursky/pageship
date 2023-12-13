package cron_test

import (
	"context"
	"testing"
	"time"

	"github.com/foxcpp/go-mockdns"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/cron"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
	"github.com/oursky/pageship/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestVerifyDomainOwnership(t *testing.T) {
	testutil.LoadTestEnvs()

	database, resetDB := testutil.SetupDB()
	defer resetDB()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger, _ := zap.NewDevelopmentConfig().Build()
	now := time.Now()

	var userId = ""
	err := db.WithTx(ctx, database, func(tx db.Tx) error {
		user := models.NewUser(now, "mock_user")
		userId = user.ID
		return tx.CreateUser(ctx, user)
	})
	if err != nil {
		panic(err)
	}
	err = db.WithTx(ctx, database, func(tx db.Tx) error {
		app := models.NewApp(now, "test", userId)
		app.Config.Domains = []config.AppDomainConfig{
			{
				Site:   "main",
				Domain: "test.com",
			},
		}
		err := tx.CreateApp(ctx, app)
		if err != nil {
			return err
		}
		return err
	})
	if err != nil {
		panic(err)
	}
	err = db.WithTx(ctx, database, func(tx db.Tx) error {
		return tx.CreateDomainVerification(ctx, models.NewDomainVerification(now, "test.com", "test"))
	})
	if err != nil {
		panic(err)
	}
	t.Run("Should verify valid domain", func(t *testing.T) {
		domainVerification, err := database.GetDomainVerificationByName(ctx, "test.com")
		if assert.NoError(t, err) {
			assert.Nil(t, domainVerification.VerifiedAt)
		}
		domain, value := domainVerification.GetTxtRecord()
		r := mockdns.Resolver{
			Zones: map[string]mockdns.Zone{
				domain + ".": {
					TXT: []string{value},
				},
			},
		}
		job := cron.VerifyDomainOwnership{
			DB:                           database,
			Resolver:                     &r,
			MaxConsumeActiveDomainCount:  1,
			MaxConsumePendingDomainCount: 1,
		}
		job.Run(ctx, logger)

		domainVerification, err = database.GetDomainVerificationByName(ctx, "test.com")
		if assert.NoError(t, err) {
			assert.NotNil(t, domainVerification.VerifiedAt)
			assert.True(t, domainVerification.VerifiedAt.After(now))
		}
	})

}
