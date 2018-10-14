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

// Powergoline holds the configuration either defined by the current user in
// the TTY session or the default settings defined by the program on startup.
// It also holds the bytes that will be printed in the command line prompt in
// the form of segments.
type Powergoline struct {
	pieces []Segment
	config Config
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

// NewPowergoline loads the config file and instantiates Powergoline.
func NewPowergoline(config Config) *Powergoline {
	return &Powergoline{config: config}
}

// AddSegment inserts a new block in the CLI prompt output.
func (p *Powergoline) AddSegment(text string, fore string, back string) {
	p.pieces = append(p.pieces, Segment{
		Text: text,
		Fore: fore,
		Back: back,
	})
}

// Render sends all the segments to the standard output.
func (p Powergoline) Render(status string) int {
	p.TermTitle()
	p.DateTime()
	p.Username()
	p.Hostname()
	p.Directories()
	p.RepoStatus()
	p.CallPlugins()
	p.RootSymbol(status)

	p.PrintSegments()

	return 0
}

// Print sends a segment to the standard output.
func (p Powergoline) Print(text string, fore string, back string) {
	var color string

	// Add the foreground and background colors.
	if fore != "" && back != "" {
		color += "38;5;" + fore + ";" + "48;5;" + back
	} else if fore != "" {
		color += "38;5;" + fore
	} else if back != "" {
		color += "48;5;" + back
	}

	// Draw the color sequences if necessary.
	if len(color) > 0 {
		fmt.Print("\\[\\e[" + color + "m\\]" + text + "\\[\\e[0m\\]")
		return
	}

	fmt.Print(text)
}

// PrintSegments prints all the segments as the command prompt.
func (p Powergoline) PrintSegments() {
	var curr Segment
	var next Segment

	ttlsegms := len(p.pieces)

	for key := 0; key < ttlsegms; key++ {
		curr = p.pieces[key]

		if curr.Back == "automatic" {
			next = p.pieces[key+1]
			curr.Back = next.Back
		}

		// prevent arbitrary code execution in subshell expressions.
		curr.Text = strings.Replace(curr.Text, "$", "\\$", -1)
		curr.Text = strings.Replace(curr.Text, "`", "\\`", -1)

		p.Print(curr.Text, curr.Fore, curr.Back)
	}

	fmt.Print("\u0020\n")
}

// IsWritable checks if the process can write in a directory.
func (p Powergoline) IsWritable(folder string) bool {
	return unix.Access(folder, unix.W_OK) == nil
}

// IsRdonlyDir checks if a directory is read only by the current user.
func (p Powergoline) IsRdonlyDir(folder string) bool {
	return !p.IsWritable(folder)
}

// TermTitle defines the template for the terminal title.
func (p *Powergoline) TermTitle() {
	p.AddSegment("\\[\\e]0;\\u@\\h: \\w\\a\\]", "", "")
}

// DateTime defines a segment with the current date and time.
func (p *Powergoline) DateTime() {
	if !p.config.Datetime.On {
		return
	}

	p.AddSegment(
		"\x20"+time.Now().Format("15:04:05")+"\x20",
		p.config.Datetime.Fg,
		p.config.Datetime.Bg,
	)
	p.AddSegment(
		"\uE0B0",
		p.config.Datetime.Bg,
		p.config.Username.Bg,
	)
}

// Username defines a segment with the name of the current account.
func (p *Powergoline) Username() {
	if !p.config.Username.On {
		return
	}

	p.AddSegment(
		"\x20\\u\x20",
		p.config.Username.Fg,
		p.config.Username.Bg,
	)
	p.AddSegment(
		"\uE0B0",
		p.config.Username.Bg,
		"automatic",
	)
}

// Hostname defines a segment with the name of this system.
func (p *Powergoline) Hostname() {
	if !p.config.Hostname.On {
		return
	}

	p.AddSegment(
		"\x20\\h\x20",
		p.config.Hostname.Fg,
		p.config.Hostname.Bg,
	)
	p.AddSegment(
		"\uE0B0",
		p.config.Hostname.Bg,
		"automatic",
	)
}

// HomeDirectory defines a segment with current directory path.
func (p *Powergoline) HomeDirectory() {
	p.AddSegment(
		"\x20~\x20",
		p.config.HomeDir.Fg,
		p.config.HomeDir.Bg,
	)
	p.AddSegment(
		"\uE0B0",
		p.config.HomeDir.Bg,
		"automatic",
	)
}

// Directories returns the full path of the current directory.
func (p *Powergoline) Directories() {
	homedir := os.Getenv("HOME")
	currdir := os.Getenv("PWD")
	shortdir := strings.Replace(currdir, homedir, "", 1)
	cleandir := strings.Trim(shortdir, "/")

	// Draw the sequence of folders of the current path.
	maxsegms := p.config.CurrentDir.Size
	dirparts := strings.Split(cleandir, "/")
	ttlparts := len(dirparts)
	lastsegm := (ttlparts - 1)

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
	if strings.Index(currdir, homedir) == 0 {
		p.HomeDirectory()
	}

	// Draw each directory segment with right arrow.
	for key, folder := range dirparts {
		if folder == "" {
			continue
		}

		p.AddSegment(
			"\x20"+folder+"\x20",
			p.config.CurrentDir.Fg,
			p.config.CurrentDir.Bg,
		)

		if key == lastsegm {
			p.AddSegment(
				"\uE0B0",
				p.config.CurrentDir.Bg,
				"automatic",
			)
		} else {
			p.AddSegment(
				"\uE0B1",
				p.config.CurrentDir.Fg,
				p.config.CurrentDir.Bg,
			)
		}
	}

	// Draw lock if current directory is read-only.
	if p.IsRdonlyDir(currdir) {
		p.AddSegment(
			"\x20\uE0A2\x20",
			p.config.RdonlyDir.Fg,
			p.config.RdonlyDir.Bg,
		)

		p.AddSegment(
			"\uE0B0",
			p.config.RdonlyDir.Bg,
			"automatic",
		)
	}
}

// RepoStatus defines a segment with information of a DCVS.
func (p *Powergoline) RepoStatus() {
	if !p.config.Repository.On {
		return
	}

	var err error
	var status RepoStatus

	// check if a repository exists in the current folder.
	if _, err = os.Stat(".git"); !os.IsNotExist(err) {
		status, err = repoStatusGit()
	} else if _, err = os.Stat(".hg"); !os.IsNotExist(err) {
		status, err = repoStatusMercurial()
	}

	if err != nil {
		fmt.Println(err)
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

	p.AddSegment(
		branch,
		p.config.Repository.Fg,
		p.config.Repository.Bg,
	)
	p.AddSegment(
		"\uE0B0",
		p.config.Repository.Bg,
		"automatic",
	)
}

// CallPlugins runs all the user defined plugins.
func (p *Powergoline) CallPlugins() {
	for _, metadata := range p.config.Plugins {
		p.ExecutePlugin(metadata)
	}
}

// ExecutePlugin runs an user defined external command.
func (p *Powergoline) ExecutePlugin(addon Plugin) {
	output, err := call(addon.Command)

	if err == errEmptyOutput {
		/* no output */
		return
	}

	if err != nil {
		/* use error message instead */
		output = []byte(err.Error())
	}

	p.AddSegment("\x20"+string(output)+"\x20", addon.Fg, addon.Bg)
	p.AddSegment("\uE0B0", addon.Bg, "automatic")
}

// RootSymbol defines a segment with an indicator for root users.
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
func (p *Powergoline) RootSymbol(status string) {
	var symbol string
	var color string

	if os.Getuid() == 0 {
		symbol = p.config.Symbol.SuperUser
	} else {
		symbol = p.config.Symbol.Regular
	}

	switch status {
	case "0":
		color = p.config.Status.Success
	case "1":
		color = p.config.Status.Failure
	case "126":
		color = p.config.Status.Permission
	case "127":
		color = p.config.Status.NotFound
	case "128":
		color = p.config.Status.InvalidExit
	case "130":
		color = p.config.Status.Terminated
	default:
		color = p.config.Status.Misuse
	}

	p.AddSegment("\x20"+symbol+"\x20", p.config.Status.Symbol, color)
	p.AddSegment("\uE0B0", color, "")
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
