package main

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/travishegner/goi3status/modules"
	"github.com/travishegner/goi3status/types"
)

// Status represents the overall status bar
type Status struct {
	modules []types.Module
	cache   [][]*types.Block
	config  *types.Config
	update  chan struct{}
	done    chan struct{}
}

// NewStatus returns an instance of Status
func NewStatus(c *types.Config) *Status {
	mods := make([]types.Module, 0)
	for _, m := range c.Modules {
		name, ok := m["name"].(string)
		if !ok {
			log.Fatalf("module name not defined")
		}
		mc, _ := m["config"].(map[interface{}]interface{})
		mod, err := modules.GetModule(name, mc)
		if err != nil {
			log.Errorf("failed to load module: %v, %v", name, err)
			continue
		}
		mods = append(mods, mod)
	}
	cache := make([][]*types.Block, len(mods))
	update := make(chan struct{})
	done := make(chan struct{})
	s := &Status{modules: mods, cache: cache, config: c, update: update, done: done}

	s.updateCache()

	go s.render(done)
	go s.watchModules(done)

	return s
}

func (s *Status) render(done chan struct{}) {
	j, err := json.Marshal(s.config)
	if err != nil {
		log.Fatalf("error marshalling json: %v", err)
	}

	s.write(string(j))
	s.write("[")

	for {
		select {
		case <-done:
			return
		case <-s.update:
			j, err = json.Marshal(s.flattenCache())
			if err != nil {
				log.Errorf("failed to render status: %v", err)
			}

			s.write(string(j) + ",")
		}
	}

}

func (s *Status) write(line string) {
	fmt.Printf("%v\n", line)
}

func (s *Status) flattenCache() []*types.Block {
	f := make([]*types.Block, 0)
	for _, ab := range s.cache {
		for _, b := range ab {
			f = append(f, b)
		}
	}
	return f

}

func (s *Status) watchModules(done chan struct{}) {
	for {
		start := time.Now()
		select {
		case <-done:
			return
		default:
			if s.updateCache() {
				s.update <- struct{}{}
			}
		}
		stop := time.Now()
		elapsed := stop.Sub(start)
		// this is effectively the maximum refresh rate of the status bar
		// should this be configurable?
		time.Sleep((100 * time.Millisecond) - elapsed)
	}
}

func (s *Status) updateCache() bool {
	update := false
	for i, m := range s.modules {
		select {
		case blocks := <-m.GetUpdateChan():
			s.cache[i] = blocks
			update = true
		default:
		}
	}
	return update
}

// Stop closes the done channel which signals all modules to stop
func (s *Status) Stop() {
	close(s.done)
}
