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
	update chan []*types.Block
	done   chan struct{}
	config *shellCommandConfig
}

type shellCommandConfig struct {
	label   string
	refresh time.Duration
	cmd     string
}

func newShellCommandConfig(mc types.ModuleConfig) *shellCommandConfig {
	label, ok := mc["label"].(string)
	if !ok {
		label = ""
	}

	refresh, ok := mc["refresh"].(int)
	if !ok {
		refresh = 1000
	}

	cmd, ok := mc["cmd"].(string)
	if !ok {
		cmd = ""
	}

	return &shellCommandConfig{
		label:   label,
		refresh: time.Duration(refresh) * time.Millisecond,
		cmd:     cmd,
	}
}

// NewShellCommand creates a new ShellCommand, starts it's ticker, then returns it
func NewShellCommand(mc types.ModuleConfig) types.Module {
	config := newShellCommandConfig(mc)
	done := make(chan struct{})
	update := make(chan []*types.Block)
	sc := &ShellCommand{update: update, done: done, config: config}

	ticker := time.NewTicker(sc.config.refresh)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				update <- sc.MakeBlocks()
			}
		}
	}()

	return sc
}

// MakeBlocks returns the Block array for this module
func (sc *ShellCommand) MakeBlocks() []*types.Block {
	b := make([]*types.Block, 0)
	if sc.config.label != "" {
		block := types.NewBlock()
		block.FullText = sc.config.label
		block.RemoveSeparator()
		b = append(b, block)
	}

	if sc.config.cmd != "" {
		block := types.NewBlock()
		output, err := exec.Command("/bin/bash", "-c", sc.config.cmd).Output()
		if err != nil {
			block.FullText = err.Error()
			return b
		}

		block.FullText = strings.TrimSpace(string(output))
	}

	return b
}

// GetUpdateChan returns the channel down which new Block arrays are sent
func (sc *ShellCommand) GetUpdateChan() chan []*types.Block {
	return sc.update
}

// Stop stops this module and prevents new Block arrays from being sent
func (sc *ShellCommand) Stop() {
	close(sc.done)
}
