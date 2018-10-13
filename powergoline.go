package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// errEmptyOutput defines an error when executing a command with no output.
var errEmptyOutput = errors.New("empty output")

// PowerGoLine holds the configuration either defined by the current user in
// the TTY session or the default settings defined by the program on startup.
// It also holds the bytes that will be printed in the command line prompt in
// the form of segments.
type PowerGoLine struct {
	Segments []Segment
	config   *Config
}

// Segment represents one single part in the command line prompt. Each segment
// contains the text and color for the foreground and background of that text.
// Notice that most segments have a spacing on the left and right side to keep
// things in shape.
type Segment struct {
	Text string
	Fore string
	Back string
}

// RepoStatus holds the information of the current state of a repository, this
// includes the number of untracked files, number of commits ahead from remote,
// number of commits behind compared to the state of the remote repository,
// and nothing in case the state of the local repository is the same as the
// remote version.
type RepoStatus struct {
	Branch   []byte
	Ahead    int
	Behind   int
	Added    int
	Deleted  int
	Modified int
}

// NewPowerGoLine loads the config file and instantiates PowerGoLine.
func NewPowerGoLine(filename string) *PowerGoLine {
	p := new(PowerGoLine)

	p.config = NewConfig(filename)

	return p
}

// AddSegment inserts a new block in the CLI prompt output.
func (pogol *PowerGoLine) AddSegment(text string, fg string, bg string) {
	pogol.Segments = append(pogol.Segments, Segment{
		Text: text,
		Fore: fg,
		Back: bg,
	})
}

// Print sends a segment to the standard output.
func (pogol PowerGoLine) Print(text string, fg string, bg string) {
	var color string

	// Add the foreground and background colors.
	if fg != "" && bg != "" {
		color += "38;5;" + fg + ";" + "48;5;" + bg
	} else if fg != "" {
		color += "38;5;" + fg
	} else if bg != "" {
		color += "48;5;" + bg
	}

	// Draw the color sequences if necessary.
	if len(color) > 0 {
		fmt.Print("\\[\\e[" + color + "m\\]" + text + "\\[\\e[0m\\]")
		return
	}

	fmt.Print(text)
}

// PrintStatusLine sends all the segments to the standard output.
func (pogol PowerGoLine) PrintStatusLine() {
	var curr Segment
	var next Segment

	ttlsegms := len(pogol.Segments)

	for key := 0; key < ttlsegms; key++ {
		curr = pogol.Segments[key]

		if curr.Back == "automatic" {
			next = pogol.Segments[key+1]
			curr.Back = next.Back
		}

		// prevent arbitrary code execution in subshell expressions.
		curr.Text = strings.Replace(curr.Text, "$", "\\$", -1)
		curr.Text = strings.Replace(curr.Text, "`", "\\`", -1)

		pogol.Print(curr.Text, curr.Fore, curr.Back)
	}

	fmt.Print("\u0020\n")
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
//
// System Status Codes:
//
//   0     - Operation success and generic status code.
//   1     - Catchall for general errors and failures.
//   2     - Misuse of shell builtins, missing command or permission problem.
//   126   - Cannot execute command, permission problem, or not an executable.
//   127   - Command not found, illegal path, or possible typo.
//   128   - Invalid argument to exit, only use range 0-255.
//   128+n - Fatal error signal where "n" is the PID.
//   130   - Script terminated by Control-C.
//   255*  - Exit status out of range.
func (pogol PowerGoLine) ExitColor(pcolor PowerColor, status string) string {
	var color string

	switch status {
	case "0":
		color = pogol.config.values.Status.Success
	case "1":
		color = pogol.config.values.Status.Failure
	case "126":
		color = pogol.config.values.Status.Permission
	case "127":
		color = pogol.config.values.Status.NotFound
	case "128":
		color = pogol.config.values.Status.InvalidExit
	case "130":
		color = pogol.config.values.Status.Terminated
	default:
		color = pogol.config.values.Status.Misuse
	}

	return color
}

// TermTitle defines the template for the terminal title.
func (pogol *PowerGoLine) TermTitle() {
	pogol.AddSegment("\\[\\e]0;\\u@\\h: \\w\\a\\]", "", "")
}

// DateTime defines a segment with the current date and time.
func (pogol *PowerGoLine) DateTime() {
	if pogol.config.values.Datetime.Status == enabled {
		pogol.AddSegment(
			"\x20"+time.Now().Format("15:04:05")+"\x20",
			pogol.config.values.Datetime.Foreground,
			pogol.config.values.Datetime.Background,
		)

		pogol.AddSegment(
			"\uE0B0",
			pogol.config.values.Datetime.Background,
			pogol.config.values.Username.Background,
		)
	}
}

// Username defines a segment with the name of the current account.
func (pogol *PowerGoLine) Username() {
	if pogol.config.values.Username.Status == enabled {
		pogol.AddSegment(
			"\x20\\u\x20",
			pogol.config.values.Username.Foreground,
			pogol.config.values.Username.Background,
		)

		pogol.AddSegment(
			"\uE0B0",
			pogol.config.values.Username.Background,
			"automatic",
		)
	}
}

// Hostname defines a segment with the name of this system.
func (pogol *PowerGoLine) Hostname() {
	if pogol.config.values.Hostname.Status == enabled {
		pogol.AddSegment(
			"\x20\\h\x20",
			pogol.config.values.Hostname.Foreground,
			pogol.config.values.Hostname.Background,
		)

		pogol.AddSegment(
			"\uE0B0",
			pogol.config.values.Hostname.Background,
			"automatic",
		)
	}
}

// HomeDirectory defines a segment with current directory path.
func (pogol *PowerGoLine) HomeDirectory() {
	pogol.AddSegment(
		"\x20~\x20",
		pogol.config.values.Directory.HomeDirectoryFg,
		pogol.config.values.Directory.HomeDirectoryBg,
	)

	pogol.AddSegment(
		"\uE0B0",
		pogol.config.values.Directory.HomeDirectoryBg,
		"automatic",
	)
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
	maxsegms, _ := strconv.Atoi(pogol.config.values.Directory.MaximumSegments)
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
		if folder == "" {
			continue
		}

		pogol.AddSegment(
			"\x20"+folder+"\x20",
			pogol.config.values.Directory.WorkingDirectoryFg,
			pogol.config.values.Directory.WorkingDirectoryBg,
		)

		if key == lastsegm {
			pogol.AddSegment(
				"\uE0B0",
				pogol.config.values.Directory.WorkingDirectoryBg,
				"automatic",
			)
		} else {
			pogol.AddSegment(
				"\uE0B1",
				pogol.config.values.Directory.WorkingDirectoryFg,
				pogol.config.values.Directory.WorkingDirectoryBg,
			)
		}
	}

	// Draw lock if current directory is read-only.
	if pogol.IsRdonlyDir(workingdir) {
		pogol.AddSegment(
			"\x20\uE0A2\x20",
			pogol.config.values.Directory.RdonlyDirectoryFg,
			pogol.config.values.Directory.RdonlyDirectoryBg,
		)

		pogol.AddSegment(
			"\uE0B0",
			pogol.config.values.Directory.RdonlyDirectoryBg,
			"automatic",
		)
	}
}

// GitInformation defines a segment with information of a Git repository.
func (pogol *PowerGoLine) GitInformation() {
	if pogol.config.values.Repository.Status != enabled {
		return
	}

	// check if a repository exists in the current folder.
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		return
	}

	status, err := repoStatusGit()

	if err != nil {
		/* git is not installed */
		/* not a git repository */
		return
	}

	branch := fmt.Sprintf(" \uE0A0 %s ", status.Branch)

	if status.Ahead > 0 {
		branch += fmt.Sprintf("\u21E1%d ", status.Ahead)
	}

	if status.Behind > 0 {
		branch += fmt.Sprintf("\u21E3%d ", status.Behind)
	}

	if status.Added > 0 {
		branch += fmt.Sprintf("+%d ", status.Added)
	}

	if status.Modified > 0 {
		branch += fmt.Sprintf("~%d ", status.Modified)
	}

	if status.Deleted > 0 {
		branch += fmt.Sprintf("-%d ", status.Deleted)
	}

	pogol.AddSegment(
		branch,
		pogol.config.values.Repository.Foreground,
		pogol.config.values.Repository.Background,
	)
	pogol.AddSegment(
		"\uE0B0",
		pogol.config.values.Repository.Background,
		"automatic",
	)
}

// MercurialInformation defines a segment with information of a Mercurial repository.
func (pogol *PowerGoLine) MercurialInformation() {
	if pogol.config.values.Repository.Status != enabled {
		return
	}

	// check if a repository exists in the current folder.
	if _, err := os.Stat(".hg"); os.IsNotExist(err) {
		return
	}

	status, err := repoStatusMercurial()

	if err != nil {
		/* mercurial is not installed */
		/* not a mercurial repository */
		return
	}

	branch := fmt.Sprintf(" \uE0A0 %s ", status.Branch)

	if status.Ahead > 0 {
		branch += fmt.Sprintf("\u21E1%d ", status.Ahead)
	}

	if status.Behind > 0 {
		branch += fmt.Sprintf("\u21E3%d ", status.Behind)
	}

	if status.Added > 0 {
		branch += fmt.Sprintf("+%d ", status.Added)
	}

	if status.Modified > 0 {
		branch += fmt.Sprintf("~%d ", status.Modified)
	}

	if status.Deleted > 0 {
		branch += fmt.Sprintf("-%d ", status.Deleted)
	}

	pogol.AddSegment(
		branch,
		pogol.config.values.Repository.Foreground,
		pogol.config.values.Repository.Background,
	)

	pogol.AddSegment(
		"\uE0B0",
		pogol.config.values.Repository.Background,
		"automatic",
	)
}

// ExecuteAllPlugins runs all the user defined plugins.
func (pogol *PowerGoLine) ExecuteAllPlugins() {
	for _, metadata := range pogol.config.values.Plugins {
		pogol.ExecutePlugin(metadata)
	}
}

// ExecutePlugin runs an user defined external command.
func (pogol *PowerGoLine) ExecutePlugin(p Plugin) {
	output, err := call(p.Command)

	if err == errEmptyOutput {
		/* no output */
		return
	}

	if err != nil {
		/* use error message instead */
		output = []byte(err.Error())
	}

	pogol.AddSegment("\x20"+string(output)+"\x20", p.Foreground, p.Background)
	pogol.AddSegment("\uE0B0", p.Background, "automatic")
}

// RootSymbol defines a segment with an indicator for root users.
func (pogol *PowerGoLine) RootSymbol(status string) {
	var symbol string

	extcolor := pogol.ExitColor(pogol.config.values, status)

	if os.Getuid() == 0 {
		symbol = pogol.config.values.Symbol.SuperUser
	} else {
		symbol = pogol.config.values.Symbol.Regular
	}

	pogol.AddSegment("\x20"+symbol+"\x20", pogol.config.values.Status.Symbol, extcolor)
	pogol.AddSegment("\uE0B0", extcolor, "")
}

// runcmd executes an external command and returns the output.
func call(name string, arg ...string) ([]byte, error) {
	out, err := exec.Command(name, arg...).CombinedOutput() // #nosec

	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return nil, errEmptyOutput
	}

	return bytes.Trim(out, "\n"), nil
}

// repoStatusGit returns information about the current state of a Git repository.
func repoStatusGit() (RepoStatus, error) {
	out, err := call("git", "status", "--branch", "--porcelain", "--ignore-submodules")

	if err != nil {
		return RepoStatus{}, err
	}

	return repoStatusGitParse(bytes.Split(out, []byte("\n")))
}

// repoStatusGitParse parses the output of the `git status` command.
//
//   ## master...origin/master [ahead 5, behind 8]
//   D  deleted.txt
//    D missing.txt
//   M  patches.go
//    M changes.go
//   A  newfile.sh
//   ?? isadded.json
func repoStatusGitParse(lines [][]byte) (RepoStatus, error) {
	var status RepoStatus

	for _, line := range lines {
		if len(line) < 4 {
			continue
		}

		if bytes.Equal(line[0:2], []byte{'#', '#'}) {
			repoStatusGitBranch(&status, line)
			continue
		}

		if line[0] == 'D' || line[1] == 'D' {
			status.Deleted++
			continue
		}

		if line[0] == 'M' || line[1] == 'M' {
			status.Modified++
			continue
		}

		if line[0] == 'A' || line[1] == '?' {
			status.Added++
			continue
		}
	}

	return status, nil
}

// repoStatusGitBranch parses the header of the `git status` command.
//
//   ## master...origin/master
//   ## master...origin/master [ahead 5]
//   ## master...origin/master [behind 8]
//   ## master...origin/master [ahead 5, behind 8]
func repoStatusGitBranch(status *RepoStatus, line []byte) {
	var bols [][]byte
	var clean []byte

	if bytes.Contains(line, []byte("...")) {
		status.Branch = line[3:bytes.Index(line, []byte("..."))]
	}

	if bytes.Contains(line, []byte{'['}) && bytes.Contains(line, []byte{']'}) {
		line = line[bytes.Index(line, []byte("["))+1 : len(line)-1]
		line = bytes.Replace(line, []byte("\x20"), []byte{}, -1)
		bols = bytes.Split(line, []byte{','})

		for _, part := range bols {
			if len(part) < 6 {
				continue
			}

			if bytes.Equal(part[0:5], []byte("ahead")) {
				clean = bytes.Replace(part, []byte("ahead"), []byte{}, 1)
				if number, err := strconv.Atoi(string(clean)); err == nil {
					status.Ahead = number
				}
			}

			if bytes.Equal(part[0:5], []byte("behin")) {
				clean = bytes.Replace(part, []byte("behind"), []byte{}, 1)
				if number, err := strconv.Atoi(string(clean)); err == nil {
					status.Behind = number
				}
			}
		}
	}
}

// repoStatusMercurial returns information about the current state of a Mercurial repository.
func repoStatusMercurial() (RepoStatus, error) {
	out, err := call("hg", "status")

	if err != nil {
		return RepoStatus{}, err
	}

	return repoStatusMercurialParse(bytes.Split(out, []byte("\n")))
}

// repoStatusMercurialParse parses the output of the `hg status` command.
//
//   A newfile.sh
//   ? isadded.json
//   M patches.go
//   M changes.go
//   R deleted.txt
//   ! missing.txt
func repoStatusMercurialParse(lines [][]byte) (RepoStatus, error) {
	var status RepoStatus

	if branch, err := ioutil.ReadFile(".hg/branch"); err == nil {
		status.Branch = bytes.TrimSpace(branch)
	} else {
		status.Branch = []byte("default")
	}

	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		if line[0] == 'A' || line[0] == '?' {
			status.Added++
			continue
		}

		if line[0] == 'M' || line[0] == 'm' {
			status.Modified++
			continue
		}

		if line[0] == 'R' || line[0] == '!' {
			status.Deleted++
			continue
		}
	}

	return status, nil
}
