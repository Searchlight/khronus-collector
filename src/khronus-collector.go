package main

import (
	"collector"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/VividCortex/godaemon"
	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "khronus-collector"
	app.Version = "0.0.1"
	app.Usage = "collect system data and send it to khronus"
	app.Action = func(c *cli.Context) {
		config := make(map[string]interface{})
		config = map[string]interface{}{
			"collectors": map[string]interface{}{
				"CpuCollector": map[string]interface{}{
					"Interval": 1,
				},
				"MemCollector": map[string]interface{}{
					"Interval": 1,
				},
				"DiskCollector": map[string]interface{}{
					"Interval": 1,
				},
				"NetCollector": map[string]interface{}{
					"Interval": 1,
				},
				"StatsdCollector": map[string]interface{}{
					"Interval": 1,
				},
			},
			"outputs": map[string]interface{}{
				"KhronusOutput": map[string]interface{}{
					"Prefix": "/khronus/metrics",
					"Urls": []string{
						"http://10.2.7.11",
					},
					"Interval": 30,
				},
			},
		}

		if c.Bool("show-config") {
			show, _ := json.Marshal(&config)
			fmt.Printf("%s\n", show)
			os.Exit(0)
		}

		if c.Bool("daemon") {
			godaemon.MakeDaemon(&godaemon.DaemonAttr{})
		}

		if c.Bool("pprof") {
			go func() {
				log.Println(http.ListenAndServe("localhost:8888", nil))
			}()
		}

		if c.String("configfile") != "" {
			content, err := ioutil.ReadFile(c.String("configfile"))
			if err != nil {
				log.Println("Error loading configfile, continue with default")
			}

			err = json.Unmarshal([]byte(strings.Join(strings.Split(string(content), "\n"), " ")), &config)

			if err != nil {
				fmt.Printf("Error loading configuration : %s\n", err)
			} else {
				fmt.Println("Configuration loaded")
			}
		} else if c.String("config") != "" {
			err := json.Unmarshal([]byte(c.String("config")), &config)

			if err != nil {
				fmt.Printf("Error loading configuration : %s\n", err)
			} else {
				fmt.Println("Configuration loaded")
			}
		}

		m := collector.Manager{}

		jsonconfig, _ := json.Marshal(&config)

		fmt.Printf("Starting khronus configuration\n")
		fmt.Printf("Config options: %s\n", jsonconfig)
		m.Config(config)
		fmt.Printf("Starting khronus collector\n")
		m.Run()

	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: "",
			Usage: "Array of json formated config",
		},
		cli.BoolFlag{
			Name:  "pprof",
			Usage: "Enable pprof on http port 8888",
		},
		cli.BoolFlag{
			Name:  "daemon",
			Usage: "Daemon mode",
		},

		cli.BoolFlag{
			Name:  "show-config",
			Usage: "Show config json available",
		},

		cli.StringFlag{
			Name:  "configfile",
			Value: "",
			Usage: "Path to json formated configfile",
		},
	}

	app.Run(os.Args)

}
