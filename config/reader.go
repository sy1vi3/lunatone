package config

import (
	"io"
	"os"

	"github.com/pelletier/go-toml/v2"
)

func ReadConfig() {
	tomlFile, err := os.Open("config.toml")
	// if we os.Open returns an error then handle it
	if err != nil {
		panic(err)
	}
	// defer the closing of our tomlFile so that we can parse it later on
	//goland:noinspection GoUnhandledErrorResult
	defer tomlFile.Close()

	byteValue, _ := io.ReadAll(tomlFile)

	err = toml.Unmarshal(byteValue, &Config)
	if err != nil {
		panic(err)
	}
}
