package console

import (
	"fmt"
	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/irvankadhafi/go-point-of-sales/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "go-point-of-sales-service",
	Short: "go point of sales service console",
	Long:  `This is go point of sales service console`,
}

// Execute :nodoc:
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	config.GetConf()
	setupLogger()
}

func setupLogger() {
	formatter := runtime.Formatter{
		ChildFormatter: &log.JSONFormatter{},
		Line:           true,
		File:           true,
	}

	if config.Env() == "development" {
		formatter = runtime.Formatter{
			ChildFormatter: &log.TextFormatter{
				ForceColors:   true,
				FullTimestamp: true,
			},
			Line: true,
			File: true,
		}
	}

	log.SetFormatter(&formatter)
	log.SetOutput(os.Stdout)

	logLevel, err := log.ParseLevel(config.LogLevel())
	if err != nil {
		logLevel = log.DebugLevel
	}
	log.SetLevel(logLevel)
}

func continueOrFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func continueOrError(err error) {
	if err != nil {
		log.Error(err)
	}
}
