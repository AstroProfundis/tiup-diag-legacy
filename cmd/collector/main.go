package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/pingcap/tidb-foresight/collector"
	log "github.com/sirupsen/logrus"
)

func main() {
	opts := Options{}
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(1)
	}

	collector, err := collector.New(&opts)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	if err := collector.Collect(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
