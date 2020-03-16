package modules

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"
	"github.com/travishegner/goi3status/types"
)

func init() {
	addModMap("Memory", NewMemory)
}

type memoryConfig struct {
	*types.BaseModuleConfig
	Attribute string
}

// Memory is a module representing the machines memory
type Memory struct {
	*types.BaseModule
	config *memoryConfig
}

func newMemoryConfig(mc types.ModuleConfig) *memoryConfig {
	bmc := types.NewBaseModuleConfig(mc)

	attr, ok := mc["attribute"].(string)
	if !ok {
		attr = "ram_used_percent"
	}

	return &memoryConfig{
		BaseModuleConfig: bmc,
		Attribute:        attr,
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

	bm.Update <- m.MakeBlocks()
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
		block := types.NewBlock(m.config.BlockSeparatorWidth)
		block.FullText = m.config.Label
		b = append(b, block)
	}

	var err error
	block := types.NewBlock(m.config.FinalSeparatorWidth)
	if m.config.FinalSeparator {
		block.AddSeparator()
	}

	var swp *mem.SwapMemoryStat
	if strings.Split(m.config.Attribute, "_")[0] == "swap" {
		swp, err = mem.SwapMemory()
		if err != nil {
			log.Warningf("failed to get swap information: %v", err.Error())
			return b
		}
		block.Color = GetColor(swp.UsedPercent / 100)
	}

	var ram *mem.VirtualMemoryStat
	if strings.Split(m.config.Attribute, "_")[0] == "ram" {
		ram, err = mem.VirtualMemory()
		if err != nil {
			log.Warningf("failed to get ram information: %v", err.Error())
			return b
		}
		block.Color = GetColor(ram.UsedPercent / 100)
	}

	switch m.config.Attribute {
	case "swap_used":
		block.FullText = humanize.IBytes(swp.Used)
	case "swap_free":
		block.FullText = humanize.IBytes(swp.Free)
	case "swap_used_percent":
		block.FullText = fmt.Sprintf("%v%%", int(swp.UsedPercent))
	case "swap_string":
		block.FullText = swp.String()
	case "ram_total":
		block.FullText = humanize.IBytes(ram.Total)
	case "ram_available":
		block.FullText = humanize.IBytes(ram.Available)
	case "ram_used":
		block.FullText = humanize.IBytes(ram.Used)
	case "ram_used_percent":
		block.FullText = fmt.Sprintf("%v%%", int(ram.UsedPercent))
	case "ram_free":
		block.FullText = humanize.IBytes(ram.Free)
	case "ram_string":
		block.FullText = ram.String()
	}

	b = append(b, block)

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
