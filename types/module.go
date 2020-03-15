package types

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
	Label          string
	Refresh        int
	Separator      bool
	SeparatorWidth int
}
