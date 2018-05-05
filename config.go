package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Configuration holds the CLI prompt settings.
type Configuration struct{}

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

// Path returns the full path of the configuration directory.
func (config Configuration) Path() string {
	return os.Getenv("HOME") + "/" + configPath
}

// Exists checks if the configuration file exists.
func (config Configuration) Exists() bool {
	path := config.Path()
	_, err := os.Stat(path)
	return err == nil
}

// Default returns an object with the default configuration.
func (config Configuration) Default() PowerColor {
	var pcolor PowerColor

	pcolor.Username.Status = usernameST
	pcolor.Username.Foreground = usernameFG
	pcolor.Username.Background = usernameBG

	pcolor.Hostname.Status = hostnameST
	pcolor.Hostname.Foreground = hostnameFG
	pcolor.Hostname.Background = hostnameBG

	pcolor.Directory.MaximumSegments = maximumSegments
	pcolor.Directory.HomeDirectoryFg = homeDirFG
	pcolor.Directory.HomeDirectoryBg = homeDirBG
	pcolor.Directory.WorkingDirectoryFg = workingDirFG
	pcolor.Directory.WorkingDirectoryBg = workingDirBG
	pcolor.Directory.RdonlyDirectoryFg = rdonlyDirFG
	pcolor.Directory.RdonlyDirectoryBg = rdonlyDirBG

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

	pcolor.Datetime.Status = datetimeST
	pcolor.Datetime.Foreground = datetimeFG
	pcolor.Datetime.Background = datetimeBG

	pcolor.Repository.Git.Status = repositoryST
	pcolor.Repository.Git.Foreground = repositoryFG
	pcolor.Repository.Git.Background = repositoryBG
	pcolor.Repository.Git.CommittedBg = repositoryCommittedBG
	pcolor.Repository.Git.UntrackedBg = repositoryUntrackedBG

	pcolor.Repository.Mercurial.Status = repositoryST
	pcolor.Repository.Mercurial.Foreground = repositoryFG
	pcolor.Repository.Mercurial.Background = repositoryBG
	pcolor.Repository.Git.CommittedBg = repositoryCommittedBG
	pcolor.Repository.Git.UntrackedBg = repositoryUntrackedBG

	return pcolor
}

// DefaultJSON returns a JSON-encoded string with the default configuration.
func (config Configuration) DefaultJSON() ([]byte, error) {
	pcolor := config.Default()

	text, err := json.MarshalIndent(pcolor, "", "\x20\x20")

	if err != nil {
		return nil, err
	}

	return text, nil
}

// Values returns the settings from the configuration file.
func (config Configuration) Values() PowerColor {
	path := config.Path()

	// Try to read the file if it exists.
	if config.Exists() {
		file, err := os.Open(path)

		if err != nil {
			fmt.Println(err) /* log error */
			return config.Default()
		}

		defer file.Close()

		var data PowerColor

		if err := json.NewDecoder(file).Decode(&data); err != nil {
			fmt.Println(err) /* log error */
			return config.Default()
		}

		return data
	}

	// Create and use the default configuration.
	file, err := os.Create(path)

	if err != nil {
		fmt.Println(err) /* log error */
		return config.Default()
	}

	defer file.Close()

	text, err := config.DefaultJSON()

	if err != nil {
		fmt.Println(err) /* log error */
		return config.Default()
	}

	if _, err := file.Write(text); err != nil {
		fmt.Println(err) /* log error */
		return config.Default()
	}

	return config.Default()
}
