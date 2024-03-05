package app

import (
	"os"

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
var dockerfileFrom = "FROM ghcr.io/oursky/pageship:v"
var dockerfileExpose = "EXPOSE 8000"
var dockerfileCopy = "COPY . /var/pageship"
var dockerfileInstructions = `
# INSTRUCTIONS:
# 1. install docker (if it is not installed yet)
# 2. open a terminal and navigate to your static page folder
# 3. run in terminal:
#      pageship generate dockerfile
# 4. build the image:
#      docker build -t IMAGETAG .
# 5. run the container:
#      docker run -d --name CONTAINERNAME -p PORT:8000 IMAGETAG
# 6. visit in browser:
#      localhost:PORT`

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
		_, err = f.Write([]byte(dockerfileFrom + Version + "\n"))
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(dockerfileExpose + "\n"))
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(dockerfileCopy + "\n"))
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(dockerfileInstructions + "\n"))
		if err != nil {
			return err
		}
		return nil
	},
}
