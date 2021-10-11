package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// auto is an invalid xterm color code used to represent automatic selection of
// the color, either foreground or background, of the previous segment in the
// list to colorize the next segment.
var auto int = 0x29a

// u000A is Unicode for `\n` (new line).
const u000A = "\u000A"

// u0020 is Unicode for `\s` (whitespace).
const u0020 = "\u0020"

// u2026 is Unicode for `…` (ellipsis).
const u2026 = "\u2026"

// u21E1 is Unicode for `⇡` (upwards dashed arrow).
const u21E1 = "\u21E1"

// u21E3 is Unicode for `⇣` (downwards dashed arrow).
const u21E3 = "\u21E3"

// uE0A0 is Unicode for `` (GitHub fork symbol).
const uE0A0 = "\uE0A0"

// uE0A2 is Unicode for `` (GitHub lock symbol).
const uE0A2 = "\uE0A2"

// uE0B0 is Unicode for `` (powerline arrow body).
const uE0B0 = "\uE0B0"

// uE0B1 is Unicode for `` (powerline arrow line).
const uE0B1 = "\uE0B1"

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
	S       string
	Fg      int
	Bg      int
	Index   int
	Runtime time.Duration
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

func colorize(n int) string {
	return fmt.Sprintf("%03d", n)
}

// AddSegment inserts a new block in the CLI prompt output.
func (p *Powergoline) AddSegment(s string, fg int, bg int) {
	p.pieces = append(p.pieces, Segment{S: s, Fg: fg, Bg: bg})
}

// Render sends all the segments to the standard output.
func (p Powergoline) Render(w io.Writer) int {
	p.Datetime()
	p.Username()
	p.Hostname()
	p.Directories()
	p.RepoStatus()
	p.CallPlugins()
	p.RootSymbol()

	if _, err := p.PrintSegments(w); err != nil {
		fmt.Println(err)
		return 1
	}

	return 0
}

// Print sends a segment to the standard output.
func (p Powergoline) Print(w io.Writer, seg Segment) (int, error) {
	var color string

	fore := colorize(seg.Fg)
	back := colorize(seg.Bg)

	// Add the foreground and background colors.
	if seg.Fg > -1 && seg.Bg > -1 {
		color += "38;5;" + fore + ";" + "48;5;" + back
	} else if seg.Fg > -1 {
		color += "38;5;" + fore
	} else if seg.Bg > -1 {
		color += "48;5;" + back
	}

	// Draw the color sequences if necessary.
	if len(color) > 0 {
		return fmt.Fprint(w, "\\[\\e["+color+"m\\]"+seg.S+"\\[\\e[0m\\]")
	}

	return fmt.Fprint(w, seg.S)
}

// PrintSegments prints all the segments as the command prompt.
func (p Powergoline) PrintSegments(w io.Writer) (int, error) {
	pieces := p.pieces
	total := len(pieces)

	for i := 0; i < total; i++ {
		section := pieces[i]

		if section.S == "" {
			continue
		}

		// Prevent arbitrary code execution in subshell expressions.
		section.S = strings.Replace(section.S, "$", "\\$", -1)
		section.S = strings.Replace(section.S, "`", "\\`", -1)

		if _, err := p.Print(w, section); err != nil {
			return 0, err
		}

		// If the segment is an arrow, then we will assume that both Fg and Bg
		// are "auto", which means we must select the corresponding colors from
		// the adjacent segments.
		//
		// The foreground color must be background of the left segment.
		//
		// The background color must be background of the right segment.
		//
		// ┌───┬───┬─────┬───┬─────────────┬───┬─────────────┬───┬───┬───┐
		// │ ~ │ > │ ... │ > │ powergoline │ > │ hello world │ > │ $ │ > │
		// └───┴───┴─────┴───┴─────────────┴───┴─────────────┴───┴───┴───┘
		//       ▲         ▲                 ▲                 ▲       ▲
		//       │         │                 │                 │       │
		//     arrow     arrow             arrow             arrow   arrow
		arrow := Segment{S: uE0B0, Fg: section.Bg, Bg: -1}
		if i+1 < total {
			arrow.Bg = pieces[i+1].Bg
		}
		if _, err := p.Print(w, arrow); err != nil {
			return 0, err
		}
	}

	return fmt.Fprint(w, u0020+u000A)
}

// IsWritable checks if the process can write in a directory.
func (p Powergoline) IsWritable(folder string) bool {
	return unix.Access(folder, unix.W_OK) == nil
}

// IsRdonlyDir checks if a directory is read only by the current user.
func (p Powergoline) IsRdonlyDir(folder string) bool {
	return !p.IsWritable(folder)
}

// Datetime defines a segment with the current date and time.
func (p *Powergoline) Datetime() {
	if !p.config.TimeOn {
		return
	}

	p.AddSegment(u0020+time.Now().Format(p.config.TimeFmt)+u0020, p.config.TimeFg, p.config.TimeBg)
}

// Username defines a segment with the name of the current account.
func (p *Powergoline) Username() {
	if !p.config.UserOn {
		return
	}

	p.AddSegment(u0020+"\\u"+u0020, p.config.UserFg, p.config.UserBg)
}

// Hostname defines a segment with the name of this system.
func (p *Powergoline) Hostname() {
	if !p.config.HostOn {
		return
	}

	p.AddSegment(u0020+"\\h"+u0020, p.config.HostFg, p.config.HostBg)
}

var sep string = "/"

// Directories returns the full path of the current directory.
func (p *Powergoline) Directories() {
	if !p.config.CwdOn {
		return
	}

	homedir := os.Getenv("HOME")
	workdir := os.Getenv("PWD")

	p.DirectoriesHome(homedir, workdir)
	p.DirectoriesOthers(homedir, workdir)
	p.DirectoriesReadOnly(homedir, workdir)
}

func (p *Powergoline) DirectoriesHome(homedir string, workdir string) {
	if strings.HasPrefix(workdir, homedir) {
		// Add a tilde to represent that we are inside the home directory.
		p.AddSegment(u0020+"~"+u0020, p.config.HomeFg, p.config.HomeBg)
	}
}

func (p *Powergoline) DirectoriesOthers(homedir string, workdir string) {
	// Since we already considered the home directory in DirectoriesHome, we do
	// not need to process that portion of the current working directory again.
	// We can safely remove it and continue with the other folders.
	workdir = strings.TrimPrefix(workdir, homedir)

	if workdir == "" {
		return
	}

	if workdir == sep {
		p.AddSegment(u0020+sep+u0020, p.config.CwdFg, p.config.CwdBg)
		return
	}

	folders := strings.Split(workdir, sep)
	ttldirs := len(folders)

	if ttldirs > p.config.CwdN {
		// Replace parent folders with an ellipsis if the path is too long.
		folders = append([]string{"", u2026}, folders[ttldirs-p.config.CwdN:]...)
	}

	// Combine adding a powerline arrow line in between folders.
	// We start at index one because the first folder is empty.
	workdir = strings.Join(folders[1:], u0020+uE0B1+u0020)

	p.AddSegment(u0020+workdir+u0020, p.config.CwdFg, p.config.CwdBg)
}

func (p *Powergoline) DirectoriesReadOnly(homedir string, workdir string) {
	if p.IsRdonlyDir(workdir) {
		// Draw lock symbol if the current directory is read-only.
		p.AddSegment(u0020+uE0A2+u0020, p.config.RodirFg, p.config.RodirBg)
	}
}

// RepoStatusExclude checks if the current folder excludes repository status.
func (p *Powergoline) RepoStatusExclude() bool {
	if !p.config.RepoOn {
		return true
	}

	workdir := os.Getenv("PWD")

	for _, folder := range p.config.RepoExclude {
		if workdir == folder {
			return true
		}
	}

	return false
}

// RepoStatus defines a segment with information of a DCVS.
func (p *Powergoline) RepoStatus() {
	if p.RepoStatusExclude() {
		return
	}

	var err error
	var stderr error
	var status RepoStatus

	// check if a repository exists in the current folder.
	if _, err = os.Stat(".git"); !os.IsNotExist(err) {
		status, stderr = repoStatusGit()
	} else if _, err = os.Stat(".hg"); !os.IsNotExist(err) {
		status, stderr = repoStatusMercurial()
	}

	if stderr != nil {
		fmt.Printf(program+"; repo %s\n", err)
		return
	}

	// there is no info to show.
	if len(status.Branch) == 0 {
		return
	}

	branch := fmt.Sprintf(" %s %s ", uE0A0, status.Branch)

	if status.Ahead > 0 {
		branch += fmt.Sprintf("%s%d ", u21E1, status.Ahead)
	}

	if status.Behind > 0 {
		branch += fmt.Sprintf("%s%d ", u21E3, status.Behind)
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

	p.AddSegment(branch, p.config.RepoFg, p.config.RepoBg)
}

// CallPlugins runs all the user defined plugins.
func (p *Powergoline) CallPlugins() {
	sem := make(chan bool, 20)
	total := len(p.config.Plugins)
	output := make(chan Segment)
	bucket := make([]Segment, total)

	for index, metadata := range p.config.Plugins {
		go p.ExecutePlugin(sem, output, index, metadata)
	}

	var seg Segment
	for i := 0; i < total; i++ {
		seg = <-output
		bucket[seg.Index] = seg
	}

	for i := 0; i < total; i++ {
		if p.config.Debug {
			fmt.Printf(
				"%s took %s\n",
				p.config.Plugins[i].Name,
				bucket[i].Runtime,
			)
		}

		if bucket[i].S == "" {
			continue
		}

		p.AddSegment(u0020+bucket[i].S+u0020, bucket[i].Fg, bucket[i].Bg)
	}
}

// ExecutePlugin runs an user defined external command.
func (p *Powergoline) ExecutePlugin(sem chan bool, out chan Segment, index int, addon Plugin) {
	sem <- true /* block */
	defer func() { <-sem }()

	start := time.Now()
	output, err := call(addon.Name, addon.Args...)
	runtime := time.Since(start)

	if errors.Is(err, errEmptyOutput) {
		out <- Segment{Index: index, Runtime: runtime}
		return
	}

	if err != nil {
		/* use error message instead */
		output = []byte(err.Error())
	}

	out <- Segment{
		S:       string(output),
		Fg:      auto,
		Bg:      auto,
		Index:   index,
		Runtime: runtime,
	}
}

// RootSymbol defines a segment with an indicator for root users.
//
// System status codes:
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
func (p *Powergoline) RootSymbol() {
	var color int
	var symbol string
	status := p.config.StatusCode

	if os.Getuid() == 0 {
		symbol = p.config.SymbolRoot
	} else {
		symbol = p.config.SymbolUser
	}

	if status == 0 {
		color = p.config.StatusSuccess
	} else if status == 1 {
		color = p.config.StatusError
	} else if status == 2 {
		color = p.config.StatusMisuse
	} else if status == 126 {
		color = p.config.StatusCantExec
	} else if status == 127 {
		color = p.config.StatusNotFound
	} else if status == 128 {
		color = p.config.StatusInvalid
	} else if status > 128 && status != 130 && status < 255 {
		color = p.config.StatusErrSignal
	} else if status == 130 {
		color = p.config.StatusTerminated
	} else {
		color = p.config.StatusOutofrange
	}

	p.AddSegment(u0020+symbol+u0020, p.config.StatusFg, color)
}

// runcmd executes an external command and returns the output.
func call(name string, arg ...string) ([]byte, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command(name, arg...)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() == 0 {
			return nil, err
		}

		// include additional information, if possible.
		return nil, fmt.Errorf("%s", stderr.String())
	}

	if stdout.Len() == 0 {
		return nil, errEmptyOutput
	}

	return bytes.Trim(stdout.Bytes(), "\n"), nil
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
//   ## master
//   ## master...origin/master
//   ## master...origin/master [ahead 5]
//   ## master...origin/master [behind 8]
//   ## master...origin/master [ahead 5, behind 8]
func repoStatusGitBranch(status *RepoStatus, line []byte) {
	var bols [][]byte
	var clean []byte

	// add ellipsis to parse branch without origin.
	line = append(line, []byte{'.', '.', '.'}...)

	if bytes.Contains(line, []byte("...")) {
		status.Branch = line[3:bytes.Index(line, []byte("..."))]
	}

	// detect limits for the ahead/behind status.
	opening := bytes.Index(line, []byte{'['}) + 1
	closing := bytes.Index(line, []byte{']'}) + 0

	if opening == -1 || closing == -1 {
		return
	}

	line = line[opening:closing]
	line = bytes.Replace(line, []byte(u0020), []byte{}, -1)
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
