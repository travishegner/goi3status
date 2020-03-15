package types

// Config represents the whole status config
type Config struct {
	Version     int                           `yaml:"version" json:"version"`
	StopSignal  int                           `yaml:"stop_signal" json:"stop_signal,omitempty"`
	ContSignal  int                           `yaml:"cont_signal" json:"cont_signal,omitempty"`
	ClickEvents bool                          `yaml:"click_events" json:"click_events,omitempty"`
	Modules     []map[interface{}]interface{} `yaml:"modules" json:"-"`
}
