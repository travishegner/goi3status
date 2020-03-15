package modules

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	log "github.com/sirupsen/logrus"

	"github.com/travishegner/goi3status/types"
)

func init() {
	addModMap("LoadAverage", NewLoadAverage)
}

type loadAverageConfig struct {
	label   string
	refresh time.Duration
}

// LoadAverage is a module representing the machines load average
type LoadAverage struct {
	update chan []*types.Block
	done   chan struct{}
	config *loadAverageConfig
}

func newLoadAverageConfig(mc types.ModuleConfig) *loadAverageConfig {
	label, ok := mc["label"].(string)
	if !ok {
		label = ""
	}

	refresh, ok := mc["refresh"].(int)
	if !ok {
		refresh = 1000
	}

	return &loadAverageConfig{
		label:   label,
		refresh: time.Duration(refresh) * time.Millisecond,
	}

}

// NewLoadAverage returns the LoadAverage module
func NewLoadAverage(mc types.ModuleConfig) types.Module {
	config := newLoadAverageConfig(mc)
	done := make(chan struct{})
	update := make(chan []*types.Block)
	la := &LoadAverage{update: update, done: done, config: config}

	ticker := time.NewTicker(la.config.refresh)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				update <- la.MakeBlocks()
			}
		}
	}()

	return la
}

// MakeBlocks returns the Block array for this module
func (la *LoadAverage) MakeBlocks() []*types.Block {
	b := make([]*types.Block, 0)
	if la.config.label != "" {
		block := types.NewBlock()
		block.FullText = la.config.label
		block.RemoveSeparator()
		b = append(b, block)
	}

	c, err := cpu.Counts(false)
	if err != nil {
		log.Error(err)
		c = 1
	}
	cores := float64(c)

	avg, err := load.Avg()
	if err != nil {
		log.Error(err)
		return b
	}

	block := types.NewBlock()
	block.FullText = fmt.Sprintf("%01.02v", avg.Load1)
	block.Color = GetColor(avg.Load1 / cores)
	block.RemoveSeparator()
	b = append(b, block)

	block = types.NewBlock()
	block.FullText = fmt.Sprintf("%01.02v", avg.Load5)
	block.Color = GetColor(avg.Load5 / cores)
	block.RemoveSeparator()
	b = append(b, block)

	block = types.NewBlock()
	block.FullText = fmt.Sprintf("%01.02v", avg.Load15)
	block.Color = GetColor(avg.Load15 / cores)
	b = append(b, block)

	return b
}

// GetUpdateChan returns the channel down which new block arrays are sent
func (la *LoadAverage) GetUpdateChan() chan []*types.Block {
	return la.update
}

// Stop stops this module from polling and sending updated Block arrays
func (la *LoadAverage) Stop() {
	close(la.done)
}
