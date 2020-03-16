package types

import "time"

// ModuleConfig is the config of any given module, already unmarshalled
type ModuleConfig map[interface{}]interface{}

// CreateModule is a type to represent a generic "NewModule" function
type CreateModule func(ModuleConfig) Module

// Module represents a collection of i3 blocks to display _something_ on the i3 bar
type Module interface {
	MakeBlocks() []*Block
	GetUpdateChan() chan []*Block
	Stop()
}

// BaseModule contains the attributes common to all modules
type BaseModule struct {
	Update chan []*Block
	Done   chan struct{}
}

// BaseModuleConfig contains the attributes common to all module configs
type BaseModuleConfig struct {
	Label               string
	Refresh             time.Duration
	FinalSeparator      bool
	FinalSeparatorWidth int
	BlockSeparatorWidth int
}

// NewBaseModuleConfig parses and returns a BaseModuleConfig
func NewBaseModuleConfig(mc ModuleConfig) *BaseModuleConfig {
	label, ok := mc["label"].(string)
	if !ok {
		label = ""
	}

	refresh, ok := mc["refresh"].(int)
	if !ok {
		refresh = 1000
	}

	fseparator, ok := mc["final_separator"].(bool)
	if !ok {
		fseparator = true
	}

	fsepWidth, ok := mc["final_separator_width"].(int)
	if !ok {
		fsepWidth = 0
	}

	bsepWidth, ok := mc["block_separator_width"].(int)
	if !ok {
		bsepWidth = 0
	}

	return &BaseModuleConfig{
		Label:               label,
		Refresh:             time.Duration(refresh) * time.Millisecond,
		FinalSeparator:      fseparator,
		FinalSeparatorWidth: fsepWidth,
		BlockSeparatorWidth: bsepWidth,
	}
}

// NewBaseModule creates a new BaseModule
func NewBaseModule() *BaseModule {
	done := make(chan struct{})
	update := make(chan []*Block, 1)
	return &BaseModule{
		Update: update,
		Done:   done,
	}
}
