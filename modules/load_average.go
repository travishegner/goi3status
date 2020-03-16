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
	*types.BaseModuleConfig
}

// LoadAverage is a module representing the machines load average
type LoadAverage struct {
	*types.BaseModule
	config *loadAverageConfig
}

func newLoadAverageConfig(mc types.ModuleConfig) *loadAverageConfig {
	bmc := types.NewBaseModuleConfig(mc)

	return &loadAverageConfig{
		BaseModuleConfig: bmc,
	}
}

// NewLoadAverage returns the LoadAverage module
func NewLoadAverage(mc types.ModuleConfig) types.Module {
	config := newLoadAverageConfig(mc)
	bm := types.NewBaseModule()
	la := &LoadAverage{
		BaseModule: bm,
		config:     config,
	}

	bm.Update <- la.MakeBlocks()
	ticker := time.NewTicker(la.config.Refresh)

	go func() {
		for {
			select {
			case <-bm.Done:
				return
			case <-ticker.C:
				bm.Update <- la.MakeBlocks()
			}
		}
	}()

	return la
}

// MakeBlocks returns the Block array for this module
func (la *LoadAverage) MakeBlocks() []*types.Block {
	b := make([]*types.Block, 0)
	if la.config.Label != "" {
		block := types.NewBlock(la.config.BlockSeparatorWidth)
		block.FullText = la.config.Label
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

	block := types.NewBlock(la.config.BlockSeparatorWidth)
	block.FullText = fmt.Sprintf("%01.02v", avg.Load1)
	block.Color = GetColor(avg.Load1 / cores)
	b = append(b, block)

	block = types.NewBlock(la.config.BlockSeparatorWidth)
	block.FullText = fmt.Sprintf("%01.02v", avg.Load5)
	block.Color = GetColor(avg.Load5 / cores)
	b = append(b, block)

	block = types.NewBlock(la.config.FinalSeparatorWidth)
	if la.config.FinalSeparator {
		block.AddSeparator()
	}
	block.FullText = fmt.Sprintf("%01.02v", avg.Load15)
	block.Color = GetColor(avg.Load15 / cores)
	b = append(b, block)

	return b
}

// GetUpdateChan returns the channel down which new block arrays are sent
func (la *LoadAverage) GetUpdateChan() chan []*types.Block {
	return la.Update
}

// Stop stops this module from polling and sending updated Block arrays
func (la *LoadAverage) Stop() {
	close(la.Done)
}
