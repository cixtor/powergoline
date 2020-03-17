package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Config holds the CLI prompt settings.
type Config struct {
	Datetime   SimpleConfig     `json:"datetime"`
	Username   SimpleConfig     `json:"username"`
	Hostname   SimpleConfig     `json:"hostname"`
	HomeDir    ColorsConfig     `json:"homedir"`
	RdonlyDir  ColorsConfig     `json:"rdonlydir"`
	CurrentDir CurrentDirectory `json:"currentdir"`
	Repository RepositoryConfig `json:"repository"`
	Plugins    []Plugin         `json:"plugins"`
	Symbol     StatusSymbol     `json:"symbol"`
	Status     StatusCode       `json:"status"`
}

// ColorsConfig is the foreground and background colors.
type ColorsConfig struct {
	Fg string `json:"foreground"`
	Bg string `json:"background"`
}

// SimpleConfig is a generic text and color object.
type SimpleConfig struct {
	On bool `json:"enabled"`
	ColorsConfig
}

type RepositoryConfig struct {
	SimpleConfig
	Exclude []string `json:"exclude,omitempty"`
	Include []string `json:"include,omitempty"`
}

// CurrentDirectory is the configuration for the current working directory.
type CurrentDirectory struct {
	Size int `json:"size"`
	ColorsConfig
}

// Plugin adds support for execution of external commands.
type Plugin struct {
	Command string `json:"command"`
	ColorsConfig
}

// StatusCode holds the settings for the program exit codes.
type StatusCode struct {
	Symbol      string `json:"symbol"`
	Success     string `json:"success"`
	Failure     string `json:"failure"`
	Misuse      string `json:"misuse"`
	Permission  string `json:"permission"`
	NotFound    string `json:"not_found"`
	InvalidExit string `json:"invalid_exit"`
	Terminated  string `json:"terminated"`
}

// StatusSymbol holds the indicator for each user.
type StatusSymbol struct {
	Regular   string `json:"regular"`
	SuperUser string `json:"super_user"`
}

// NewConfig creates a new instance of Config.
//
// The function attempts to read and decode`$HOME/.powergoline.json`
//
// If the configuration file does not exist, then it returns the default values
// and attempts to write the default values into the aforementioned file for
// future reads. If the file exists but contains malformed data, it returns the
// default values and displays a warning to explain the file load issues.
func NewConfig(filename string) (Config, error) {
	var config Config

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		config = defaultConfig()
		data, _ := json.MarshalIndent(config, "", "\t")
		_ = ioutil.WriteFile(filename, data, 0644)
		return config, nil
	}

	file, err := os.Open(filename)

	if err != nil {
		return defaultConfig(), fmt.Errorf(program+"; open config %s", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(program+"; exit config %s", err)
		}
	}()

	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return defaultConfig(), fmt.Errorf(program+"; read config %s", err)
	}

	return config, nil
}
