package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/travishegner/goi3status/types"
	"gopkg.in/yaml.v2"
)

func main() {
	conf := `
version: 1
modules:
  - name: DateTime
    config:
      format: "15:04:05 MST"
`

	c := types.Config{}

	err := yaml.Unmarshal([]byte(conf), &c)
	if err != nil {
		log.Fatalf("error unmarshalling config: %v", err)
		os.Exit(1)
	}

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
