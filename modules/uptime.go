package modules

import (
	"time"

	"github.com/davidscholberg/go-durationfmt"
	"github.com/shirou/gopsutil/host"
	log "github.com/sirupsen/logrus"
	"github.com/travishegner/goi3status/types"
)

func init() {
	addModMap("Uptime", NewUptime)
}

type uptimeConfig struct {
	*types.BaseModuleConfig
	format string
}

// Uptime is a module representing the machines uptime
type Uptime struct {
	*types.BaseModule
	config *uptimeConfig
}

func newUptimeConfig(mc types.ModuleConfig) *uptimeConfig {
	bmc := types.NewBaseModuleConfig(mc)

	format, ok := mc["format"].(string)
	if !ok {
		format = "deafult"
	}

	return &uptimeConfig{
		BaseModuleConfig: bmc,
		format:           format,
	}
}

// NewUptime returns the Uptime module
func NewUptime(mc types.ModuleConfig) types.Module {
	config := newUptimeConfig(mc)
	bm := types.NewBaseModule()
	m := &Uptime{
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
func (u *Uptime) MakeBlocks() []*types.Block {
	b := make([]*types.Block, 0)
	if u.config.Label != "" {
		block := types.NewBlock()
		block.FullText = u.config.Label
		block.RemoveSeparator()
		b = append(b, block)
	}

	block := types.NewBlock()

	ut, err := host.Uptime()
	if err != nil {
		log.Warningf("failed to get host uptime: %v", err.Error())
		return b
	}

	d := time.Duration(int64(ut)) * time.Second
	format := u.config.format
	if u.config.format == "default" {
		format = getFormat(d)
	}

	block.FullText, err = durationfmt.Format(d, format)
	if err != nil {
		log.Errorf("error parsing uptime format: %v", err.Error())
		block.FullText = err.Error()
	}
	b = append(b, block)
	return b
}

// GetUpdateChan returns the channel down which new block arrays are sent
func (u *Uptime) GetUpdateChan() chan []*types.Block {
	return u.Update
}

// Stop stops this module from polling and sending updated Block arrays
func (u *Uptime) Stop() {
	close(u.Done)
}

func getFormat(d time.Duration) string {
	yearStr, _ := durationfmt.Format(d, "%yy")
	if yearStr != "0y" {
		return "%yy%ww%dd%hh%mm%ss"
	}
	weekStr, _ := durationfmt.Format(d, "%ww")
	if weekStr != "0w" {
		return "%ww%dd%hh%mm%ss"
	}
	dayStr, _ := durationfmt.Format(d, "%dd")
	if dayStr != "0d" {
		return "%dd%hh%mm%ss"
	}
	hourStr, _ := durationfmt.Format(d, "%hh")
	if hourStr != "0h" {
		return "%hh%mm%ss"
	}
	return "%mm%ss"
}
