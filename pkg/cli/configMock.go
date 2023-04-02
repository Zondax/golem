package cli

type ConfigMock struct {
	Foo string `json:"foo"`
}

func (c ConfigMock) SetDefaults() {
}

func (c ConfigMock) Validate() error {
	return nil
}
