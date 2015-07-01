package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// External configuration file path.
const config_path string = ".powergoline.json"
const temp_file string = "_tempfile_0f060643f7.txt"

// Define default username values.
const username_status string = "enabled"
const username_foreground string = "255"
const username_background string = "006"

// Define default hostname values.
const hostname_status string = "enabled"
const hostname_foreground string = "255"
const hostname_background string = "012"

// Define default working directory values.
const maximum_segments string = "2"
const home_directory_fg string = "255"
const home_directory_bg string = "161"
const working_directory_fg string = "251"
const working_directory_bg string = "238"
const rdonly_directory_fg string = "255"
const rdonly_directory_bg string = "124"

// Define default datetime values.
const datetime_status string = "disabled"
const datetime_foreground string = "255"
const datetime_background string = "023"

// Define default status symbols.
const symbol_regular string = "$"
const symbol_super_user string = "#"

// Status code default colors.
const status_symbol string = "255"
const status_success string = "070"
const status_failure string = "001"
const status_misuse string = "003"
const status_permission string = "004"
const status_not_found string = "014"
const status_invalid_exit string = "202"
const status_terminated string = "013"

// CVS default colors.
const repository_status string = "enabled"
const repository_foreground string = "000"
const repository_background string = "148"
const repository_committed_bg string = "214"
const repository_untracked_bg string = "175"

type Configuration struct {
}

type PowerColor struct {
	Username   StandardConfig   `json:"username"`
	Hostname   StandardConfig   `json:"hostname"`
	Directory  DirectoryConfig  `json:"directory"`
	Status     StatusCode       `json:"status"`
	Symbol     StatusSymbol     `json:"symbol"`
	Datetime   StandardConfig   `json:"datetime"`
	Repository RepositoryConfig `json:"repository"`
}

type StandardConfig struct {
	Status     string `json:"status"`
	Foreground string `json:"foreground"`
	Background string `json:"background"`
}

type DirectoryConfig struct {
	MaximumSegments    string `json:"maximum_segments"`
	HomeDirectoryFg    string `json:"home_directory_fg"`
	HomeDirectoryBg    string `json:"home_directory_bg"`
	WorkingDirectoryFg string `json:"working_directory_fg"`
	WorkingDirectoryBg string `json:"working_directory_bg"`
	RdonlyDirectoryFg  string `json:"rdonly_directory_fg"`
	RdonlyDirectoryBg  string `json:"rdonly_directory_bg"`
}

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

type StatusSymbol struct {
	Regular   string `json:"regular"`
	SuperUser string `json:"super_user"`
}

type RepositoryExtraConfig struct {
	StandardConfig
	CommittedBg string `json:"committed_bg"`
	UntrackedBg string `json:"untracked_bg"`
}

type RepositoryConfig struct {
	Git       RepositoryExtraConfig `json:"git"`
	Mercurial RepositoryExtraConfig `json:"mercurial"`
}

func (config Configuration) Path() string {
	var homedir string = os.Getenv("HOME")

	return fmt.Sprintf("%s/%s", homedir, config_path)
}

func (config Configuration) Exists() bool {
	var path string = config.Path()
	_, err := os.Stat(path)

	if err != nil {
		return false
	}

	return true
}

func (config Configuration) Default() PowerColor {
	var pcolor PowerColor

	pcolor.Username.Status = username_status
	pcolor.Username.Foreground = username_foreground
	pcolor.Username.Background = username_background

	pcolor.Hostname.Status = hostname_status
	pcolor.Hostname.Foreground = hostname_foreground
	pcolor.Hostname.Background = hostname_background

	pcolor.Directory.MaximumSegments = maximum_segments
	pcolor.Directory.HomeDirectoryFg = home_directory_fg
	pcolor.Directory.HomeDirectoryBg = home_directory_bg
	pcolor.Directory.WorkingDirectoryFg = working_directory_fg
	pcolor.Directory.WorkingDirectoryBg = working_directory_bg
	pcolor.Directory.RdonlyDirectoryFg = rdonly_directory_fg
	pcolor.Directory.RdonlyDirectoryBg = rdonly_directory_bg

	pcolor.Status.Symbol = status_symbol
	pcolor.Status.Success = status_success
	pcolor.Status.Failure = status_failure
	pcolor.Status.Misuse = status_misuse
	pcolor.Status.Permission = status_permission
	pcolor.Status.NotFound = status_not_found
	pcolor.Status.InvalidExit = status_invalid_exit
	pcolor.Status.Terminated = status_terminated

	pcolor.Symbol.Regular = symbol_regular
	pcolor.Symbol.SuperUser = symbol_super_user

	pcolor.Datetime.Status = datetime_status
	pcolor.Datetime.Foreground = datetime_foreground
	pcolor.Datetime.Background = datetime_background

	pcolor.Repository.Git.Status = repository_status
	pcolor.Repository.Git.Foreground = repository_foreground
	pcolor.Repository.Git.Background = repository_background
	pcolor.Repository.Git.CommittedBg = repository_committed_bg
	pcolor.Repository.Git.UntrackedBg = repository_untracked_bg

	pcolor.Repository.Mercurial.Status = repository_status
	pcolor.Repository.Mercurial.Foreground = repository_foreground
	pcolor.Repository.Mercurial.Background = repository_background
	pcolor.Repository.Git.CommittedBg = repository_committed_bg
	pcolor.Repository.Git.UntrackedBg = repository_untracked_bg

	return pcolor
}

func (config Configuration) DefaultJson() ([]byte, error) {
	var pcolor PowerColor = config.Default()
	json_str, err := json.MarshalIndent(pcolor, "", "    ")

	if err != nil {
		return nil, err
	}

	return json_str, nil
}

func (config Configuration) Values() PowerColor {
	var path string = config.Path()
	var exists bool = config.Exists()

	// Create and use the default configuration.
	if exists == false {
		file, err := os.Create(path)
		defer file.Close()

		if err == nil {
			json_str, err := config.DefaultJson()

			if err == nil {
				file.Write(json_str)
			}

			return config.Default()
		}
	}

	// Read the external configuration.
	file, err := os.Open(path)
	defer file.Close()

	if err == nil {
		var alt_pcolor PowerColor
		body, _ := ioutil.ReadAll(file)

		err = json.Unmarshal(body, &alt_pcolor)

		if err == nil {
			return alt_pcolor
		}
	}

	return config.Default()
}
