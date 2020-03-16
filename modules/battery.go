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
	Config *batteryConfig
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
		Config:     config,
	}

	ticker := time.NewTicker(bat.Config.Refresh)

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
	if bat.Config.Label != "" {
		block := types.NewBlock()
		block.FullText = bat.Config.Label
		block.RemoveSeparator()
		b = append(b, block)
	}

	batteries, err := battery.GetAll()
	if err != nil {
		log.Errorf("failed to get batteries: %v", err.Error())
		return b
	}

	for i, tb := range batteries {
		text := ""
		color := ""
		switch bat.Config.Attribute {
		case "percent":
			text = fmt.Sprintf("%v", int((tb.Current/tb.Full)*100))
			color = GetColor(tb.Current / tb.Full)
		case "state":
			text = tb.State.String()
		}

		block := types.NewBlock()
		block.FullText = text
		if color != "" {
			block.Color = color
		}
		if i != len(batteries) {
			block.RemoveSeparator()
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