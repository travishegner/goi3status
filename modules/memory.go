package modules

import (
	"time"

	"github.com/travishegner/goi3status/types"
)

func init() {
	addModMap("Memory", NewMemory)
}

type memoryConfig struct {
	*types.BaseModuleConfig
}

// Memory is a module representing the machines load average
type Memory struct {
	*types.BaseModule
	config *memoryConfig
}

func newMemoryConfig(mc types.ModuleConfig) *memoryConfig {
	bmc := types.NewBaseModuleConfig(mc)

	return &memoryConfig{
		BaseModuleConfig: bmc,
	}
}

// NewMemory returns the LoadAverage module
func NewMemory(mc types.ModuleConfig) types.Module {
	config := newMemoryConfig(mc)
	bm := types.NewBaseModule()
	m := &Memory{
		BaseModule: bm,
		config:     config,
	}

	ticker := time.NewTicker(m.config.Refresh)

	go func() {
		for {
			select {
			case <-bm.Done:
				return
			case <-ticker.C:
				bm.Update <- m.MakeBlocks()
			}
		}
	}()

	return m
}

// MakeBlocks returns the Block array for this module
func (m *Memory) MakeBlocks() []*types.Block {
	b := make([]*types.Block, 0)
	if m.config.Label != "" {
		block := types.NewBlock()
		block.FullText = m.config.Label
		block.RemoveSeparator()
		b = append(b, block)
	}

	return b
}

// GetUpdateChan returns the channel down which new block arrays are sent
func (m *Memory) GetUpdateChan() chan []*types.Block {
	return m.Update
}

// Stop stops this module from polling and sending updated Block arrays
func (m *Memory) Stop() {
	close(m.Done)
}
