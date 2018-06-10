package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// Config holds the CLI prompt settings.
type Config struct {
	filename string
	values   PowerColor
}

// PowerColor is the base of the JSON object.
type PowerColor struct {
	Username   StandardConfig   `json:"username"`
	Hostname   StandardConfig   `json:"hostname"`
	Directory  DirectoryConfig  `json:"directory"`
	Status     StatusCode       `json:"status"`
	Symbol     StatusSymbol     `json:"symbol"`
	Datetime   StandardConfig   `json:"datetime"`
	Repository RepositoryConfig `json:"repository"`
	Plugins    []Plugin         `json:"plugins"`
}

// StandardConfig is a generic text and color object.
type StandardConfig struct {
	Status     string `json:"status"`
	Foreground string `json:"foreground"`
	Background string `json:"background"`
}

// DirectoryConfig holds the settings for the directory segment.
type DirectoryConfig struct {
	MaximumSegments    string `json:"maximum_segments"`
	HomeDirectoryFg    string `json:"home_directory_fg"`
	HomeDirectoryBg    string `json:"home_directory_bg"`
	WorkingDirectoryFg string `json:"working_directory_fg"`
	WorkingDirectoryBg string `json:"working_directory_bg"`
	RdonlyDirectoryFg  string `json:"rdonly_directory_fg"`
	RdonlyDirectoryBg  string `json:"rdonly_directory_bg"`
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

// Plugin adds support for execution of external commands.
type Plugin struct {
	Command    string `json:"command"`
	Background string `json:"background"`
	Foreground string `json:"foreground"`
}

// RepositoryExtraConfig adds additional settings.
type RepositoryExtraConfig struct {
	StandardConfig
	CommittedBg string `json:"committed_bg"`
	UntrackedBg string `json:"untracked_bg"`
}

// RepositoryConfig holds the settings for the repositories.
type RepositoryConfig struct {
	Git       RepositoryExtraConfig `json:"git"`
	Mercurial RepositoryExtraConfig `json:"mercurial"`
}

// NewConfig creates a new instance of Config.
func NewConfig(filename string) *Config {
	var config Config

	config.filename = filename

	values, err := config.Values()

	if err != nil {
		log.Println("powergoline;", err)
		/* do not return; continue */
	}

	config.values = values

	return &config
}

// exists checks if the configuration file exists.
func (config Config) exists() bool {
	_, err := os.Stat(config.filename)
	return err == nil
}

// Default returns an object with the default configuration.
func (config Config) Default() PowerColor {
	var pcolor PowerColor

	pcolor.Username.Status = usernameStatus
	pcolor.Username.Foreground = usernameForeground
	pcolor.Username.Background = usernameBackground

	pcolor.Hostname.Status = hostnameStatus
	pcolor.Hostname.Foreground = hostnameForeground
	pcolor.Hostname.Background = hostnameBackground

	pcolor.Directory.MaximumSegments = maximumSegments
	pcolor.Directory.HomeDirectoryFg = homeDirForeground
	pcolor.Directory.HomeDirectoryBg = homeDirBackground
	pcolor.Directory.WorkingDirectoryFg = workingDirForeground
	pcolor.Directory.WorkingDirectoryBg = workingDirBackground
	pcolor.Directory.RdonlyDirectoryFg = readOnlyDirForeground
	pcolor.Directory.RdonlyDirectoryBg = readOnlyDirBackground

	pcolor.Status.Symbol = statusSymbol
	pcolor.Status.Success = statusSuccess
	pcolor.Status.Failure = statusFailure
	pcolor.Status.Misuse = statusMisuse
	pcolor.Status.Permission = statusPermission
	pcolor.Status.NotFound = statusNotFound
	pcolor.Status.InvalidExit = statusInvalidExit
	pcolor.Status.Terminated = statusTerminated

	pcolor.Symbol.Regular = symbolRegular
	pcolor.Symbol.SuperUser = symbolSuperUser

	pcolor.Datetime.Status = datetimeStatus
	pcolor.Datetime.Foreground = datetimeForeground
	pcolor.Datetime.Background = datetimeBackground

	pcolor.Repository.Git.Status = repositoryStatus
	pcolor.Repository.Git.Foreground = repositoryForeground
	pcolor.Repository.Git.Background = repositoryBackground
	pcolor.Repository.Git.CommittedBg = repositoryCommittedBG
	pcolor.Repository.Git.UntrackedBg = repositoryUntrackedBG

	pcolor.Repository.Mercurial.Status = repositoryStatus
	pcolor.Repository.Mercurial.Foreground = repositoryForeground
	pcolor.Repository.Mercurial.Background = repositoryBackground
	pcolor.Repository.Git.CommittedBg = repositoryCommittedBG
	pcolor.Repository.Git.UntrackedBg = repositoryUntrackedBG

	return pcolor
}

// Values returns the settings from the configuration file.
func (config Config) Values() (PowerColor, error) {
	if config.exists() {
		return config.ExistingValues(config.filename)
	}

	return config.NonExistingValues(config.filename)
}

// ExistingValues returns the configuration from a local file.
func (config Config) ExistingValues(path string) (PowerColor, error) {
	file, err := os.Open(path)

	if err != nil {
		return config.Default(), err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Println("file.close;", err)
		}
	}()

	var data PowerColor

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return config.Default(), err
	}

	return data, nil
}

// NonExistingValues creates the configuration file using default values.
func (config Config) NonExistingValues(filename string) (PowerColor, error) {
	data, err := json.MarshalIndent(config.Default(), "", "\x20\x20")

	if err != nil {
		return config.Default(), err
	}

	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return config.Default(), err
	}

	return config.Default(), nil
}
