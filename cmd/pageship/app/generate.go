package app

import (
	"bytes"
	"os"
	"text/template"

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

var Version = ""

type dockerfileTemplate struct {
	Version    string
	PublicPath string
}

var dockerfileTemplateString = `FROM ghcr.io/oursky/pageship:{{if .Version}}v{{.Version}}{{else}}dev{{end}}
EXPOSE 8000
COPY ./pageship.toml /var/pageship
COPY ./{{.PublicPath}} /var/pageship/{{.PublicPath}}

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

func generateContent() (string, error) {
	cfg, err := loadConfig(".")
	if err != nil {
		return "", err
	}

	df := dockerfileTemplate{Version, cfg.Site.Public}

	tmpl, err := template.New("Dockerfile").Parse(dockerfileTemplateString)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	err = tmpl.Execute(&b, df)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

var generateDockerfileCmd = &cobra.Command{
	Use:   "dockerfile",
	Short: "Generate Dockerfile",
	RunE: func(cmd *cobra.Command, args []string) error {
		if Version == "dev" {
			Version = ""
		}

		f, err := os.Create("Dockerfile")
		if err != nil {
			return err
		}
		defer f.Close()

		Info("generating Dockerfile...")
		s, err := generateContent()
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
