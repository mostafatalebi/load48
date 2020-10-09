package main

import (
	"fmt"
	"github.com/mostafatalebi/dynamic-params"
	"github.com/mostafatalebi/loadtest/pkg/config"
	"github.com/mostafatalebi/loadtest/pkg/loadtest"
	"log"
	"os"
	"os/signal"
)

var Version = ""

func main() {
	CheckCommandEntry()
	cp := dyanmic_params.NewDynamicParams(dyanmic_params.SrcNameArgs, os.Args)
	fileName, _ := cp.GetAsString("file")
	var cnf []*config.Config
	var err error
	if fileName == ""  {
		var configLoader = config.NewConfig("cli")
		cnf, err = configLoader.LoadConfigs(os.Args)
		if err != nil {
			log.Print("incorrect config", err.Error())
			return
		}
	} else {
		var configLoader = config.NewConfig("yaml")
		cnf, err = configLoader.LoadConfigs(fileName)
		if err != nil {
			log.Print("incorrect config", err.Error())
			return
		}
	}

	if cnf == nil {
		log.Panic("cannot understand config type, 'cli' and 'yml' are supported")
	}
	lt := loadtest.NewLoadTest(cnf...)
	initCancelInterrupt()
	lt.StartWorkers()
	lt.PrintWorkersStats()
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

func initCancelInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func(){
		for sig := range c {
			if sig == os.Interrupt {
				fmt.Printf("Shutting down the test...\n")
				os.Exit(1)
			}
		}
	}()
}