// Powergoline
//
// A lightweight status line for your terminal emulator. This project aims to
// be a lightweight alternative for powerline a popular statusline plugin for
// VIm that statuslines and prompts for several other applications, including
// zsh, bash, tmux, IPython, Awesome and Qtile.
//
//   - https://cixtor.com/
//   - https://github.com/cixtor/powergoline
//   - https://en.wikipedia.org/wiki/Status_bar
//
// A status bar is a graphical control element which poses an information area
// typically found at the window's bottom. It can be divided into sections to
// group information. Its job is primarily to display information about the
// current state of its window, although some status bars have extra
// functionality.
//
// A status bar can also be text-based, primarily in console-based applications,
// in which case it is usually the last row in an 80x25 text mode configuration,
// leaving the top 24 rows for application data. Usually the status bar (called
// a status line in this context) displays the current state of the application,
// as well as helpful keyboard shortcuts.
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

// config is the user-provided configuration.
var config Config

const defaultPluginTimeout time.Duration = time.Second * 3

var themes = map[string]func(Config) Config{
	"agnoster":   ApplyAgnosterTheme,
	"astrocom":   ApplyAstrocomTheme,
	"bluescale":  ApplyBlueScaleTheme,
	"colorish":   ApplyColorishTheme,
	"grayscale":  ApplyGrayScaleTheme,
	"wildcherry": ApplyWildCherryTheme,
}

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
	flag.IntVar(&config.RepoFg, "repo.fg", 0, "Defines the repository status foreground color")
	flag.IntVar(&config.RepoBg, "repo.bg", 255, "Defines the repository status background color")
	flag.Var(&config.RepoExclude, "repo.exclude", "Sets repo.on=false for the specified folder")
	flag.Var(&config.RepoInclude, "repo.include", "Sets repo.on=true for the specified folder")
	flag.Var(&config.Plugins, "plugin", "Defines a plugin with optional arguments (e.g. -plugin=\"echo hello world\")\nDefine multiple plugins like this: -plugin=A -plugin=B -plugin=C")
	flag.IntVar(&config.PluginFg, "plugin.fg", 0, "Defines the plugin output foreground color")
	flag.IntVar(&config.PluginBg, "plugin.bg", 11, "Defines the plugin output background color")
	flag.DurationVar(&config.PluginTimeout, "plugin.timeout", time.Second*5, "Maximum time to wait for a plugin execution")
	flag.StringVar(&config.SymbolRoot, "symbol.root", "#", "Defines the prompt symbol for the Root user session")
	flag.StringVar(&config.SymbolUser, "symbol.user", "$", "Defines the prompt symbol for a Regular user session")
	flag.IntVar(&config.StatusFg, "status.fg", 255, "Defines the program exit status foreground color")
	flag.IntVar(&config.StatusCode, "status.code", -1, "Exit status code of the most recent program execution")
	flag.IntVar(&config.StatusSuccess, "status.success", 41, "Defines the background color for exit(0)\nOperation success and generic status code.")
	flag.IntVar(&config.StatusError, "status.error", 1, "Defines the background color for exit(1)\nCatchall for general errors and failures.")
	flag.IntVar(&config.StatusMisuse, "status.misuse", 3, "Defines the background color for exit(2)\nMisuse of shell builtins, missing command or permission problem.")
	flag.IntVar(&config.StatusCantExec, "status.cantexec", 4, "Defines the background color for exit(126)\nCannot execute command, permission problem, or not an executable.")
	flag.IntVar(&config.StatusNotFound, "status.notfound", 14, "Defines the background color for exit(127)\nCommand not found, illegal path, or possible typo.")
	flag.IntVar(&config.StatusInvalid, "status.invalid", 202, "Defines the background color for exit(128)\nInvalid argument to exit, only use range 0-255.")
	flag.IntVar(&config.StatusErrSignal, "status.errsignal", 8, "Defines the background color for exit(128+n)\nFatal error signal where \"n\" is the PID.")
	flag.IntVar(&config.StatusTerminated, "status.terminated", 13, "Defines the background color for exit(130)\nScript terminated by Control-C.")
	flag.IntVar(&config.StatusOutofrange, "status.outofrange", 0, "Defines the background color for exit(255*)\nExit status out of range.")
	flag.StringVar(&config.Theme, "theme", "", "Automatic configuration based on a predefined color scheme")

	flag.Parse()

	NewPowergoline(config).Render(os.Stdout, []SegmentFunc{
		segmentDatetime,
		segmentUsername,
		segmentHostname,
		segmentDirectories,
		segmentRepoStatus,
		segmentCallPlugins,
		segmentPromptSymbol,
	})
}

const (
	u000A string = "\u000A" // u000A is Unicode for `\n` (new line).
	u0020 string = "\u0020" // u0020 is Unicode for `\s` (whitespace).
	u2026 string = "\u2026" // u2026 is Unicode for `…` (ellipsis).
	u21E1 string = "\u21E1" // u21E1 is Unicode for `⇡` (upwards dashed arrow).
	u21E3 string = "\u21E3" // u21E3 is Unicode for `⇣` (downwards dashed arrow).
	uE0A0 string = "\uE0A0" // uE0A0 is Unicode for `` (GitHub fork symbol).
	uE0A2 string = "\uE0A2" // uE0A2 is Unicode for `` (GitHub lock symbol).
	uE0B0 string = "\uE0B0" // uE0B0 is Unicode for `` (powerline arrow body).
	uE0B1 string = "\uE0B1" // uE0B1 is Unicode for `` (powerline arrow line).
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
	Index uint   // order in which to render.
	Show  bool   // render if true, hide if false.
	Fg    int    // foreground color.
	Bg    int    // background color.
	Text  string // text to render.
}

// PluginOutput struct represents the output of an external program after its
// execution along with some runtime information and an index. The index field
// is used to keep track of the order in which the programs were executed; for
// example, the first program to execute will have an index of 0, the second
// will have an index of 1, and so on.
//
// This struct is typically used in conjunction with a slice of PluginOutput
// structs, where each struct in the slice represents the output of a single
// program execution.
type PluginOutput struct {
	Index   int
	Output  string
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
	if config.Theme != "" {
		if applyThemeConfig, ok := themes[config.Theme]; ok {
			config = applyThemeConfig(config)
		}
	}
	return &Powergoline{config: config}
}

type SegmentFunc func(*sync.WaitGroup, chan struct{}, chan Segment, uint, Config)

func (p *Powergoline) Render(w io.Writer, arr []SegmentFunc) {
	var wg sync.WaitGroup
	out := make(chan Segment)
	sem := make(chan struct{}, 10)
	done := make(chan struct{})
	go consumer(w, done, out)
	for priority, fn := range arr {
		wg.Add(1)
		sem <- struct{}{ /* lock */ }
		go fn(&wg, sem, out, uint(priority), p.config)
	}
	wg.Wait()
	close(sem)
	close(out)
	<-done
}

func consumer(w io.Writer, done chan struct{}, out chan Segment) {
	defer close(done)
	var segments []Segment
	for item := range out {
		segments = append(segments, item)
	}
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].Index < segments[j].Index
	})
	_, _ = printAllSegments(w, segments)
}

func printAllSegments(w io.Writer, segments []Segment) (int, error) {
	n := len(segments)
	for i, section := range segments {
		if section.Text == "" {
			continue
		}
		// Prevent arbitrary code execution in subshell expressions.
		section.Text = strings.ReplaceAll(section.Text, "$", "\\$")
		section.Text = strings.ReplaceAll(section.Text, "`", "\\`")
		if _, err := printOneSegment(w, section); err != nil {
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
		arrow := Segment{Text: uE0B0, Fg: section.Bg, Bg: -1}
		if i+1 < n {
			arrow.Bg = segments[i+1].Bg
		}
		if _, err := printOneSegment(w, arrow); err != nil {
			return 0, err
		}
	}
	return fmt.Fprint(w, u0020+u000A)
}

func printOneSegment(w io.Writer, seg Segment) (int, error) {
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
		return fmt.Fprint(w, "\\[\\e["+color+"m\\]"+seg.Text+"\\[\\e[0m\\]")
	}
	return fmt.Fprint(w, seg.Text)
}

// segmentDatetime prints the current date and time.
func segmentDatetime(wg *sync.WaitGroup, sem chan struct{}, out chan Segment, priority uint, config Config) {
	defer wg.Done()
	defer func() { <-sem }()
	if !config.TimeOn {
		out <- Segment{ /* disabled */ }
		return
	}
	out <- Segment{Index: priority, Show: true, Fg: config.TimeFg, Bg: config.TimeBg, Text: u0020 + time.Now().Format(config.TimeFmt) + u0020}
}

func segmentUsername(wg *sync.WaitGroup, sem chan struct{}, out chan Segment, priority uint, config Config) {
	defer wg.Done()
	defer func() { <-sem }()
	// if !config.UserOn { return }
	out <- Segment{Index: priority, Text: "segmentUsername", Show: true}
}

func segmentHostname(wg *sync.WaitGroup, sem chan struct{}, out chan Segment, priority uint, config Config) {
	defer wg.Done()
	defer func() { <-sem }()
	// if !config.HostOn { return }
	out <- Segment{Index: priority, Text: "segmentHostname", Show: true}
}

func segmentDirectories(wg *sync.WaitGroup, sem chan struct{}, out chan Segment, priority uint, config Config) {
	defer wg.Done()
	defer func() { <-sem }()
	// if !config.CwdOn { return }
	out <- Segment{Index: priority, Text: "segmentDirectories", Show: true}
}

func segmentRepoStatus(wg *sync.WaitGroup, sem chan struct{}, out chan Segment, priority uint, config Config) {
	defer wg.Done()
	defer func() { <-sem }()
	// if !config.RepoOn { return }
	out <- Segment{Index: priority, Text: "segmentRepoStatus", Show: true}
}

func segmentCallPlugins(wg *sync.WaitGroup, sem chan struct{}, out chan Segment, priority uint, config Config) {
	defer wg.Done()
	defer func() { <-sem }()
	for i, command := range config.Plugins {
		wg.Add(1)
		sem <- struct{}{ /* lock */ }
		go segmentCallOnePlugin(wg, sem, out, uint(i+100), config, command)
	}
}

func segmentCallOnePlugin(wg *sync.WaitGroup, sem chan struct{}, out chan Segment, priority uint, config Config, command Plugin) {
	defer wg.Done()
	defer func() { <-sem }()
	out <- Segment{Index: priority, Text: "segmentCallOnePlugin " + command.Name, Show: true}
}

func segmentPromptSymbol(wg *sync.WaitGroup, sem chan struct{}, out chan Segment, _ uint, config Config) {
	defer wg.Done()
	defer func() { <-sem }()
	out <- Segment{Index: 999999, Text: "segmentPromptSymbol", Show: true}
}

func colorize(n int) string {
	return fmt.Sprintf("%03d", n)
}

// AddSegment inserts a new block in the CLI prompt output.
func (p *Powergoline) AddSegment(s string, fg int, bg int) {
	p.pieces = append(p.pieces, Segment{Text: s, Fg: fg, Bg: bg})
}

// IsWritable checks if the process can write in a directory.
func (p Powergoline) IsWritable(folder string) bool {
	return unix.Access(folder, unix.W_OK) == nil
}

// IsRdonlyDir checks if a directory is read only by the current user.
func (p Powergoline) IsRdonlyDir(folder string) bool {
	return !p.IsWritable(folder)
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

// IsRepoStatusEnabled checks if the current folder excludes repository status.
func (p *Powergoline) IsRepoStatusEnabled() bool {
	return p.config.RepoOn && !slices.Contains(p.config.RepoExclude, os.Getenv("PWD"))
}

// RepoStatus defines a segment with information of a DCVS.
func (p *Powergoline) RepoStatus() {
	if !p.IsRepoStatusEnabled() {
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
		fmt.Println("hgrepo", err)
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
	output := make(chan PluginOutput)
	bucket := make([]PluginOutput, total)

	for index, metadata := range p.config.Plugins {
		go p.ExecutePlugin(sem, output, index, metadata)
	}

	var pout PluginOutput

	for i := 0; i < total; i++ {
		pout = <-output
		bucket[pout.Index] = pout
	}

	allOutputs := []string{}

	for i := 0; i < total; i++ {
		if p.config.Debug {
			fmt.Printf("%s took %s\n", p.config.Plugins[i].Name, bucket[i].Runtime)
		}

		if bucket[i].Output == "" {
			continue
		}

		allOutputs = append(allOutputs, bucket[i].Output)
	}

	if len(allOutputs) == 0 {
		return
	}

	outSeq := strings.Join(allOutputs, u0020+uE0B1+u0020)
	p.AddSegment(u0020+outSeq+u0020, p.config.PluginFg, p.config.PluginBg)
}

// ExecutePlugin runs an user defined external command.
func (p *Powergoline) ExecutePlugin(sem chan bool, out chan PluginOutput, index int, addon Plugin) {
	sem <- true /* block */
	defer func() { <-sem }()

	start := time.Now()
	output, err := call(p.config.PluginTimeout, addon.Name, addon.Args...)
	runtime := time.Since(start)

	if errors.Is(err, errEmptyOutput) {
		out <- PluginOutput{Index: index, Runtime: runtime}
		return
	}

	if err != nil {
		/* use error message instead */
		output = []byte(err.Error())
	}

	out <- PluginOutput{
		Output:  string(output),
		Index:   index,
		Runtime: runtime,
	}
}

// RootSymbol defines a segment with an indicator for root users.
//
// System status codes:
//
//	> 0     - Operation success and generic status code.
//	> 1     - Catchall for general errors and failures.
//	> 2     - Misuse of shell builtins, missing command or permission problem.
//	> 126   - Cannot execute command, permission problem, or not an executable.
//	> 127   - Command not found, illegal path, or possible typo.
//	> 128   - Invalid argument to exit, only use range 0-255.
//	> 128+n - Fatal error signal where "n" is the PID.
//	> 130   - Script terminated by Control-C.
//	> 255*  - Exit status out of range.
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
func call(timeout time.Duration, name string, arg ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("%s timeout after %s", name, timeout)
		}
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
	out, err := call(defaultPluginTimeout, "git", "status", "--branch", "--porcelain", "--ignore-submodules")

	if err != nil {
		return RepoStatus{}, err
	}

	return repoStatusGitParse(bytes.Split(out, []byte("\n")))
}

// repoStatusGitParse parses the output of the `git status` command.
//
//	> ## master...origin/master [ahead 5, behind 8]
//	> D  deleted.txt
//	>  D missing.txt
//	> M  patches.go
//	>  M changes.go
//	> A  newfile.sh
//	> ?? isadded.json
func repoStatusGitParse(lines [][]byte) (RepoStatus, error) {
	var status RepoStatus

	for _, line := range lines {
		if len(line) < 4 {
			continue
		}

		if line[0] == '#' && line[1] == '#' {
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
//	> ## master
//	> ## master...origin/master
//	> ## master...origin/master [ahead 5]
//	> ## master...origin/master [behind 8]
//	> ## master...origin/master [ahead 5, behind 8]
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
	line = bytes.ReplaceAll(line, []byte(u0020), []byte{})
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
	out, err := call(defaultPluginTimeout, "hg", "status")

	if err != nil {
		return RepoStatus{}, err
	}

	return repoStatusMercurialParse(bytes.Split(out, []byte("\n")))
}

// repoStatusMercurialParse parses the output of the `hg status` command.
//
//	> A newfile.sh
//	> ? isadded.json
//	> M patches.go
//	> M changes.go
//	> R deleted.txt
//	> ! missing.txt
func repoStatusMercurialParse(lines [][]byte) (RepoStatus, error) {
	var status RepoStatus

	if branch, err := os.ReadFile(".hg/branch"); err == nil {
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
