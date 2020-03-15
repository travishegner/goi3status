package modules

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/cpu"
	log "github.com/sirupsen/logrus"
	"github.com/travishegner/goi3status/types"
)

func init() {
	addModMap("CPU", NewCPU)
}

// CPU is a module to collect cpu information
type CPU struct {
	update    chan []*types.Block
	done      chan struct{}
	config    *cpuConfig
	graphChar []string
}

type cpuConfig struct {
	refresh     time.Duration
	monitorType string
	average     bool
	label       string
	tempGreen   int64
	tempRed     int64
}

func newCPUConfig(mc types.ModuleConfig) *cpuConfig {
	refresh, ok := mc["refresh"].(int)
	if !ok {
		refresh = 500
	}

	mon, ok := mc["monitor"].(string)
	if !ok {
		mon = "graph"
	}

	avg, ok := mc["average"].(bool)
	if !ok {
		avg = false
	}

	label, ok := mc["label"].(string)
	if !ok {
		label = ""
	}

	tempGreen, ok := mc["temp_green"].(int)
	if !ok {
		tempGreen = 40
	}

	tempRed, ok := mc["temp_red"].(int)
	if !ok {
		tempRed = 80
	}

	return &cpuConfig{
		refresh:     time.Duration(refresh) * time.Millisecond,
		monitorType: mon,
		average:     avg,
		label:       label,
		tempGreen:   int64(tempGreen),
		tempRed:     int64(tempRed),
	}
}

// NewCPU returns a new CPU module
func NewCPU(mc types.ModuleConfig) types.Module {
	config := newCPUConfig(mc)

	char := make([]string, 8)
	char[0] = "\u2581"
	char[1] = "\u2582"
	char[2] = "\u2583"
	char[3] = "\u2584"
	char[4] = "\u2585"
	char[5] = "\u2586"
	char[6] = "\u2587"
	char[7] = "\u2588"

	done := make(chan struct{})
	update := make(chan []*types.Block)
	cpuMod := &CPU{
		update:    update,
		done:      done,
		config:    config,
		graphChar: char,
	}

	ticker := time.NewTicker(config.refresh)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				update <- cpuMod.MakeBlocks()
			}
		}
	}()

	return cpuMod
}

// MakeBlocks returns the i3 blocks for this module
func (c *CPU) MakeBlocks() []*types.Block {
	b := make([]*types.Block, 0)
	if c.config.label != "" {
		block := types.NewBlock()
		block.FullText = c.config.label
		block.RemoveSeparator()
		block.SeparatorBlockWidth = 3
		b = append(b, block)
	}

	switch c.config.monitorType {
	case "graph":
		fallthrough
	case "percent":
		b = append(b, c.makeUtilBlocks()...)
	case "temp":
		b = append(b, c.makeTempBlocks()...)
	}
	return b
}

func (c *CPU) makeTempBlocks() []*types.Block {
	b := make([]*types.Block, 0)
	zones, err := filepath.Glob("/sys/class/thermal/thermal_zone*")
	if err != nil {
		log.Warnf("error getting thermal zones")
		return b
	}

	for i, z := range zones {
		i32, _ := strconv.Atoi(readLine(z + "/temp"))
		temp := int64(i32) / 1000
		block := types.NewBlock()
		block.FullText = fmt.Sprintf("%v\u2103", temp)
		base := temp - c.config.tempGreen
		if base < 0 {
			base = 0
		}
		block.Color = GetColor(float64(base) / float64(c.config.tempRed-c.config.tempGreen))
		if i != len(zones) {
			block.RemoveSeparator()
		}
		b = append(b, block)
	}

	return b
}

func (c *CPU) makeUtilBlocks() []*types.Block {
	b := make([]*types.Block, 0)

	cpus, err := cpu.Percent(c.config.refresh, !c.config.average)
	if err != nil {
		log.Warnf("err getting cpu percentages: %v", err)
	}

	for i, v := range cpus {
		b = append(b, c.getUtilBlock(v, i == len(cpus)))
	}
	return b
}

func (c *CPU) getUtilBlock(val float64, last bool) *types.Block {
	sepWidth := 3
	block := types.NewBlock()
	switch c.config.monitorType {
	case "graph":
		block.FullText = fmt.Sprintf("%v", c.graphChar[int((val/100)*7)])
	case "percent":
		block.FullText = fmt.Sprintf("%v", int(val))
		block.MinWidth = "99"
		block.Align = "right"
		sepWidth = 6
	}
	block.Color = GetColor(val / 100)
	if !last {
		block.RemoveSeparator()
		block.SeparatorBlockWidth = sepWidth
	}
	return block
}

// GetUpdateChan returns the channel down which Block arrays are sent
func (c *CPU) GetUpdateChan() chan []*types.Block {
	return c.update
}

// Stop stops the module
func (c *CPU) Stop() {
	close(c.done)
}