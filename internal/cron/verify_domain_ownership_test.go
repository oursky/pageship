package cron_test

import (
	"context"
	"net"
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

type DBData struct {
	userId string
}

func setupDB(now time.Time, ctx context.Context, database db.DB) DBData {
	userId := ""
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
	return DBData{
		userId: userId,
	}
}

type RaiseErrorDNSResolver struct {
	err net.DNSError
}

func (r *RaiseErrorDNSResolver) LookupTXT(context context.Context, key string) ([]string, error) {
	return nil, &r.err
}

func TestVerifyDomainOwnership(t *testing.T) {
	testutil.LoadTestEnvs()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger, _ := zap.NewDevelopmentConfig().Build()
	now := time.Now()

	t.Run("Should verify valid domain", func(t *testing.T) {
		testutil.WithTestDB(func(database db.DB) {
			setupDB(now, ctx, database)
			domainVerification, err := database.GetDomainVerificationByName(ctx, "test.com", "test")
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
				RevalidatePeriod:             time.Hour,
			}
			job.Run(ctx, logger)

			domainVerification, err = database.GetDomainVerificationByName(ctx, "test.com", "test")
			if assert.NoError(t, err) {
				assert.NotNil(t, domainVerification.VerifiedAt)
				assert.True(t, domainVerification.VerifiedAt.After(now))
			}
		})
	})

	t.Run("Should drop domain when TXT value is invalid", func(t *testing.T) {
		invalidTXTValues := [][]string{
			{"incorrect value"},
			nil,
		}
		for _, values := range invalidTXTValues {
			testutil.WithTestDB(func(database db.DB) {

				setupDB(now, ctx, database)
				domainVerification, err := database.GetDomainVerificationByName(ctx, "test.com", "test")
				if assert.NoError(t, err) {
					assert.Nil(t, domainVerification.VerifiedAt)
				}
				domain, _ := domainVerification.GetTxtRecord()
				db.WithTx(ctx, database, func(tx db.Tx) error {
					return tx.CreateDomain(ctx, models.NewDomain(
						now,
						"test.com",
						"test",
						"main",
					))
				})
				r := mockdns.Resolver{
					Zones: map[string]mockdns.Zone{
						domain + ".": {
							TXT: values,
						},
					},
				}
				job := cron.VerifyDomainOwnership{
					DB:                           database,
					Resolver:                     &r,
					MaxConsumeActiveDomainCount:  1,
					MaxConsumePendingDomainCount: 1,
					RevalidatePeriod:             time.Hour,
				}
				job.Run(ctx, logger)
				_, err = database.GetDomainByName(ctx, "test.com")
				if assert.Error(t, err) {
					assert.ErrorIs(t, models.ErrDomainNotFound, err)
				}
			})
		}
	})
	t.Run("Should drop domain when TXT value is not reacheable", func(t *testing.T) {
		testutil.WithTestDB(func(database db.DB) {

			setupDB(now, ctx, database)
			domainVerification, err := database.GetDomainVerificationByName(ctx, "test.com", "test")
			if assert.NoError(t, err) {
				assert.Nil(t, domainVerification.VerifiedAt)
			}
			db.WithTx(ctx, database, func(tx db.Tx) error {
				return tx.CreateDomain(ctx, models.NewDomain(
					now,
					"test.com",
					"test",
					"main",
				))
			})
			r := RaiseErrorDNSResolver{
				err: net.DNSError{
					IsNotFound: true,
				},
			}
			job := cron.VerifyDomainOwnership{
				DB:                           database,
				Resolver:                     &r,
				MaxConsumeActiveDomainCount:  1,
				MaxConsumePendingDomainCount: 1,
				RevalidatePeriod:             time.Hour,
			}
			job.Run(ctx, logger)
			_, err = database.GetDomainByName(ctx, "test.com")
			if assert.Error(t, err) {
				assert.ErrorIs(t, models.ErrDomainNotFound, err)
			}
		})
	})
	t.Run("Should replace the conflict domain", func(t *testing.T) {
		testutil.WithTestDB(func(database db.DB) {
			data := setupDB(now, ctx, database)
			err := db.WithTx(ctx, database, func(tx db.Tx) error {
				return tx.CreateApp(ctx, models.NewApp(
					now,
					"test2",
					data.userId,
				))
			})
			assert.NoError(t, err)
			err = db.WithTx(ctx, database, func(tx db.Tx) error {
				return tx.CreateDomain(ctx, models.NewDomain(
					now,
					"test.com",
					"test2",
					"main",
				))
			})
			assert.NoError(t, err)
			domainVerification, err := database.GetDomainVerificationByName(ctx, "test.com", "test")
			txtDomain, value := domainVerification.GetTxtRecord()
			r := mockdns.Resolver{
				Zones: map[string]mockdns.Zone{
					txtDomain + ".": {
						TXT: []string{
							value,
						},
					},
				},
			}
			job := cron.VerifyDomainOwnership{
				DB:                           database,
				Resolver:                     &r,
				MaxConsumeActiveDomainCount:  1,
				MaxConsumePendingDomainCount: 1,
				RevalidatePeriod:             time.Hour,
			}
			err = job.Run(ctx, logger)
			assert.NoError(t, err)

			domain, err := database.GetDomainByName(ctx, "test.com")
			if assert.NoError(t, err) {
				assert.Equal(t, "test.com", domain.Domain)
				assert.Equal(t, "test", domain.AppID)
			}
		})
	})
}
