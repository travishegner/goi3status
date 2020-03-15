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
