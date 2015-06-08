package main

import "fmt"
import "os"
import "strings"

type PowerGoLine struct {
}

func (pogol PowerGoLine) Print(text string, fg string, bg string) {
	var color_seq string

	// Print foreground color.
	if fg != "" && bg != "" {
		color_seq += fmt.Sprintf("38;5;%s", fg)
		color_seq += fmt.Sprintf(";")
		color_seq += fmt.Sprintf("48;5;%s", bg)
	} else if fg != "" {
		color_seq += fmt.Sprintf("38;5;%s", fg)
	} else if bg != "" {
		color_seq += fmt.Sprintf("48;5;%s", bg)
	}

	// Add color sequences if necessary.
	if len(color_seq) > 0 {
		fmt.Printf("\\[\\e[%sm\\]", color_seq)
		fmt.Printf("%s", text)
		fmt.Printf("\\[\\e[0m\\]")
	} else {
		fmt.Printf("%s", text)
	}
}

func (pogol PowerGoLine) ExitColor(pcolor PowerColor, status string) string {
	var extcolor string

	/**
	 * System Status Codes.
	 *
	 * 0     - Operation success and generic status code.
	 * 1     - Catchall for general errors and failures.
	 * 2     - Misuse of shell builtins, missing command, or permission problem.
	 * 126   - Command invoked cannot execute, permission problem,
	 *         or the command is not an executable binary.
	 * 127   - Command not found, illegal path, or possible typo.
	 * 128   - Invalid argument to exit, only use range 0-255.
	 * 128+n - Fatal error signal where "n" is the PID.
	 * 130   - Script terminated by Control-C.
	 * 255*  - Exit status out of range.
	 */

	if status == "0" {
		extcolor = pcolor.Status.Success
	} else if status == "1" {
		extcolor = pcolor.Status.Failure
	} else if status == "126" {
		extcolor = pcolor.Status.Permission
	} else if status == "127" {
		extcolor = pcolor.Status.NotFound
	} else if status == "128" {
		extcolor = pcolor.Status.InvalidExit
	} else if status == "130" {
		extcolor = pcolor.Status.Terminated
	} else {
		extcolor = pcolor.Status.Misuse
	}

	return extcolor
}

func (pogol PowerGoLine) TermTitle() {
	fmt.Printf("\\[\\e]0;\\u@\\h: \\w\\a\\]")
}

func (pogol PowerGoLine) Username(pcolor PowerColor) {
	var username string = os.Getenv("USERNAME")

	var fg string = pcolor.UsernameFg
	var bg string = pcolor.UsernameBg
	var hbg string = pcolor.HostnameBg

	username = fmt.Sprintf(" %s ", username)
	pogol.Print(username, fg, bg)
	pogol.Print("\uE0B0", bg, hbg)
}

func (pogol PowerGoLine) Hostname(pcolor PowerColor) {
	hostname, err := os.Hostname()

	var fg string = pcolor.HostnameFg
	var bg string = pcolor.HostnameBg
	var hbg string = pcolor.HomeDirectoryBg

	if err != nil {
		hostname = "localhost"
	}

	hostname = fmt.Sprintf(" %s ", hostname)
	pogol.Print(hostname, fg, bg)
	pogol.Print("\uE0B0", bg, hbg)
}

func (pogol PowerGoLine) WorkingDirectory(pcolor PowerColor, status string) {
	var homedir string = os.Getenv("HOME")
	var workingdir string = os.Getenv("PWD")
	var shortdir string = strings.Replace(workingdir, homedir, "", -1)
	var cleandir string = strings.Trim(shortdir, "/")
	var extcolor string = pogol.ExitColor(pcolor, status)

	// Get configured colors.
	var home_fg string = pcolor.HomeDirectoryFg
	var home_bg string = pcolor.HomeDirectoryBg
	var wd_fg string = pcolor.WorkingDirectoryFg
	var wd_bg string = pcolor.WorkingDirectoryBg

	// Print the user home directory path.
	pogol.Print(" ~ ", home_fg, home_bg)

	if cleandir == "" {
		pogol.Print("\uE0B0", home_bg, extcolor)
	} else {
		pogol.Print("\uE0B0", home_bg, wd_bg)
	}

	// Print the sequence of folders of the current path.
	var maxsegms int = 4
	var segments []string = strings.Split(cleandir, "/")
	var ttlsegms int = len(segments)
	var lastsegm int = (ttlsegms - 1)

	if ttlsegms > maxsegms {
		var newsegms []string = make([]string, 0)

		newsegms = append(newsegms, segments[0])
		newsegms = append(newsegms, "\u2026")
		newsegms = append(newsegms, segments[lastsegm-2])
		newsegms = append(newsegms, segments[lastsegm-1])
		newsegms = append(newsegms, segments[lastsegm])

		segments = newsegms
		lastsegm = maxsegms
	}

	for key, folder := range segments {
		if folder != "" {
			folder = fmt.Sprintf(" %s ", folder)
			pogol.Print(folder, wd_fg, wd_bg)

			if key == lastsegm {
				pogol.Print("\uE0B0", wd_bg, extcolor)
			} else {
				pogol.Print("\uE0B1", wd_fg, wd_bg)
			}
		}
	}
}

func (pogol PowerGoLine) RootSymbol(pcolor PowerColor, status string) {
	var symbol string
	var uid int = os.Getuid()
	var extcolor string = pogol.ExitColor(pcolor, status)
	var fg string = pcolor.Status.Symbol

	if uid == 0 {
		symbol = pcolor.Symbol.SuperUser
	} else {
		symbol = pcolor.Symbol.Regular
	}

	symbol = fmt.Sprintf(" %s ", symbol)
	pogol.Print(symbol, fg, extcolor)
	pogol.Print("\uE0B0", extcolor, "")
	pogol.Print("\u0020\n", "", "")
}
