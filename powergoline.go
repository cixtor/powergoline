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
var auto int = 0x9999

// u000a is Unicode for `\n` (new line).
const u000a = "\u000a"

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
	Print   bool
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
	p.TermTitle()
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
func (p Powergoline) Print(w io.Writer, s string, fg int, bg int) (int, error) {
	var color string

	fore := colorize(fg)
	back := colorize(bg)

	// Add the foreground and background colors.
	if fg > -1 && bg > -1 {
		color += "38;5;" + fore + ";" + "48;5;" + back
	} else if fg > -1 {
		color += "38;5;" + fore
	} else if bg > -1 {
		color += "48;5;" + back
	}

	// Draw the color sequences if necessary.
	if len(color) > 0 {
		return fmt.Fprint(w, "\\[\\e["+color+"m\\]"+s+"\\[\\e[0m\\]")
	}

	return fmt.Fprint(w, s)
}

// PrintSegments prints all the segments as the command prompt.
func (p Powergoline) PrintSegments(w io.Writer) (int, error) {
	var curr Segment

	ttlsegms := len(p.pieces)

	for key := 0; key < ttlsegms; key++ {
		curr = p.pieces[key]

		// Use the next segment's background color for the current segment.
		if curr.Bg == auto && len(p.pieces) > key+1 {
			curr.Bg = p.pieces[key+1].Bg
		}

		// prevent arbitrary code execution in subshell expressions.
		curr.S = strings.Replace(curr.S, "$", "\\$", -1)
		curr.S = strings.Replace(curr.S, "`", "\\`", -1)

		if _, err := p.Print(w, curr.S, curr.Fg, curr.Bg); err != nil {
			return 0, err
		}
	}

	return fmt.Fprint(w, u0020+u000a)
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
	p.AddSegment("\\[\\e]0;\\u@\\h: \\w\\a\\]", -1, -1)
}

// Datetime defines a segment with the current date and time.
func (p *Powergoline) Datetime() {
	if !p.config.TimeOn {
		return
	}

	p.AddSegment(u0020+time.Now().Format(p.config.TimeFmt)+u0020, p.config.TimeFg, p.config.TimeBg)
	p.AddSegment(uE0B0, p.config.TimeBg, auto)
}

// Username defines a segment with the name of the current account.
func (p *Powergoline) Username() {
	if !p.config.UserOn {
		return
	}

	p.AddSegment(u0020+"\\u"+u0020, p.config.UserFg, p.config.UserBg)
	p.AddSegment(uE0B0, p.config.UserBg, auto)
}

// Hostname defines a segment with the name of this system.
func (p *Powergoline) Hostname() {
	if !p.config.HostOn {
		return
	}

	p.AddSegment(u0020+"\\h"+u0020, p.config.HostFg, p.config.HostBg)
	p.AddSegment(uE0B0, p.config.HostBg, auto)
}

// HomeDirectory defines a segment with current directory path.
func (p *Powergoline) HomeDirectory() {
	p.AddSegment(u0020+"~"+u0020, p.config.HomeFg, p.config.HomeBg)
	p.AddSegment(uE0B0, p.config.HomeBg, auto)
}

// Directories returns the full path of the current directory.
func (p *Powergoline) Directories() {
	if !p.config.CwdOn {
		return
	}

	homedir := os.Getenv("HOME")
	currdir := os.Getenv("PWD")
	shortdir := strings.Replace(currdir, homedir, "", 1)
	cleandir := strings.Trim(shortdir, "/")

	// Draw the sequence of folders of the current path.
	maxsegms := p.config.CwdN
	dirparts := strings.Split(cleandir, "/")
	ttlparts := len(dirparts)
	lastsegm := (ttlparts - 1)

	if maxsegms < 1 {
		maxsegms = 1
	}

	if ttlparts > maxsegms {
		newparts := make([]string, 0)
		offset := (maxsegms - 1)
		newparts = append(newparts, u2026)
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

		p.AddSegment(u0020+folder+u0020, p.config.CwdFg, p.config.CwdBg)

		if key == lastsegm {
			p.AddSegment(uE0B0, p.config.CwdBg, auto)
		} else {
			p.AddSegment(uE0B1, p.config.CwdFg, p.config.CwdBg)
		}
	}

	// Draw lock if current directory is read-only.
	if p.IsRdonlyDir(currdir) {
		p.AddSegment(u0020+uE0A2+u0020, p.config.RodirFg, p.config.RodirBg)
		p.AddSegment(uE0B0, p.config.RodirBg, auto)
	}
}

// RepoStatusExclude checks if the current folder excludes repository status.
func (p *Powergoline) RepoStatusExclude() bool {
	if !p.config.RepoOn {
		return true
	}

	currdir := os.Getenv("PWD")

	for _, folder := range p.config.RepoExclude {
		if currdir == folder {
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
	p.AddSegment(uE0B0, p.config.RepoBg, auto)
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

		if !bucket[i].Print {
			continue
		}

		p.AddSegment(u0020+bucket[i].S+u0020, bucket[i].Fg, bucket[i].Bg)
		p.AddSegment(uE0B0, auto, auto)
	}
}

// ExecutePlugin runs an user defined external command.
func (p *Powergoline) ExecutePlugin(sem chan bool, out chan Segment, index int, addon Plugin) {
	sem <- true /* block */
	defer func() { <-sem }()

	start := time.Now()
	output, err := call(addon.Name, addon.Args...)
	runtime := time.Since(start)

	if err == errEmptyOutput {
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
		Print:   true,
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
	p.AddSegment(uE0B0, color, -1)
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
