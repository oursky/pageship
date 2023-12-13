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
	Schedule                     string
	DB                           db.DB
	MaxConsumeActiveDomainCount  uint
	MaxConsumePendingDomainCount uint
	Resolver                     DNSResolver
}

func (v *VerifyDomainOwnership) Name() string { return "verify-domain-ownership" }

func (v *VerifyDomainOwnership) CronSchedule() string { return v.Schedule }

func (v *VerifyDomainOwnership) Run(ctx context.Context, logger *zap.Logger) error {
	now := time.SystemClock.Now().UTC()

	return db.WithTx(ctx, v.DB, func(c db.Tx) error {
		verified := true
		unverified := false
		domainVerifications := []*models.DomainVerification{}
		verifiedRecords, _ := c.ListDomainVerifications(ctx, nil, &v.MaxConsumeActiveDomainCount, &verified)
		unverifedRecords, _ := c.ListDomainVerifications(ctx, nil, &v.MaxConsumePendingDomainCount, &unverified)
		for _, verifiedRecord := range verifiedRecords {
			domainVerifications = append(domainVerifications, verifiedRecord)
		}
		for _, unverifiedRecord := range unverifedRecords {
			domainVerifications = append(domainVerifications, unverifiedRecord)
		}

		var consumedCount = 0
		for _, domainVerification := range domainVerifications {
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
			if matched {
				domainVerification.VerifiedAt = &now
				if errors.Is(err, models.ErrDomainNotFound) {
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
				}
			} else {
				domainVerification.VerifiedAt = nil
				if domain != nil {
					c.DeleteDomain(ctx, domain.ID, now)
				}
			}
			domainVerification.UpdatedAt = now
			err = c.UpdateDomainVerification(ctx, domainVerification)
			if err != nil {
				return err
			}
		}
		logger.Info("verify domain ownership", zap.Int64("n", int64(consumedCount)))
		return nil
	})
}
