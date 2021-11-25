// Powergoline
//
// A lightweight status line for your terminal emulator. This project aims to
// be a lightweight alternative for powerline a popular statusline plugin for
// VIm that statuslines and prompts for several other applications, including
// zsh, bash, tmux, IPython, Awesome and Qtile.
//
//   - https://cixtor.com/
//   - https://github.com/cixtor/powergoline
//   - https://en.wikipedia.org/wiki/Status_bar
//
// A status bar is a graphical control element which poses an information area
// typically found at the window's bottom. It can be divided into sections to
// group information. Its job is primarily to display information about the
// current state of its window, although some status bars have extra
// functionality.
//
// A status bar can also be text-based, primarily in console-based applications,
// in which case it is usually the last row in an 80x25 text mode configuration,
// leaving the top 24 rows for application data. Usually the status bar (called
// a status line in this context) displays the current state of the application,
// as well as helpful keyboard shortcuts.
package main

import (
	"flag"
	"os"
)

// config is the user-provided configuration.
var config Config

var themes = map[string]func(Config) Config{
	"agnoster":   ApplyAgnosterTheme,
	"astrocom":   ApplyAstrocomTheme,
	"bluescale":  ApplyBlueScaleTheme,
	"colorish":   ApplyColorishTheme,
	"grayscale":  ApplyGrayScaleTheme,
	"wildcherry": ApplyWildCherryTheme,
}

func main() {
	flag.BoolVar(&config.Debug, "debug", false, "Prints plugin runtime statistics")
	flag.BoolVar(&config.TimeOn, "time.on", false, "Prints date and time, use -time.fmt to format")
	flag.IntVar(&config.TimeFg, "time.fg", 255, "Defines the date and time foreground color")
	flag.IntVar(&config.TimeBg, "time.bg", 13, "Defines the date and time background color")
	flag.StringVar(&config.TimeFmt, "time.fmt", "2006-01-02 15:04:05", "Defines the date and time segment format")
	flag.BoolVar(&config.UserOn, "user.on", false, "Prints the current username")
	flag.IntVar(&config.UserFg, "user.fg", 255, "Defines the username foreground color")
	flag.IntVar(&config.UserBg, "user.bg", 33, "Defines the username background color")
	flag.BoolVar(&config.HostOn, "host.on", false, "Prints the current hostname")
	flag.IntVar(&config.HostFg, "host.fg", 255, "Defines the hostname foreground color")
	flag.IntVar(&config.HostBg, "host.bg", 75, "Defines the hostname background color")
	flag.IntVar(&config.HomeFg, "home.fg", 255, "Defines the home directory foreground color")
	flag.IntVar(&config.HomeBg, "home.bg", 105, "Defines the home directory background color")
	flag.IntVar(&config.RodirFg, "rodir.fg", 255, "Defines the read-only directory foreground color")
	flag.IntVar(&config.RodirBg, "rodir.bg", 124, "Defines the read-only directory background color")
	flag.IntVar(&config.CwdN, "cwd.n", 1, "Defines how many folder levels to print")
	flag.BoolVar(&config.CwdOn, "cwd.on", true, "Prints the current working directory")
	flag.IntVar(&config.CwdFg, "cwd.fg", 255, "Defines the current working directory foreground color")
	flag.IntVar(&config.CwdBg, "cwd.bg", 99, "Defines the current working directory background color")
	flag.BoolVar(&config.RepoOn, "repo.on", false, "Prints the Git/Mercurial/Subversion status")
	flag.IntVar(&config.RepoFg, "repo.fg", 0, "Defines the repository status foreground color")
	flag.IntVar(&config.RepoBg, "repo.bg", 255, "Defines the repository status background color")
	flag.Var(&config.RepoExclude, "repo.exclude", "Sets repo.on=false for the specified folder")
	flag.Var(&config.RepoInclude, "repo.include", "Sets repo.on=true for the specified folder")
	flag.Var(&config.Plugins, "plugin", "Defines a plugin with optional arguments (e.g. -plugin=\"echo hello world\")\nDefine multiple plugins like this: -plugin=A -plugin=B -plugin=C")
	flag.IntVar(&config.PluginFg, "plugin.fg", 0, "Defines the plugin output foreground color")
	flag.IntVar(&config.PluginBg, "plugin.bg", 11, "Defines the plugin output background color")
	flag.StringVar(&config.SymbolRoot, "symbol.root", "#", "Defines the prompt symbol for the Root user session")
	flag.StringVar(&config.SymbolUser, "symbol.user", "$", "Defines the prompt symbol for a Regular user session")
	flag.IntVar(&config.StatusFg, "status.fg", 255, "Defines the program exit status foreground color")
	flag.IntVar(&config.StatusCode, "status.code", -1, "Exit status code of the most recent program execution")
	flag.IntVar(&config.StatusSuccess, "status.success", 41, "Defines the background color for exit(0)\nOperation success and generic status code.")
	flag.IntVar(&config.StatusError, "status.error", 1, "Defines the background color for exit(1)\nCatchall for general errors and failures.")
	flag.IntVar(&config.StatusMisuse, "status.misuse", 3, "Defines the background color for exit(2)\nMisuse of shell builtins, missing command or permission problem.")
	flag.IntVar(&config.StatusCantExec, "status.cantexec", 4, "Defines the background color for exit(126)\nCannot execute command, permission problem, or not an executable.")
	flag.IntVar(&config.StatusNotFound, "status.notfound", 14, "Defines the background color for exit(127)\nCommand not found, illegal path, or possible typo.")
	flag.IntVar(&config.StatusInvalid, "status.invalid", 202, "Defines the background color for exit(128)\nInvalid argument to exit, only use range 0-255.")
	flag.IntVar(&config.StatusErrSignal, "status.errsignal", 8, "Defines the background color for exit(128+n)\nFatal error signal where \"n\" is the PID.")
	flag.IntVar(&config.StatusTerminated, "status.terminated", 13, "Defines the background color for exit(130)\nScript terminated by Control-C.")
	flag.IntVar(&config.StatusOutofrange, "status.outofrange", 0, "Defines the background color for exit(255*)\nExit status out of range.")
	flag.StringVar(&config.Theme, "theme", "", "Automatic configuration based on a predefined color scheme")

	flag.Parse()

	os.Exit(NewPowergoline(config).Render(os.Stdout))
}
