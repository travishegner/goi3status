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
	update chan []*types.Block
	done   chan struct{}
	config *dateTimeConfig
}

type dateTimeConfig struct {
	label    string
	refresh  time.Duration
	format   string
	location *time.Location
}

func newDateTimeConfig(mc types.ModuleConfig) *dateTimeConfig {
	label, ok := mc["label"].(string)
	if !ok {
		label = ""
	}

	refresh, ok := mc["refresh"].(int)
	if !ok {
		refresh = 1000
	}

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
		label:    label,
		refresh:  time.Duration(refresh) * time.Millisecond,
		format:   format,
		location: loc,
	}
}

// NewDateTime creates a new DateTime, starts it's ticker, then returns it
func NewDateTime(mc types.ModuleConfig) types.Module {
	config := newDateTimeConfig(mc)
	done := make(chan struct{})
	update := make(chan []*types.Block)
	dt := &DateTime{update: update, done: done, config: config}

	ticker := time.NewTicker(dt.config.refresh)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				update <- dt.MakeBlocks()
			}
		}
	}()

	return dt
}

// MakeBlocks returns the Block array for this module
func (dt *DateTime) MakeBlocks() []*types.Block {
	b := make([]*types.Block, 0)
	if dt.config.label != "" {
		block := types.NewBlock()
		block.FullText = dt.config.label
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
	return dt.update
}

// Stop stops this module and prevents new Block arrays from being sent
func (dt *DateTime) Stop() {
	close(dt.done)
}
