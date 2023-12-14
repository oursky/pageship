package cron

import (
	"context"
	"errors"
	"net"

	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
	"github.com/oursky/pageship/internal/time"
	"go.uber.org/zap"
)

type DNSResolver interface {
	LookupTXT(context context.Context, key string) ([]string, error)
}

type VerifyDomainOwnership struct {
	Clock                        time.Clock
	Schedule                     string
	DB                           db.DB
	MaxConsumeActiveDomainCount  uint
	MaxConsumePendingDomainCount uint
	RevalidatePeriod             time.Duration
	Resolver                     DNSResolver
}

func (v *VerifyDomainOwnership) Name() string { return "verify-domain-ownership" }

func (v *VerifyDomainOwnership) CronSchedule() string { return v.Schedule }

func (v *VerifyDomainOwnership) Run(ctx context.Context, logger *zap.Logger) error {
	clock := v.Clock
	if clock == nil {
		clock = time.SystemClock
	}
	now := clock.Now().UTC()

	return db.WithTx(ctx, v.DB, func(c db.Tx) error {
		domainVerifications := []*models.DomainVerification{}
		verifiedRecords, err := c.ListLeastRecentlyCheckedDomain(ctx, now, true, v.MaxConsumeActiveDomainCount)
		if err != nil {
			return err
		}
		unverifedRecords, err := c.ListLeastRecentlyCheckedDomain(ctx, now, false, v.MaxConsumePendingDomainCount)
		if err != nil {
			return err
		}
		for _, verifiedRecord := range verifiedRecords {
			domainVerifications = append(domainVerifications, verifiedRecord)
		}
		for _, unverifiedRecord := range unverifedRecords {
			domainVerifications = append(domainVerifications, unverifiedRecord)
		}

		var consumedCount = 0
		for _, domainVerification := range domainVerifications {
			if domainVerification.WillCheckAt.After(now) {
				continue
			}
			domainVerification.LastCheckedAt = &now
			key, value := domainVerification.GetTxtRecord()
			values, err := v.Resolver.LookupTXT(ctx, key)
			var dnsErr *net.DNSError
			if err != nil && errors.As(err, &dnsErr) && !dnsErr.IsNotFound {
				continue
			}
			consumedCount += 1
			matched := false
			for _, v := range values {
				if v == value {
					matched = true
					break
				}
			}
			domain, err := c.GetDomainByName(ctx, domainVerification.Domain)
			if err != nil && !errors.Is(err, models.ErrDomainNotFound) {
				continue
			}
			if matched {
				if domain == nil {
					domainName := domainVerification.Domain
					app, err := c.GetApp(ctx, domainVerification.AppID)
					if err != nil {
						continue
					}
					config, ok := app.Config.ResolveDomain(domainName)
					if ok {
						c.CreateDomain(ctx, models.NewDomain(
							now, domainName, domainVerification.AppID, config.Site,
						))
					}
				} else if domain.AppID != domainVerification.AppID {
					domainName := domainVerification.Domain
					app, err := c.GetApp(ctx, domainVerification.AppID)
					if err != nil {
						continue
					}
					config, ok := app.Config.ResolveDomain(domainName)
					if ok {
						c.DeleteDomain(ctx, domain.ID, now)
						c.CreateDomain(ctx, models.NewDomain(
							now, domainName, domainVerification.AppID, config.Site,
						))
					}
				}
				err = c.LabelDomainVerificationAsVerified(ctx, domainVerification.ID, now, now.Add(v.RevalidatePeriod))
				if err != nil {
					return err
				}
			} else {
				if domain != nil {
					c.DeleteDomain(ctx, domain.ID, now)
				}
				err = c.LabelDomainVerificationAsInvalid(ctx, domainVerification.ID, now)
				if err != nil {
					return err
				}
			}
		}
		logger.Info("verify domain ownership", zap.Int64("n", int64(consumedCount)))
		return nil
	})
}
