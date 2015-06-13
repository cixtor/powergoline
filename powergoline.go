package main

import "fmt"
import "os"
import "path"
import "strings"
import "time"

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

func (pogol PowerGoLine) IsRdonlyDir(folder string) bool {
	var temp_path string = path.Join(folder, temp_file)
	_, err := os.Create(temp_path)

	if err != nil {
		return true
	}

	os.Remove(temp_path)
	return false
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

func (pogol PowerGoLine) DateTime(pcolor PowerColor) {
	if pcolor.Datetime.Status == "enabled" {
		var date_time string = time.Now().Format("15:04:05")
		var fg string = pcolor.Datetime.Foreground
		var bg string = pcolor.Datetime.Background
		var ubg string = pcolor.Username.Background

		date_time = fmt.Sprintf(" %s ", date_time)
		pogol.Print(date_time, fg, bg)
		pogol.Print("\uE0B2", ubg, bg)
	}
}

func (pogol PowerGoLine) Username(pcolor PowerColor) {
	if pcolor.Username.Status == "enabled" {
		var fg string = pcolor.Username.Foreground
		var bg string = pcolor.Username.Background
		var hbg string = pcolor.Hostname.Background

		if pcolor.Hostname.Status != "enabled" {
			hbg = pcolor.HomeDirectoryBg
		}

		pogol.Print(" \\u ", fg, bg)
		pogol.Print("\uE0B0", bg, hbg)
	}
}

func (pogol PowerGoLine) Hostname(pcolor PowerColor) {
	if pcolor.Hostname.Status == "enabled" {
		var fg string = pcolor.Hostname.Foreground
		var bg string = pcolor.Hostname.Background
		var hbg string = pcolor.HomeDirectoryBg

		pogol.Print(" \\h ", fg, bg)
		pogol.Print("\uE0B0", bg, hbg)
	}
}

func (pogol PowerGoLine) WorkingDirectory(pcolor PowerColor, status string) {
	var homedir string = os.Getenv("HOME")
	var workingdir string = os.Getenv("PWD")
	var shortdir string = strings.Replace(workingdir, homedir, "", -1)
	var cleandir string = strings.Trim(shortdir, "/")
	var extcolor string = pogol.ExitColor(pcolor, status)
	var is_rdonly_dir bool = pogol.IsRdonlyDir(workingdir)

	// Get configured colors.
	var home_fg string = pcolor.HomeDirectoryFg
	var home_bg string = pcolor.HomeDirectoryBg
	var wd_fg string = pcolor.WorkingDirectoryFg
	var wd_bg string = pcolor.WorkingDirectoryBg
	var rd_fg string = pcolor.RdonlyDirectoryFg
	var rd_bg string = pcolor.RdonlyDirectoryBg

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

	// Draw each directory segment with right arrow.
	for key, folder := range segments {
		if folder != "" {
			folder = fmt.Sprintf(" %s ", folder)
			pogol.Print(folder, wd_fg, wd_bg)

			if key == lastsegm {
				// Draw last arrow and read-only lock.
				if is_rdonly_dir == true {
					pogol.Print("\uE0B0", wd_bg, rd_bg)
					pogol.Print(" \uE0A2 ", rd_fg, rd_bg)
					pogol.Print("\uE0B0", rd_bg, extcolor)
				} else {
					pogol.Print("\uE0B0", wd_bg, extcolor)
				}
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
