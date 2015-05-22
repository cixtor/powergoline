package main

const config_path = ".powergoline.json"
const username_fg = "255"
const username_bg = "006"
const hostname_fg = "255"
const hostname_bg = "012"
const home_directory_fg = "255"
const home_directory_bg = "161"
const working_directory_fg = "251"
const working_directory_bg = "238"
const status_success = "070"
const status_failure = "001"
const status_misuse = "003"
const status_permission = "004"
const status_not_found = "014"
const status_invalid_exit = "008"
const status_terminated = "013"

type Configuration struct {
}

type PowerColor struct {
	UsernameFg         string      `json:"username_fg"`
	UsernameBg         string      `json:"username_bg"`
	HostnameFg         string      `json:"hostname_fg"`
	HostnameBg         string      `json:"hostname_bg"`
	HomeDirectoryFg    string      `json:"home_directory_fg"`
	HomeDirectoryBg    string      `json:"home_directory_bg"`
	WorkingDirectoryFg string      `json:"working_directory_fg"`
	WorkingDirectoryBg string      `json:"working_directory_bg"`
	Status             StatusColor `json:"status"`
}

type StatusColor struct {
	Success     string `json:"success"`
	Failure     string `json:"failure"`
	Misuse      string `json:"misuse"`
	Permission  string `json:"permission"`
	NotFound    string `json:"not_found"`
	InvalidExit string `json:"invalid_exit"`
	Terminated  string `json:"terminated"`
}

func (config Configuration) Path() string {
	var homedir string = os.Getenv("HOME")

	return fmt.Sprintf("%s/%s", homedir, config_path)
}

func (config Configuration) Exists() (bool, error) {
	var path string = config.Path()
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func (config Configuration) Default() PowerColor {
	var pcolor PowerColor

	pcolor.UsernameFg = username_fg
	pcolor.UsernameBg = username_bg
	pcolor.HostnameFg = hostname_fg
	pcolor.HostnameBg = hostname_bg
	pcolor.HomeDirectoryFg = home_directory_fg
	pcolor.HomeDirectoryBg = home_directory_bg
	pcolor.WorkingDirectoryFg = working_directory_fg
	pcolor.WorkingDirectoryBg = working_directory_bg
	pcolor.Status.Success = status_success
	pcolor.Status.Failure = status_failure
	pcolor.Status.Misuse = status_misuse
	pcolor.Status.Permission = status_permission
	pcolor.Status.NotFound = status_not_found
	pcolor.Status.InvalidExit = status_invalid_exit
	pcolor.Status.Terminated = status_terminated

	return pcolor
}
