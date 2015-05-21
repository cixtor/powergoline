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
