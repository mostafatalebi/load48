package main

import (
	"github.com/mostafatalebi/dynamic-params"
	"github.com/mostafatalebi/loadtest/pkg/config"
	"github.com/mostafatalebi/loadtest/pkg/core"
	"github.com/mostafatalebi/loadtest/pkg/stats"
	"log"
	"os"
)

var Version = ""

func main() {
	CheckCommandEntry()
	cp := dyanmic_params.NewDynamicParams(dyanmic_params.SrcNameArgs, os.Args)
	configType, _ := cp.GetAsString("type")
	var cnf *config.Config
	var err error
	if configType == "cli"  {
		var configLoader = config.NewConfig("cli")
		cnf, err = configLoader.LoadConfig(os.Args)
		if err != nil {
			log.Print("incorrect config", err.Error())
			return
		}
	} else if configType == "config" {

	} else if configType != "" {
		log.Panic("type must be either 'cli' or 'config'")
	}
	lt := core.NewLoadTest(cnf)

	err = lt.Process()
	if err != nil {
		log.Panic(err)
	}
	st := lt.MergeAll()
	st.PrintPretty(stats.DefaultPresetWithAutoFailedCodes)
	lt.PrintGeneralInfo()
}

func CheckCommandEntry() {
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		PrintHelp()
		os.Exit(0)
		return
	} else if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		PrintVersion()
		os.Exit(0)
		return
	}
}