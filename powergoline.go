package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type PowerGoLine struct {
	Segments []Segment
	Config   PowerColor
}

type Segment struct {
	Text       string
	Foreground string
	Background string
}

func (pogol *PowerGoLine) AddSegment(text string, fg string, bg string) {
	var segment Segment

	segment.Text = text
	segment.Foreground = fg
	segment.Background = bg

	pogol.Segments = append(pogol.Segments, segment)
}

func (pogol PowerGoLine) Print(text string, fg string, bg string) {
	var color_seq string

	// Add the foreground and background colors.
	if fg != "" && bg != "" {
		color_seq += fmt.Sprintf("38;5;%s", fg)
		color_seq += fmt.Sprintf(";")
		color_seq += fmt.Sprintf("48;5;%s", bg)
	} else if fg != "" {
		color_seq += fmt.Sprintf("38;5;%s", fg)
	} else if bg != "" {
		color_seq += fmt.Sprintf("48;5;%s", bg)
	}

	// Draw the color sequences if necessary.
	if len(color_seq) > 0 {
		fmt.Printf("\\[\\e[%sm\\]", color_seq)
		fmt.Printf("%s", text)
		fmt.Printf("\\[\\e[0m\\]")
	} else {
		fmt.Printf("%s", text)
	}
}

func (pogol PowerGoLine) PrintStatusLine() {
	var key int
	var current Segment
	var nextsegm Segment
	var ttlsegms int = len(pogol.Segments)

	for key = 0; key < ttlsegms; key++ {
		current = pogol.Segments[key]

		if current.Background == "automatic" {
			nextsegm = pogol.Segments[key+1]
			current.Background = nextsegm.Background
		}

		pogol.Print(current.Text,
			current.Foreground,
			current.Background)
	}

	fmt.Printf("\u0020\n")
}

func (pogol PowerGoLine) IsRdonlyDir(folder string) bool {
	var temp_path string = filepath.Join(folder, temp_file)
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
		extcolor = pogol.Config.Status.Success
	} else if status == "1" {
		extcolor = pogol.Config.Status.Failure
	} else if status == "126" {
		extcolor = pogol.Config.Status.Permission
	} else if status == "127" {
		extcolor = pogol.Config.Status.NotFound
	} else if status == "128" {
		extcolor = pogol.Config.Status.InvalidExit
	} else if status == "130" {
		extcolor = pogol.Config.Status.Terminated
	} else {
		extcolor = pogol.Config.Status.Misuse
	}

	return extcolor
}

func (pogol *PowerGoLine) TermTitle() {
	pogol.AddSegment("\\[\\e]0;\\u@\\h: \\w\\a\\]", "", "")
}

func (pogol *PowerGoLine) DateTime() {
	if pogol.Config.Datetime.Status == "enabled" {
		var date_time string = time.Now().Format("15:04:05")
		date_time = fmt.Sprintf(" %s ", date_time)

		pogol.AddSegment(date_time,
			pogol.Config.Datetime.Foreground,
			pogol.Config.Datetime.Background)

		pogol.AddSegment("\uE0B0",
			pogol.Config.Datetime.Background,
			pogol.Config.Username.Background)
	}
}

func (pogol *PowerGoLine) Username() {
	if pogol.Config.Username.Status == "enabled" {
		pogol.AddSegment(" \\u ",
			pogol.Config.Username.Foreground,
			pogol.Config.Username.Background)

		pogol.AddSegment("\uE0B0",
			pogol.Config.Username.Background,
			"automatic")
	}
}

func (pogol *PowerGoLine) Hostname() {
	if pogol.Config.Hostname.Status == "enabled" {
		pogol.AddSegment(" \\h ",
			pogol.Config.Hostname.Foreground,
			pogol.Config.Hostname.Background)

		pogol.AddSegment("\uE0B0",
			pogol.Config.Hostname.Background,
			"automatic")
	}
}

func (pogol *PowerGoLine) HomeDirectory() {
	pogol.AddSegment(" ~ ",
		pogol.Config.Directory.HomeDirectoryFg,
		pogol.Config.Directory.HomeDirectoryBg)

	pogol.AddSegment("\uE0B0",
		pogol.Config.Directory.HomeDirectoryBg,
		"automatic")
}

func (pogol *PowerGoLine) WorkingDirectory() {
	var homedir string = os.Getenv("HOME")
	var workingdir string = os.Getenv("PWD")
	var shortdir string = strings.Replace(workingdir, homedir, "", 1)
	var cleandir string = strings.Trim(shortdir, "/")
	var is_rdonly_dir bool = pogol.IsRdonlyDir(workingdir)
	var print_home_dir int = strings.Index(workingdir, homedir)

	// Draw the sequence of folders of the current path.
	var dirparts []string = strings.Split(cleandir, "/")
	var ttlparts int = len(dirparts)
	var lastsegm int = (ttlparts - 1)

	// Determine the maximum number of directory segments.
	maxsegms, _ := strconv.Atoi(pogol.Config.Directory.MaximumSegments)
	if maxsegms < 1 {
		maxsegms = 1
	}

	if ttlparts > maxsegms {
		var newparts []string = make([]string, 0)
		var offset int = (maxsegms - 1)
		newparts = append(newparts, "\u2026")
		for k := offset; k >= 0; k-- {
			newparts = append(newparts, dirparts[lastsegm-k])
		}
		dirparts = newparts
		lastsegm = maxsegms
	}

	// Print home directory segment if necessary.
	if print_home_dir == 0 {
		pogol.HomeDirectory()
	}

	// Draw each directory segment with right arrow.
	for key, folder := range dirparts {
		if folder != "" {
			folder = fmt.Sprintf(" %s ", folder)
			pogol.AddSegment(folder,
				pogol.Config.Directory.WorkingDirectoryFg,
				pogol.Config.Directory.WorkingDirectoryBg)

			if key == lastsegm {
				pogol.AddSegment("\uE0B0",
					pogol.Config.Directory.WorkingDirectoryBg,
					"automatic")
			} else {
				pogol.AddSegment("\uE0B1",
					pogol.Config.Directory.WorkingDirectoryFg,
					pogol.Config.Directory.WorkingDirectoryBg)
			}
		}
	}

	// Draw lock if current directory is read-only.
	if is_rdonly_dir == true {
		pogol.AddSegment(" \uE0A2 ",
			pogol.Config.Directory.RdonlyDirectoryFg,
			pogol.Config.Directory.RdonlyDirectoryBg)

		pogol.AddSegment("\uE0B0",
			pogol.Config.Directory.RdonlyDirectoryBg,
			"automatic")
	}
}

func (pogol *PowerGoLine) GitInformation() {
	if pogol.Config.Repository.Git.Status == "enabled" {
		var extbin ExtBinary
		branch, _ := extbin.GitBranch()

		if branch != nil {
			status, err := extbin.GitStatus()
			var branch_str string = fmt.Sprintf(" \uE0A0 %s ", branch)

			if err == nil {
				if status["modified"] > 0 {
					branch_str += fmt.Sprintf("~%d ", status["modified"])
				}

				if status["added"] > 0 {
					branch_str += fmt.Sprintf("+%d ", status["added"])
				}

				if status["deleted"] > 0 {
					branch_str += fmt.Sprintf("-%d ", status["deleted"])
				}
			}

			extra, err := extbin.GitStatusExtra()
			var foreground string = pogol.Config.Repository.Git.Foreground
			var background string = pogol.Config.Repository.Git.Background

			if err == nil {
				if extra["committed"] {
					background = pogol.Config.Repository.Git.CommittedBg
				} else if extra["untracked"] {
					background = pogol.Config.Repository.Git.UntrackedBg
				}
			}

			pogol.AddSegment(branch_str, foreground, background)
			pogol.AddSegment("\uE0B0", background, "automatic")
		}
	}
}

func (pogol *PowerGoLine) MercurialInformation() {
	if pogol.Config.Repository.Mercurial.Status == "enabled" {
		var extbin ExtBinary
		branch, _ := extbin.MercurialBranch()

		if branch != nil {
			status, err := extbin.MercurialStatus()
			var branch_str string = fmt.Sprintf(" \uE0A0 %s ", branch)

			if err == nil {
				if status["modified"] > 0 {
					branch_str += fmt.Sprintf("~%d ", status["modified"])
				}

				if status["added"] > 0 {
					branch_str += fmt.Sprintf("+%d ", status["added"])
				}

				if status["deleted"] > 0 {
					branch_str += fmt.Sprintf("-%d ", status["deleted"])
				}
			}

			pogol.AddSegment(branch_str,
				pogol.Config.Repository.Mercurial.Foreground,
				pogol.Config.Repository.Mercurial.Background)

			pogol.AddSegment("\uE0B0",
				pogol.Config.Repository.Mercurial.Background,
				"automatic")
		}
	}
}

func (pogol *PowerGoLine) RootSymbol(status string) {
	var symbol string
	var uid int = os.Getuid()
	var extcolor string = pogol.ExitColor(pogol.Config, status)

	if uid == 0 {
		symbol = pogol.Config.Symbol.SuperUser
	} else {
		symbol = pogol.Config.Symbol.Regular
	}

	symbol = fmt.Sprintf(" %s ", symbol)
	pogol.AddSegment(symbol, pogol.Config.Status.Symbol, extcolor)
	pogol.AddSegment("\uE0B0", extcolor, "")
}
