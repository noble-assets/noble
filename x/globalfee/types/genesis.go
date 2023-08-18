package types

// NewGenesisState creates a new GenesisState object.
func NewGenesisState(params Params) *GenesisState {
	return &GenesisState{
		Params: params,
	}
}

// DefaultGenesisState creates a default GenesisState object.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams())
}

// Validate validates the provided genesis state.
func (gs *GenesisState) Validate() error {
	return gs.Params.Validate()
}
