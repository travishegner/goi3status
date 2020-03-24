package modules

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/net"
	log "github.com/sirupsen/logrus"
	"github.com/travishegner/goi3status/types"
)

func init() {
	addModMap("Network", NewNetwork)
}

type networkConfig struct {
	*types.BaseModuleConfig
	Interface string
	Attribute string
	DownSpeed int
	UpSpeed   int
}

// Network is a module representing the named interface
type Network struct {
	*types.BaseModule
	config        *networkConfig
	lastUpBytes   uint64
	lastDownBytes uint64
}

func newNetworkConfig(mc types.ModuleConfig) *networkConfig {
	bmc := types.NewBaseModuleConfig(mc)
	iface, ok := mc["interface"].(string)
	if !ok {
		iface = "all"
	}

	attribute, ok := mc["attribute"].(string)
	if !ok {
		attribute = "down"
	}

	dnspd, ok := mc["down_speed"].(int)
	if !ok {
		dnspd = 1000
	}

	upspd, ok := mc["up_speed"].(int)
	if !ok {
		upspd = 1000
	}

	return &networkConfig{
		BaseModuleConfig: bmc,
		Interface:        iface,
		Attribute:        attribute,
		DownSpeed:        dnspd,
		UpSpeed:          upspd,
	}

}

// NewNetwork returns the Network module
func NewNetwork(mc types.ModuleConfig) types.Module {
	config := newNetworkConfig(mc)
	bm := types.NewBaseModule()
	n := &Network{
		BaseModule: bm,
		config:     config,
	}

	bm.Update <- n.MakeBlocks()
	ticker := time.NewTicker(n.config.Refresh)

	go func() {
		for {
			select {
			case <-n.Done:
				return
			case <-ticker.C:
				bm.Update <- n.MakeBlocks()
			}
		}
	}()

	return n
}

// MakeBlocks returns the Block array for this module
func (n *Network) MakeBlocks() []*types.Block {
	b := make([]*types.Block, 0)
	if n.config.Label != "" {
		block := types.NewBlock(n.config.BlockSeparatorWidth)
		block.FullText = n.config.Label
		b = append(b, block)
	}

	pernic := true
	if n.config.Interface == "all" {
		pernic = false
	}

	stats, err := net.IOCounters(pernic)
	if err != nil {
		log.Errorf("failed to get network stats: %v", err.Error())
		return b
	}

	for _, s := range stats {
		if !pernic || s.Name == n.config.Interface {
			block := types.NewBlock(n.config.BlockSeparatorWidth)

			switch n.config.Attribute {
			case "down":
				br := s.BytesRecv
				spd := (br - n.lastDownBytes) * 8 / 1000000
				if n.lastDownBytes != 0 {
					block.FullText = fmt.Sprintf("%v\u2193", spd)
					block.Color = GetColor(float64(spd) / float64(n.config.DownSpeed))
				}
				n.lastDownBytes = br

			case "up":
				bs := s.BytesSent
				spd := (bs - n.lastUpBytes) * 8 / 1000000
				if n.lastUpBytes != 0 {
					block.FullText = fmt.Sprintf("%v\u2191", spd)
					block.Color = GetColor(float64(spd) / float64(n.config.UpSpeed))
				}
				n.lastUpBytes = bs
			}

			b = append(b, block)

			break
		}
	}

	return b
}

// GetUpdateChan returns the channel down which new block arrays are sent
func (n *Network) GetUpdateChan() chan []*types.Block {
	return n.Update
}

// Stop stops this module from polling and sending updated Block arrays
func (n *Network) Stop() {
	close(n.Done)
}
