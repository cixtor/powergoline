package main

import "fmt"
import "os"
import "path/filepath"
import "strings"
import "time"

type PowerGoLine struct {
	Segments []Segment
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

func (pogol *PowerGoLine) TermTitle() {
	pogol.AddSegment("\\[\\e]0;\\u@\\h: \\w\\a\\]", "", "")
}

func (pogol *PowerGoLine) DateTime(pcolor PowerColor) {
	if pcolor.Datetime.Status == "enabled" {
		var date_time string = time.Now().Format("15:04:05")
		date_time = fmt.Sprintf(" %s ", date_time)

		pogol.AddSegment(date_time,
			pcolor.Datetime.Foreground,
			pcolor.Datetime.Background)

		pogol.AddSegment("\uE0B2",
			pcolor.Username.Background,
			"automatic")
	}
}

func (pogol *PowerGoLine) Username(pcolor PowerColor) {
	if pcolor.Username.Status == "enabled" {
		pogol.AddSegment(" \\u ",
			pcolor.Username.Foreground,
			pcolor.Username.Background)

		pogol.AddSegment("\uE0B0",
			pcolor.Username.Background,
			"automatic")
	}
}

func (pogol *PowerGoLine) Hostname(pcolor PowerColor) {
	if pcolor.Hostname.Status == "enabled" {
		pogol.AddSegment(" \\h ",
			pcolor.Hostname.Foreground,
			pcolor.Hostname.Background)

		pogol.AddSegment("\uE0B0",
			pcolor.Hostname.Background,
			"automatic")
	}
}

func (pogol *PowerGoLine) HomeDirectory(pcolor PowerColor) {
	pogol.AddSegment(" ~ ",
		pcolor.Directory.HomeDirectoryFg,
		pcolor.Directory.HomeDirectoryBg)

	pogol.AddSegment("\uE0B0",
		pcolor.Directory.HomeDirectoryBg,
		"automatic")
}

func (pogol *PowerGoLine) WorkingDirectory(pcolor PowerColor) {
	var homedir string = os.Getenv("HOME")
	var workingdir string = os.Getenv("PWD")
	var shortdir string = strings.Replace(workingdir, homedir, "", 1)
	var cleandir string = strings.Trim(shortdir, "/")
	var is_rdonly_dir bool = pogol.IsRdonlyDir(workingdir)

	// Draw the sequence of folders of the current path.
	var maxsegms int = 4
	var dirparts []string = strings.Split(cleandir, "/")
	var ttlparts int = len(dirparts)
	var lastsegm int = (ttlparts - 1)

	if ttlparts > maxsegms {
		var newparts []string = make([]string, 0)

		newparts = append(newparts, dirparts[0])
		newparts = append(newparts, "\u2026")
		newparts = append(newparts, dirparts[lastsegm-2])
		newparts = append(newparts, dirparts[lastsegm-1])
		newparts = append(newparts, dirparts[lastsegm])

		dirparts = newparts
		lastsegm = maxsegms
	}

	// Draw each directory segment with right arrow.
	for key, folder := range dirparts {
		if folder != "" {
			folder = fmt.Sprintf(" %s ", folder)
			pogol.AddSegment(folder,
				pcolor.Directory.WorkingDirectoryFg,
				pcolor.Directory.WorkingDirectoryBg)

			if key == lastsegm {
				pogol.AddSegment("\uE0B0",
					pcolor.Directory.WorkingDirectoryBg,
					"automatic")
			} else {
				pogol.AddSegment("\uE0B1",
					pcolor.Directory.WorkingDirectoryFg,
					pcolor.Directory.WorkingDirectoryBg)
			}
		}
	}

	// Draw lock if current directory is read-only.
	if is_rdonly_dir == true {
		pogol.AddSegment(" \uE0A2 ",
			pcolor.Directory.RdonlyDirectoryFg,
			pcolor.Directory.RdonlyDirectoryBg)

		pogol.AddSegment("\uE0B0",
			pcolor.Directory.RdonlyDirectoryBg,
			"automatic")
	}
}

func (pogol *PowerGoLine) MercurialInformation(pcolor PowerColor) {
	if pcolor.Repository.Mercurial.Status == "enabled" {
		var extbin ExtBinary
		branch, _ := extbin.MercurialBranch()

		if branch != nil {
			var branch_str string = fmt.Sprintf(" %s ", branch)
			pogol.AddSegment(branch_str,
				pcolor.Repository.Mercurial.Foreground,
				pcolor.Repository.Mercurial.Background)

			pogol.AddSegment("\uE0B0",
				pcolor.Repository.Mercurial.Background,
				"automatic")
		}
	}
}

func (pogol *PowerGoLine) RootSymbol(pcolor PowerColor, status string) {
	var symbol string
	var uid int = os.Getuid()
	var extcolor string = pogol.ExitColor(pcolor, status)

	if uid == 0 {
		symbol = pcolor.Symbol.SuperUser
	} else {
		symbol = pcolor.Symbol.Regular
	}

	symbol = fmt.Sprintf(" %s ", symbol)
	pogol.AddSegment(symbol, pcolor.Status.Symbol, extcolor)
	pogol.AddSegment("\uE0B0", extcolor, "")
}
