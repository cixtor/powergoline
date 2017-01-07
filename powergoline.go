package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

const enabled string = "enabled"

// PowerGoLine holds the configuration either defined by the current user in the
// TTY session or the default settings defined by the program on startup. It
// also holds the bytes that will be printed in the command line prompt in the
// form of segments.
type PowerGoLine struct {
	Segments []Segment
	Config   PowerColor
}

// Segment represents one single part in the command line prompt. Each segment
// contains the text and color for the foreground and background of that text.
// Notice that most segments have a spacing on the left and right side to keep
// things in shape.
type Segment struct {
	Text       string
	Foreground string
	Background string
}

// AddSegment inserts a new block in the CLI prompt output.
func (pogol *PowerGoLine) AddSegment(text string, fg string, bg string) {
	var segment Segment

	segment.Text = text
	segment.Foreground = fg
	segment.Background = bg

	pogol.Segments = append(pogol.Segments, segment)
}

// Print sends a segment to the standard output.
func (pogol PowerGoLine) Print(text string, fg string, bg string) {
	var colorSeq string

	// Add the foreground and background colors.
	if fg != "" && bg != "" {
		colorSeq += fmt.Sprintf("38;5;%s", fg)
		colorSeq += fmt.Sprintf(";")
		colorSeq += fmt.Sprintf("48;5;%s", bg)
	} else if fg != "" {
		colorSeq += fmt.Sprintf("38;5;%s", fg)
	} else if bg != "" {
		colorSeq += fmt.Sprintf("48;5;%s", bg)
	}

	// Draw the color sequences if necessary.
	if len(colorSeq) > 0 {
		fmt.Printf("\\[\\e[%sm\\]", colorSeq)
		fmt.Printf("%s", text)
		fmt.Printf("\\[\\e[0m\\]")
	} else {
		fmt.Printf("%s", text)
	}
}

// PrintStatusLine sends all the segments to the standard output.
func (pogol PowerGoLine) PrintStatusLine() {
	var key int
	var current Segment
	var nextsegm Segment

	ttlsegms := len(pogol.Segments)

	for key = 0; key < ttlsegms; key++ {
		current = pogol.Segments[key]

		if current.Background == "automatic" {
			nextsegm = pogol.Segments[key+1]
			current.Background = nextsegm.Background
		}

		// Escape subshell expressions to prevent arbitrary code execution.
		current.Text = strings.Replace(current.Text, "$", "\\$", -1)
		current.Text = strings.Replace(current.Text, "`", "\\`", -1)

		pogol.Print(current.Text,
			current.Foreground,
			current.Background)
	}

	fmt.Printf("\u0020\n")
}

// IsWritable checks if the process can write in a directory.
func (pogol PowerGoLine) IsWritable(folder string) bool {
	return unix.Access(folder, unix.W_OK) == nil
}

// IsRdonlyDir checks if a directory is read only by the current user.
func (pogol PowerGoLine) IsRdonlyDir(folder string) bool {
	return !pogol.IsWritable(folder)
}

// ExitColor determines the color for the result of each previous command.
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

// TermTitle defines the template for the terminal title.
func (pogol *PowerGoLine) TermTitle() {
	pogol.AddSegment("\\[\\e]0;\\u@\\h: \\w\\a\\]", "", "")
}

// DateTime defines a segment with the current date and time.
func (pogol *PowerGoLine) DateTime() {
	if pogol.Config.Datetime.Status == enabled {
		pogol.AddSegment("\x20"+time.Now().Format("15:04:05")+"\x20",
			pogol.Config.Datetime.Foreground,
			pogol.Config.Datetime.Background)

		pogol.AddSegment("\uE0B0",
			pogol.Config.Datetime.Background,
			pogol.Config.Username.Background)
	}
}

// Username defines a segment with the name of the current account.
func (pogol *PowerGoLine) Username() {
	if pogol.Config.Username.Status == enabled {
		pogol.AddSegment(" \\u ",
			pogol.Config.Username.Foreground,
			pogol.Config.Username.Background)

		pogol.AddSegment("\uE0B0",
			pogol.Config.Username.Background,
			"automatic")
	}
}

// Hostname defines a segment with the name of this system.
func (pogol *PowerGoLine) Hostname() {
	if pogol.Config.Hostname.Status == enabled {
		pogol.AddSegment(" \\h ",
			pogol.Config.Hostname.Foreground,
			pogol.Config.Hostname.Background)

		pogol.AddSegment("\uE0B0",
			pogol.Config.Hostname.Background,
			"automatic")
	}
}

// HomeDirectory defines a segment with current directory path.
func (pogol *PowerGoLine) HomeDirectory() {
	pogol.AddSegment(" ~ ",
		pogol.Config.Directory.HomeDirectoryFg,
		pogol.Config.Directory.HomeDirectoryBg)

	pogol.AddSegment("\uE0B0",
		pogol.Config.Directory.HomeDirectoryBg,
		"automatic")
}

// WorkingDirectory returns the full path of the current directory.
func (pogol *PowerGoLine) WorkingDirectory() {
	homedir := os.Getenv("HOME")
	workingdir := os.Getenv("PWD")
	shortdir := strings.Replace(workingdir, homedir, "", 1)
	cleandir := strings.Trim(shortdir, "/")

	// Draw the sequence of folders of the current path.
	dirparts := strings.Split(cleandir, "/")
	ttlparts := len(dirparts)
	lastsegm := (ttlparts - 1)

	// Determine the maximum number of directory segments.
	maxsegms, _ := strconv.Atoi(pogol.Config.Directory.MaximumSegments)
	if maxsegms < 1 {
		maxsegms = 1
	}

	if ttlparts > maxsegms {
		newparts := make([]string, 0)
		offset := (maxsegms - 1)
		newparts = append(newparts, "\u2026")
		for k := offset; k >= 0; k-- {
			newparts = append(newparts, dirparts[lastsegm-k])
		}
		dirparts = newparts
		lastsegm = maxsegms
	}

	// Print home directory segment if necessary.
	if strings.Index(workingdir, homedir) == 0 {
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
	if pogol.IsRdonlyDir(workingdir) {
		pogol.AddSegment(" \uE0A2 ",
			pogol.Config.Directory.RdonlyDirectoryFg,
			pogol.Config.Directory.RdonlyDirectoryBg)

		pogol.AddSegment("\uE0B0",
			pogol.Config.Directory.RdonlyDirectoryBg,
			"automatic")
	}
}

// GitInformation defines a segment with information of a Git repository.
func (pogol *PowerGoLine) GitInformation() {
	if pogol.Config.Repository.Git.Status == enabled {
		var extbin ExtBinary

		branch, _ := extbin.GitBranch()

		if branch != nil {
			extra, err := extbin.GitStatusExtra()
			branchName := fmt.Sprintf(" \uE0A0 %s ", branch)
			foreground := pogol.Config.Repository.Git.Foreground
			background := pogol.Config.Repository.Git.Background

			if err == nil {
				if extra.Committed {
					background = pogol.Config.Repository.Git.CommittedBg
				} else if extra.Untracked {
					background = pogol.Config.Repository.Git.UntrackedBg
				}

				if extra.AheadCommits > 0 {
					branchName += fmt.Sprintf("\u21E1%d ", extra.AheadCommits)
				} else if extra.BehindCommits > 0 {
					branchName += fmt.Sprintf("\u21E3%d ", extra.BehindCommits)
				}

				status, err := extbin.GitStatus()

				if err == nil {
					if status["modified"] > 0 {
						branchName += fmt.Sprintf("~%d ", status["modified"])
					}

					if status["added"] > 0 {
						branchName += fmt.Sprintf("+%d ", status["added"])
					}

					if status["deleted"] > 0 {
						branchName += fmt.Sprintf("-%d ", status["deleted"])
					}
				}
			}

			pogol.AddSegment(branchName, foreground, background)
			pogol.AddSegment("\uE0B0", background, "automatic")
		}
	}
}

// MercurialInformation defines a segment with information of a Mercurial repository.
func (pogol *PowerGoLine) MercurialInformation() {
	if pogol.Config.Repository.Mercurial.Status == enabled {
		var extbin ExtBinary

		branch, _ := extbin.MercurialBranch()

		if branch != nil {
			status, err := extbin.MercurialStatus()
			branchName := fmt.Sprintf(" \uE0A0 %s ", branch)

			if err == nil {
				if status["modified"] > 0 {
					branchName += fmt.Sprintf("~%d ", status["modified"])
				}

				if status["added"] > 0 {
					branchName += fmt.Sprintf("+%d ", status["added"])
				}

				if status["deleted"] > 0 {
					branchName += fmt.Sprintf("-%d ", status["deleted"])
				}
			}

			pogol.AddSegment(branchName,
				pogol.Config.Repository.Mercurial.Foreground,
				pogol.Config.Repository.Mercurial.Background)

			pogol.AddSegment("\uE0B0",
				pogol.Config.Repository.Mercurial.Background,
				"automatic")
		}
	}
}

// RootSymbol defines a segment with an indicator for root users.
func (pogol *PowerGoLine) RootSymbol(status string) {
	var symbol string

	uid := os.Getuid()
	extcolor := pogol.ExitColor(pogol.Config, status)

	if uid == 0 {
		symbol = pogol.Config.Symbol.SuperUser
	} else {
		symbol = pogol.Config.Symbol.Regular
	}

	symbol = fmt.Sprintf(" %s ", symbol)
	pogol.AddSegment(symbol, pogol.Config.Status.Symbol, extcolor)
	pogol.AddSegment("\uE0B0", extcolor, "")
}
