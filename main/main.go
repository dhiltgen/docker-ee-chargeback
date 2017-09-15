package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dhiltgen/docker-ee-chargeback"

	"github.com/codegangsta/cli"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"

	log "github.com/Sirupsen/logrus"
)

func connect(ucp, certsDir string) v1.API {
	tlsCfg, err := tlsconfig.Client(tlsconfig.Options{
		CAFile:   filepath.Join(certsDir, "ca.pem"),
		CertFile: filepath.Join(certsDir, "cert.pem"),
		KeyFile:  filepath.Join(certsDir, "key.pem"),
	})
	if err != nil {
		log.Fatalf("Unable to load certs: %s", err)
	}

	transport := &http.Transport{
		TLSClientConfig: tlsCfg,
	}
	promCfg := api.Config{
		Address:      fmt.Sprintf("https://%s:12387", ucp),
		RoundTripper: transport,
	}
	promClient, err := api.NewClient(promCfg)
	if err != nil {
		log.Fatalf("Unable to connect: %s", err)
	}
	return v1.NewAPI(promClient)
}

func main() {
	app := cli.NewApp()
	app.Name = "report"
	app.Usage = "Generate a CSV report showing usage"
	app.Action = DoGather
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "duration",
			Usage: "Number of minutes to gather data over (you should run this tool at the same interval to not miss data)",
			Value: 60,
		},
		cli.StringFlag{
			Name:   "certs",
			Usage:  "The docker to admin (or cluster) certs to use to connect to the system",
			EnvVar: "CERTS_DIR",
		},
		cli.StringFlag{
			Name:  "ucp",
			Usage: "The IP or hostname of a UCP manager node (without protocol or port - e.g. \"192.168.1.2\")",
		},
		cli.BoolFlag{
			Name:  "system",
			Usage: "include system resources in the results",
		},
		cli.BoolFlag{
			Name:  "omit-header",
			Usage: "omit the header from the CSV output",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "log debug messages to stderr",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func DoGather(c *cli.Context) error {

	ucp := c.String("ucp")
	if ucp == "" {
		return fmt.Errorf("you must specify the UCP IP or hostname with \"--ucp XXX\"")
	}
	certsDir := c.String("certs")
	if certsDir == "" {
		return fmt.Errorf("you must specify the certs dir IP or hostname with \"--certs XXX\"")
	}
	skipSystemResources := !c.Bool("system")
	omitHeader := c.Bool("omit-header")
	if c.Bool("debug") {
		log.SetLevel(log.DebugLevel)
	}
	duration := c.Int("duration")

	log.Debug("Getting started...")
	promAPI := connect(ucp, certsDir)

	log.Debug("Querying...")
	now := time.Now()
	start := now.Add(time.Minute * -time.Duration(duration))
	step := time.Minute
	r := v1.Range{
		Start: start,
		End:   now,
		Step:  step,
	}

	results := []chargeback.Entry{}
	for _, f := range chargeback.Gatherers {
		res, err := f(promAPI, r, skipSystemResources)
		if err != nil {
			log.Fatalf("Failed to query: %s", err)
		}
		results = append(results, res...)
	}

	if !omitHeader {
		fmt.Printf("TYPE,COLLECTION,CONTAINER ID,CONTAINER NAME,SAMPLE DURATION IN SECONDS,CUMULATIVE VALUE,MIN VALUE,MAX VALUE,AVE VALUE\n")
	}
	for _, entry := range results {
		fmt.Printf("%s,%s,%s,%s,%f,%f,%f,%f,%f\n",
			entry.Label,
			entry.Collection,
			entry.ID,
			entry.Name,
			entry.TotalSeconds,
			entry.Cumulative,
			entry.Min,
			entry.Max,
			entry.Ave,
		)

	}
	return nil
}
