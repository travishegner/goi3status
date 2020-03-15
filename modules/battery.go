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
	types.BaseModuleConfig
	Attribute string
}

// Battery is a module representing the any machine batteries
type Battery struct {
	types.BaseModule
	Config *batteryConfig
}

func newBatteryConfig(mc types.ModuleConfig) *batteryConfig {
	label, ok := mc["label"].(string)
	if !ok {
		label = ""
	}

	refresh, ok := mc["refresh"].(int)
	if !ok {
		refresh = 60000
	}

	attribute, ok := mc["attribute"].(string)
	if !ok {
		attribute = "percent"
	}

	return &batteryConfig{
		Label:     label,
		Refresh:   time.Duration(refresh) * time.Millisecond,
		Attribute: attribute,
	}

}

// NewBattery returns the Battery module
func NewBattery(mc types.ModuleConfig) types.Module {
	config := newBatteryConfig(mc)
	done := make(chan struct{})
	update := make(chan []*types.Block)
	bat := &Battery{Update: update, Done: done, Config: config}

	ticker := time.NewTicker(bat.Config.refresh)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				update <- bat.MakeBlocks()
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
