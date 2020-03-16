package modules

import (
	"os/exec"
	"strings"
	"time"

	"github.com/travishegner/goi3status/types"
)

func init() {
	addModMap("ShellCommand", NewShellCommand)
}

// ShellCommand is a module for executing shell commands and displaying the output as a block
type ShellCommand struct {
	*types.BaseModule
	config *shellCommandConfig
}

type shellCommandConfig struct {
	*types.BaseModuleConfig
	cmd string
}

func newShellCommandConfig(mc types.ModuleConfig) *shellCommandConfig {
	bmc := types.NewBaseModuleConfig(mc)
	cmd, ok := mc["cmd"].(string)
	if !ok {
		cmd = ""
	}

	return &shellCommandConfig{
		BaseModuleConfig: bmc,
		cmd:              cmd,
	}
}

// NewShellCommand creates a new ShellCommand, starts it's ticker, then returns it
func NewShellCommand(mc types.ModuleConfig) types.Module {
	config := newShellCommandConfig(mc)
	bm := types.NewBaseModule()
	sc := &ShellCommand{
		BaseModule: bm,
		config:     config,
	}

	bm.Update <- sc.MakeBlocks()
	ticker := time.NewTicker(sc.config.Refresh)

	go func() {
		for {
			select {
			case <-bm.Done:
				return
			case <-ticker.C:
				bm.Update <- sc.MakeBlocks()
			}
		}
	}()

	return sc
}

// MakeBlocks returns the Block array for this module
func (sc *ShellCommand) MakeBlocks() []*types.Block {
	b := make([]*types.Block, 0)
	if sc.config.Label != "" {
		block := types.NewBlock(sc.config.BlockSeparatorWidth)
		block.FullText = sc.config.Label
		b = append(b, block)
	}

	if sc.config.cmd != "" {
		block := types.NewBlock(sc.config.FinalSeparatorWidth)
		if sc.config.FinalSeparator {
			block.AddSeparator()
		}

		cmd := sc.config.cmd
		output, err := exec.Command("/bin/bash", "-c", cmd).Output()
		if err != nil {
			block.FullText = err.Error()
			return b
		}

		block.FullText = strings.TrimSpace(string(output))
		b = append(b, block)
	}

	return b
}

// GetUpdateChan returns the channel down which new Block arrays are sent
func (sc *ShellCommand) GetUpdateChan() chan []*types.Block {
	return sc.Update
}

// Stop stops this module and prevents new Block arrays from being sent
func (sc *ShellCommand) Stop() {
	close(sc.Done)
}
