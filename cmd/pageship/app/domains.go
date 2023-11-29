package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/oursky/pageship/internal/api"
	"github.com/oursky/pageship/internal/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(domainsCmd)
	domainsCmd.PersistentFlags().String("app", "", "app ID")

	domainsCmd.AddCommand(domainsActivateCmd)
	domainsCmd.AddCommand(domainsDeactivateCmd)
}

var domainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "Manage custom domains",
	RunE: func(cmd *cobra.Command, args []string) error {
		appID := viper.GetString("app")
		if appID == "" {
			appID = tryLoadAppID()
		}
		if appID == "" {
			return fmt.Errorf("app ID is not set")
		}

		app, err := API().GetApp(cmd.Context(), appID)
		if err != nil {
			return fmt.Errorf("failed to get app: %w", err)
		}

		type domainEntry struct {
			name  string
			site  string
			model *api.APIDomain
		}
		domains := map[string]domainEntry{}
		for _, dconf := range app.Config.Domains {
			domains[dconf.Domain] = domainEntry{
				name:  dconf.Domain,
				site:  dconf.Site,
				model: nil,
			}
		}

		apiDomains, err := API().ListDomains(cmd.Context(), appID)
		if err != nil {
			return fmt.Errorf("failed to list domains: %w", err)
		}

		for _, d := range apiDomains {
			dd := d
			domains[d.Domain.Domain] = domainEntry{
				name:  d.Domain.Domain,
				site:  d.Domain.SiteName,
				model: &dd,
			}
		}

		w := tabwriter.NewWriter(os.Stdout, 1, 4, 4, ' ', 0)
		fmt.Fprintln(w, "NAME\tSITE\tCREATED AT\tSTATUS")
		for _, domain := range domains {
			createdAt := "-"
			site := "-"
			if domain.model != nil {
				createdAt = domain.model.CreatedAt.Local().Format(time.DateTime)
				site = fmt.Sprintf("%s/%s", domain.model.AppID, domain.model.SiteName)
			} else {
				site = fmt.Sprintf("%s/%s", app.ID, domain.site)
			}

			var status string
			switch {
			case domain.model != nil && domain.model.AppID != app.ID:
				status = "IN_USE"
			case domain.model != nil && domain.model.AppID == app.ID:
				status = "ACTIVE"
			default:
				status = "INACTIVE"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", domain.name, site, createdAt, status)
		}
		w.Flush()
		return nil
	},
}

func promptDomainReplaceApp(ctx context.Context, appID string, domainName string) (replaceApp string, err error) {
	domains, err := API().ListDomains(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("failed list domain: %w", err)
	}

	appID = ""
	for _, d := range domains {
		if d.Domain.Domain == domainName {
			appID = d.AppID
		}
	}

	if appID == "" {
		return "", models.ErrDomainUsedName
	}

	label := fmt.Sprintf("Domain %q is in use by app %q; activates the domain anyways", domainName, appID)

	prompt := promptui.Prompt{Label: label, IsConfirm: true}
	_, err = prompt.Run()
	if err != nil {
		Info("Cancelled.")
		return "", ErrCancelled
	}

	return appID, nil
}

var domainsActivateCmd = &cobra.Command{
	Use:   "activate",
	Short: "Activate domain for the app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domainName := args[0]

		appID := viper.GetString("app")
		if appID == "" {
			appID = tryLoadAppID()
		}
		if appID == "" {
			return fmt.Errorf("app ID is not set")
		}

		app, err := API().GetApp(cmd.Context(), appID)
		if err != nil {
			return fmt.Errorf("failed to get app: %w", err)
		}
		if _, ok := app.Config.ResolveDomain(domainName); !ok {
			return fmt.Errorf("undefined domain")
		}

		_, err = API().CreateDomain(cmd.Context(), appID, domainName, "")
		if code, ok := api.ErrorStatusCode(err); ok && code == http.StatusConflict {
			var replaceApp string
			replaceApp, err = promptDomainReplaceApp(cmd.Context(), appID, domainName)
			if err != nil {
				return err
			}
			_, err = API().CreateDomain(cmd.Context(), appID, domainName, replaceApp)
		}

		if err != nil {
			return fmt.Errorf("failed to create domain: %w", err)
		}

		Info("Domain %q activated.", domainName)
		return nil
	},
}

var domainsDeactivateCmd = &cobra.Command{
	Use:   "deactivate",
	Short: "Deactivate domain for the app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domainName := args[0]

		appID := viper.GetString("app")
		if appID == "" {
			appID = tryLoadAppID()
		}
		if appID == "" {
			return fmt.Errorf("app ID is not set")
		}

		_, err := API().DeleteDomain(cmd.Context(), appID, domainName)
		if err != nil {
			return fmt.Errorf("failed to delete domain: %w", err)
		}

		Info("Domain %q deactivated.", domainName)
		return nil
	},
}
