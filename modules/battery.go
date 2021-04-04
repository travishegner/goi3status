package modules

import (
	"fmt"
	"time"

	"github.com/distatus/battery"
	log "github.com/sirupsen/logrus"
	"github.com/travishegner/goi3status/types"
)

func init() {
	addModMap("Battery", NewBattery)
}

type batteryConfig struct {
	*types.BaseModuleConfig
	Attribute string
}

// Battery is a module representing the any machine batteries
type Battery struct {
	*types.BaseModule
	config *batteryConfig
}

func newBatteryConfig(mc types.ModuleConfig) *batteryConfig {
	bmc := types.NewBaseModuleConfig(mc)
	attribute, ok := mc["attribute"].(string)
	if !ok {
		attribute = "percent"
	}

	return &batteryConfig{
		BaseModuleConfig: bmc,
		Attribute:        attribute,
	}

}

// NewBattery returns the Battery module
func NewBattery(mc types.ModuleConfig) types.Module {
	config := newBatteryConfig(mc)
	bm := types.NewBaseModule()
	bat := &Battery{
		BaseModule: bm,
		config:     config,
	}

	bm.Update <- bat.MakeBlocks()
	ticker := time.NewTicker(bat.config.Refresh)

	go func() {
		for {
			select {
			case <-bm.Done:
				return
			case <-ticker.C:
				bm.Update <- bat.MakeBlocks()
			}
		}
	}()

	return bat
}

// MakeBlocks returns the Block array for this module
func (bat *Battery) MakeBlocks() []*types.Block {
	b := make([]*types.Block, 0)
	if bat.config.Label != "" {
		block := types.NewBlock(bat.config.BlockSeparatorWidth)
		block.FullText = bat.config.Label
		b = append(b, block)
	}

	batteries, err := battery.GetAll()
	if err != nil {
		if fe, ok := err.(battery.ErrFatal); ok {
			log.Errorf("fatal error getting batteries: %v", fe.Error())
			return b
		}
	}

	goodBats := make([]*battery.Battery, 0)
	for i, tb := range batteries {
		if es, ok := err.(battery.Errors); ok {
			if es[i] != nil {
				continue
			}
		}
		goodBats = append(goodBats, tb)
	}

	for i, tb := range goodBats {
		text := ""
		color := ""
		switch bat.config.Attribute {
		case "percent":
			text = fmt.Sprintf("%v", int((tb.Current/tb.Full)*100))
			color = GetColor(1.0 - (tb.Current / tb.Full))
		case "state":
			text = tb.State.String()
		}

		block := types.NewBlock(bat.config.BlockSeparatorWidth)
		block.FullText = text
		if color != "" {
			block.Color = color
		}
		block.SeparatorBlockWidth = bat.config.FinalSeparatorWidth
		if i == len(batteries) && bat.config.FinalSeparator {
			block.AddSeparator()
		}

		b = append(b, block)
	}

	return b
}

// GetUpdateChan returns the channel down which new block arrays are sent
func (bat *Battery) GetUpdateChan() chan []*types.Block {
	return bat.Update
}

// Stop stops this module from polling and sending updated Block arrays
func (bat *Battery) Stop() {
	close(bat.Done)
}
