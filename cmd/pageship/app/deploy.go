package app

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/manifoldco/promptui"
	"github.com/oursky/pageship/internal/api"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/deploy"
	"github.com/oursky/pageship/internal/models"
	"github.com/oursky/pageship/internal/time"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.PersistentFlags().String("site", "", "target site")
	deployCmd.PersistentFlags().BoolP("yes", "y", false, "skip confirmation")
}

func packTar(fsys fs.FS, tarfile *os.File) ([]models.FileEntry, int64, error) {
	now := time.SystemClock.Now()
	files, err := deploy.CollectFileList(fsys, now, tarfile)
	if err != nil {
		return nil, 0, err
	}

	_, err = tarfile.Seek(0, io.SeekStart)
	if err != nil {
		return nil, 0, err
	}

	fi, err := tarfile.Stat()
	if err != nil {
		return nil, 0, err
	}

	return files, fi.Size(), nil
}

func doDeploy(ctx context.Context, appID string, site string, conf *config.SiteConfig, fsys fs.FS) {
	tarfile, err := os.CreateTemp("", fmt.Sprintf("pageship-%s-%s-*.tar.zst", appID, site))
	if err != nil {
		Error("Failed to create temp file: %s", err)
		return
	}
	defer os.Remove(tarfile.Name())

	Info("Collecting files...")
	Debug("Tarball: %s", tarfile.Name())
	files, tarSize, err := packTar(fsys, tarfile)
	if err != nil {
		Error("failed to collect files: %s", err)
		return
	}

	Info("%d files found. Tarball size: %s", len(files), humanize.Bytes(uint64(tarSize)))

	Info("Setting up deployment...")
	deployment, err := apiClient.SetupDeployment(ctx, appID, site, files, conf)
	if err != nil {
		Error("Failed to setup deployment: %s", err)
		return
	}

	Debug("Deployment ID: %s", deployment.ID)
	Debug("Site ID: %s", deployment.SiteID)

	bar := progressbar.DefaultBytes(tarSize, "uploading")
	body := io.TeeReader(tarfile, bar)
	deployment, err = apiClient.UploadDeploymentTarball(ctx, appID, site, deployment.ID, body)
	if err != nil {
		Error("Failed to upload tarball: %s", err)
		return
	}

	Info("Activating deployment...")
	status := models.DeploymentStatusActive
	deployment, err = apiClient.PatchDeployment(ctx, appID, site, deployment.ID, &api.DeploymentPatchRequest{
		Status: &status,
	})
	if err != nil {
		Error("Failed to activate deployment: %s", err)
		return
	}

	Debug("Deployment: %+v", deployment)
	Info("Done!")
}

var deployCmd = &cobra.Command{
	Use:   "deploy directory",
	Short: "Deploy site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		site := viper.GetString("site")
		yes := viper.GetBool("yes")
		fsys := os.DirFS(args[0])

		loader := config.NewLoader(config.SiteConfigName)

		conf := config.DefaultConfig()
		if err := loader.Load(fsys, &conf); err != nil {
			Error("Failed to load config: %s", err)
			return
		}
		conf.SetDefaults()

		appID := conf.ID
		if site == "" {
			site = conf.DefaultSite
		}

		if !config.ValidateDNSLabel(site) {
			Error("Invalid site name; site name must be a valid DNS label: %s", site)
			return
		}

		env, ok := conf.ResolveSite(site)
		if !ok {
			Error("Site is not defined by any environment: %s", site)
			return
		}

		if !yes {
			var label string
			if site == env.Name {
				label = fmt.Sprintf("Deploy to site '%s' of app '%s'?", site, appID)
			} else {
				label = fmt.Sprintf("Deploy to site '%s' (%s) of app '%s'?", site, env.Name, appID)
			}

			prompt := promptui.Prompt{Label: label, IsConfirm: true}
			_, err := prompt.Run()
			if err != nil {
				Info("Cancelled.")
				return
			}
		}

		doDeploy(cmd.Context(), appID, site, &conf.Site, fsys)
	},
}
