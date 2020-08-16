/**
 * Powergoline
 * https://cixtor.com/
 * https://github.com/cixtor/powergoline
 * https://en.wikipedia.org/wiki/Status_bar
 *
 * A status bar is a graphical control element which poses an information area
 * typically found at the window's bottom. It can be divided into sections to
 * group information. Its job is primarily to display information about the
 * current state of its window, although some status bars have extra
 * functionality.
 *
 * A status bar can also be text-based, primarily in console-based applications,
 * in which case it is usually the last row in an 80x25 text mode configuration,
 * leaving the top 24 rows for application data. Usually the status bar (called
 * a status line in this context) displays the current state of the application,
 * as well as helpful keyboard shortcuts.
 */

package main

import (
	"flag"
	"os"
)

// program is the canonical program name.
const program = "powergoline"

// config is the user-provided configuration.
var config Config

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
	flag.IntVar(&config.RepoFg, "repo.fg", 0, "Defines the foreground color")
	flag.IntVar(&config.RepoBg, "repo.bg", 255, "Defines the background color")
	flag.Var(&config.RepoExclude, "repo.exclude", "Sets repo.on=false for this folder")
	flag.Var(&config.RepoInclude, "repo.include", "Sets repo.on=true for this folder")
	flag.Var(&config.Plugins, "plugin", "Defines a plugin (e.g. -plugin=\"fg:255,bg:26,cmd=echo hello world\")")
	flag.StringVar(&config.SymbolRoot, "symbol.root", "#", "Defines the root prompt symbol")
	flag.StringVar(&config.SymbolUser, "symbol.user", "$", "Defines the user prompt symbol")
	flag.IntVar(&config.StatusFg, "status.fg", 255, "Defines the exit status foreground color")
	flag.IntVar(&config.StatusCode, "status.code", -1, "Exit code of the most recent command")
	flag.IntVar(&config.StatusSuccess, "status.success", 41, "exit(0) Successful operationDefine the background color")
	flag.IntVar(&config.StatusError, "status.error", 1, "exit(1) Catchall for general errors")
	flag.IntVar(&config.StatusMisuse, "status.misuse", 3, "exit(2) Misuse of shell builtins (according to Bash documentation)")
	flag.IntVar(&config.StatusCantExec, "status.cantexec", 4, "exit(126) Command invoked cannot execute")
	flag.IntVar(&config.StatusNotFound, "status.notfound", 14, "exit(127) \"command not found\"")
	flag.IntVar(&config.StatusInvalid, "status.invalid", 202, "exit(128) Invalid argument to exit")
	flag.IntVar(&config.StatusErrSignal, "status.errsignal", 8, "exit(128+n) Fatal error signal \"n\"")
	flag.IntVar(&config.StatusTerminated, "status.terminated", 13, "exit(130) Script terminated by Control-C")
	flag.IntVar(&config.StatusOutofrange, "status.outofrange", 0, "exit(255*) Exit status out of range")

	flag.Parse()

	os.Exit(NewPowergoline(config).Render())
}
