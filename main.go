package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/travishegner/goi3status/types"
	"gopkg.in/yaml.v2"
)

func main() {
	cf := flag.String("config", "config.yaml", "config file describing status layout")
	flag.Parse()

	conf, err := ioutil.ReadFile(*cf)
	if err != nil {
		log.Fatalf("error reading config file: %v", err.Error())
	}

	c := types.Config{}
	err = yaml.Unmarshal(conf, &c)
	if err != nil {
		log.Fatalf("error unmarshalling config: %v", err)
		os.Exit(1)
	}
	// This software supports version 1 of the i3bar protocol
	// https://i3wm.org/docs/i3bar-protocol.html
	c.Version = 1

	done := make(chan struct{})
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	status := NewStatus(&c)

	go func() {
		select {
		case <-sig:
			status.Stop()
			close(done)
		}
	}()

	<-done
}
