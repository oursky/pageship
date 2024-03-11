package app

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/oursky/pageship/internal/config"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.AddCommand(generateDockerfileCmd)
}

var generateCmd = &cobra.Command{
	Use:   "generate [command]",
	Short: "Generate files",
	Args:  cobra.NoArgs, //if unknown command, will return error just like main pageship command
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Help() //show help if no subcommand supplied
		return nil
	},
}

var Version = "dev"
var dockerfileTemplate = `FROM ghcr.io/oursky/pageship:v%s
EXPOSE 8000
COPY %s /var/pageship

# INSTRUCTIONS:
# 1. install docker (if it is not installed yet)
# 2. open a terminal and navigate to folder containing pageship.toml
# 3. run in terminal:
#      pageship generate dockerfile
# 4. build the image:
#      docker build -t IMAGETAG .
# 5. run the container:
#      docker run -d --name CONTAINERNAME -p PORT:8000 IMAGETAG
# 6. visit in browser (URL):
#      localhost:PORT`

func generateContent(myfs fs.FS) (string, error) {
	pageshiptoml, err := fs.ReadFile(myfs, "./pageship.toml")
	if err != nil {
		return "", err
	}

	var cfg config.Config
	err = toml.Unmarshal([]byte(pageshiptoml), &cfg)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(dockerfileTemplate, Version, cfg.Site.Public), nil
}

var generateDockerfileCmd = &cobra.Command{
	Use:   "dockerfile",
	Short: "Generate dockerfile",
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Create("dockerfile")
		if err != nil {
			return err
		}
		defer f.Close()

		Info("generating dockerfile...")
		s, err := generateContent(os.DirFS("."))
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(s))
		if err != nil {
			return err
		}
		return nil
	},
}
