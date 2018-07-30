package main

import (
	"fmt"
	"log"

	"github.com/falzm/chaos"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	chaosAddr = kingpin.Flag("controller-addr", "Chaos controller address").Default(chaos.DefaultBindAddr).String()

	addCmd                  = kingpin.Command("add", "Add route chaos")
	addCmdFlagDuring        = addCmd.Flag("duration", "Chaos specification duration").String()
	addCmdArgMethod         = addCmd.Arg("method", "HTTP route method").Required().String()
	addCmdArgPath           = addCmd.Arg("path", "HTTP route URL path").Required().String()
	addCmdFlagDelayDuration = addCmd.Flag("delay-duration", "Delay injection duration (in milliseconds)").
				Int()
	addCmdFlagDelayProbability = addCmd.Flag("delay-probability", "Delay injection probability (0 < p < 1)").
					Default("1.0").Float64()
	addCmdFlagErrorStatusCode  = addCmd.Flag("error-status-code", "Error injection status code").Int()
	addCmdFlagErrorMessage     = addCmd.Flag("error-message", "Error injection message").String()
	addCmdFlagErrorProbability = addCmd.Flag("error-probability", "Error injection probability (0 < p < 1)").
					Default("1.0").Float64()

	delCmd          = kingpin.Command("delete", "Delete route chaos").Alias("del")
	delCmdArgMethod = delCmd.Arg("method", "HTTP route method").Required().String()
	delCmdArgPath   = delCmd.Arg("path", "HTTP route URL path").Required().String()
)

func main() {
	kingpin.CommandLine.HelpFlag.Short('h')

	switch kingpin.Parse() {
	case "add":
		spec := chaos.NewSpec()

		if *addCmdFlagDelayDuration > 0 {
			spec.Delay(*addCmdFlagDelayDuration, *addCmdFlagDelayProbability)
		}

		if *addCmdFlagErrorStatusCode > 0 {
			spec.Error(*addCmdFlagErrorStatusCode, *addCmdFlagErrorMessage, *addCmdFlagErrorProbability)
		}

		if *addCmdFlagDuring != "" {
			spec.During(*addCmdFlagDuring)
		}

		if err := chaos.NewClient(*chaosAddr).AddRouteChaos(*addCmdArgMethod, *addCmdArgPath, spec); err != nil {
			log.Fatalf("%s", err)
		}

	case "del", "delete":
		if err := chaos.NewClient(*chaosAddr).DeleteRouteChaos(*delCmdArgMethod, *delCmdArgPath); err != nil {
			log.Fatalf("%s", err)
		}
	}

	fmt.Println("OK")
}
