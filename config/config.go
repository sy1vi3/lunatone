package config

type configDefinition struct {
	Settings settings
}

type settings struct {
	DragoURL     string `toml:"drago_url"`
	DragoAuth    string `toml:"drago_auth"`
	ExcludeAreas []int  `toml:"exclude"`
	EnableHour   int    `toml:"enable_hour"`
	DisableHour  int    `toml:"disable_hour"`
}

var Config = configDefinition{
	Settings: settings{},
}
