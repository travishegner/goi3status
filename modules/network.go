package modules

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
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

	return &networkConfig{
		BaseModuleConfig: bmc,
		Interface:        iface,
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

			upDiff := humanize.IBytes(s.BytesSent - n.lastUpBytes)
			dnDiff := humanize.IBytes(s.BytesRecv - n.lastDownBytes)

			block.FullText = fmt.Sprintf("%s up - %s dn", upDiff, dnDiff)

			b = append(b, block)
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
