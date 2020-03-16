package types

// Block is an i3 block as defined here: https://i3wm.org/docs/i3bar-protocol.html
type Block struct {
	FullText     string `json:"full_text"`
	ShortText    string `json:"short_text,omitempty"`
	Color        string `json:"color,omitempty"`
	Background   string `json:"background,omitempty"`
	Border       string `json:"border,omitempty"`
	BorderTop    int    `json:"border_top,omitempty"`
	BorderRight  int    `json:"border_right,omitempty"`
	BorderBottom int    `json:"border_bottom,omitempty"`
	BorderLeft   int    `json:"border_left,omitempty"`
	MinWidth     string `json:"min_width,omitempty"`
	Align        string `json:"align,omitempty"`
	Urgent       bool   `json:"urgent,omitempty"`
	Name         string `json:"name,omitempty"`
	Instance     string `json:"instance,omitempty"`
	//this is a pointer to work around the default true case
	Separator           *bool  `json:"separator,omitempty"`
	SeparatorBlockWidth int    `json:"separator_block_width,omitempty"`
	Markup              string `json:"markup,omitempty"`
}

// NewBlock returns a new Block
func NewBlock(sepWidth int) *Block {
	f := new(bool)
	*f = false
	return &Block{
		Separator:           f,
		SeparatorBlockWidth: sepWidth,
	}
}

// AddSeparator sets the separator to nil
// (this is because the separator defaults to true)
func (b *Block) AddSeparator() {
	b.Separator = nil
}

// RemoveSeparator sets the separator false
func (b *Block) RemoveSeparator() {
	f := new(bool)
	*f = false
	b.Separator = f
}
