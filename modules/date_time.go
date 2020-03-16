package modules

import (
	"time"

	"github.com/travishegner/goi3status/types"
)

func init() {
	addModMap("DateTime", NewDateTime)
}

// DateTime is a module for displaying date and/or time in an arbitrary format
type DateTime struct {
	*types.BaseModule
	config *dateTimeConfig
}

type dateTimeConfig struct {
	*types.BaseModuleConfig
	format   string
	location *time.Location
}

func newDateTimeConfig(mc types.ModuleConfig) *dateTimeConfig {
	bmc := types.NewBaseModuleConfig(mc)
	format, ok := mc["format"].(string)
	if !ok {
		format = ""
	}

	timezone, ok := mc["timezone"].(string)
	if !ok {
		timezone = "Local"
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc, _ = time.LoadLocation("Local")
	}

	return &dateTimeConfig{
		BaseModuleConfig: bmc,
		format:           format,
		location:         loc,
	}
}

// NewDateTime creates a new DateTime, starts it's ticker, then returns it
func NewDateTime(mc types.ModuleConfig) types.Module {
	bm := types.NewBaseModule()
	config := newDateTimeConfig(mc)
	dt := &DateTime{
		BaseModule: bm,
		config:     config,
	}

	ticker := time.NewTicker(dt.config.Refresh)

	go func() {
		for {
			select {
			case <-bm.Done:
				return
			case <-ticker.C:
				bm.Update <- dt.MakeBlocks()
			}
		}
	}()

	return dt
}

// MakeBlocks returns the Block array for this module
func (dt *DateTime) MakeBlocks() []*types.Block {
	b := make([]*types.Block, 0)
	if dt.config.Label != "" {
		block := types.NewBlock()
		block.FullText = dt.config.Label
		block.RemoveSeparator()
		b = append(b, block)
	}

	t := time.Now().In(dt.config.location)
	block := types.NewBlock()
	block.FullText = t.Format(dt.config.format)
	b = append(b, block)

	return b
}

// GetUpdateChan returns the channel down which new Block arrays are sent
func (dt *DateTime) GetUpdateChan() chan []*types.Block {
	return dt.Update
}

// Stop stops this module and prevents new Block arrays from being sent
func (dt *DateTime) Stop() {
	close(dt.Done)
}
