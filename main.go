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
		segmentExitCode,
		segmentInputSeparator,
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

type SegmentKind int

const (
	TextBox SegmentKind = iota
	FolderBox
	ArrowBox
	LockBox
	RepoStatusBox
	PluginBox
	ExitCodeBox
	LastBox
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
	Kind  SegmentKind // type of box that the segment represents.
	Index int         // order in which to render.
	Show  bool        // render if true, hide if false.
	Fg    int         // foreground color.
	Bg    int         // background color.
	Text  string      // text to render.
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

type SegmentFunc func(*sync.WaitGroup, chan struct{}, chan Segment, int, Config)

func (p *Powergoline) Render(w io.Writer, arr []SegmentFunc) {
	var wg sync.WaitGroup
	out := make(chan Segment)
	sem := make(chan struct{}, 10)
	done := make(chan struct{})
	go consumer(w, done, out)
	for priority, fn := range arr {
		wg.Add(1)
		sem <- struct{}{ /* lock */ }
		// Multiply priority by ten to create a buffer in between segments in
		// case the program needs to add additional (virtual) segments like
		// arrows or indicators after the explicit segment.
		go fn(&wg, sem, out, priority*10, p.config)
	}
	wg.Wait()
	close(sem)
	close(out)
	<-done
}

func consumer(w io.Writer, done chan struct{}, out chan Segment) {
	defer close(done)
	var segments []Segment
	for box := range out {
		if !box.Show || box.Text == "" {
			// Skip unnecessary segments.
			continue
		}
		// Prevent arbitrary code execution in subshell expressions.
		box.Text = strings.ReplaceAll(box.Text, "$", "\\$")
		box.Text = strings.ReplaceAll(box.Text, "`", "\\`")
		segments = append(segments, box)
		// Add an arrow pointing to the next segment; set colors later.
		//
		// ┌───┬───┬─────┬───┬─────────────┬───┬────────┬───┬───┬───┬───┐
		// │ ~ │ > │ ... │ > │ powergoline │ > │ foobar │ > │ $ │ > │   │
		// └───┴───┴─────┴───┴─────────────┴───┴────────┴───┴───┴───┴───┘
		//       ▲         ▲                 ▲            ▲       ▲   ▲
		//       │         │                 │            │       │   │
		//     arrow     arrow             arrow        arrow   arrow empty
		arrow := Segment{Kind: ArrowBox, Index: box.Index + 1, Text: uE0B0}
		segments = append(segments, arrow)
	}
	// Sort segments based on their original priority.
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].Index < segments[j].Index
	})
	// Once sorted, check the list again and, if the segment is an arrow, then
	// set the correct foreground and background colors. Foreground color must
	// be the background color of the previous segment. Background color must
	// be the background color of the next segment, if exists.
	for i := 1; i+1 < len(segments); i += 2 {
		if segments[i].Kind == ArrowBox {
			segments[i].Fg = segments[i-1].Bg
			segments[i].Bg = segments[i+1].Bg
			// Replace type of arrow if between plugin outputs.
			if segments[i-1].Kind == PluginBox && segments[i+1].Kind == PluginBox {
				segments[i].Fg = -1
				segments[i].Text = uE0B1
			}
		}
	}
	for _, box := range segments {
		if box.Show || box.Kind == ArrowBox {
			printOneSegment(w, box)
		}
	}
}

func printOneSegment(w io.Writer, seg Segment) {
	var color string
	fore := fmt.Sprintf("%03d", seg.Fg)
	back := fmt.Sprintf("%03d", seg.Bg)
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
		fmt.Fprint(w, "\\[\\e["+color+"m\\]"+seg.Text+"\\[\\e[0m\\]")
	} else {
		fmt.Fprint(w, seg.Text)
	}
}

// segmentDatetime prints the current date and time.
func segmentDatetime(wg *sync.WaitGroup, sem chan struct{}, out chan Segment, priority int, config Config) {
	defer wg.Done()
	defer func() { <-sem }()
	if !config.TimeOn {
		return
	}
	out <- Segment{Kind: TextBox, Index: priority, Show: true, Fg: config.TimeFg, Bg: config.TimeBg, Text: u0020 + time.Now().Format(config.TimeFmt) + u0020}
}

// segmentUsername prints the name of the current system user, e.g. root.
func segmentUsername(wg *sync.WaitGroup, sem chan struct{}, out chan Segment, priority int, config Config) {
	defer wg.Done()
	defer func() { <-sem }()
	if !config.UserOn {
		return
	}
	out <- Segment{Kind: TextBox, Index: priority, Show: true, Fg: config.UserFg, Bg: config.UserBg, Text: u0020 + "\\u" + u0020}
}

// segmentHostname prints the name of this system.
func segmentHostname(wg *sync.WaitGroup, sem chan struct{}, out chan Segment, priority int, config Config) {
	defer wg.Done()
	defer func() { <-sem }()
	if !config.HostOn {
		return
	}
	out <- Segment{Kind: TextBox, Index: priority, Show: true, Fg: config.HostFg, Bg: config.HostBg, Text: u0020 + "\\h" + u0020}
}

// SEP is same as os.PathSeparator but as a string.
const SEP string = "/"

// segmentDirectories prints the current location of the user in the system.
func segmentDirectories(wg *sync.WaitGroup, sem chan struct{}, out chan Segment, priority int, config Config) {
	defer wg.Done()
	defer func() { <-sem }()
	if !config.CwdOn {
		return
	}

	// Do not use os.UserHomeDir() and os.Getwd() as they resolve the path on
	// disk and may expand symbolic links or return an absolute path different
	// from what the user typed. We need the shell-exported HOME and PWD so the
	// prompt matches the exact working directory shown by the shell (including
	// symlinks and automount views) when composing PS1.
	homedir := os.Getenv("HOME")
	workdir := os.Getenv("PWD")

	// start with the entire folder path, then reduce as we remove sections.
	subfolders := workdir

	// first character in the folder path, e.g. / (forward-slash) or ~ (tilde).
	root := SEP

	if workdir == SEP {
		// User is at the root of the file system, so simply print a forward slash.
		subfolders = ""
	} else if workdir == homedir {
		// Add a tilde to represent that we are inside the home directory.
		root = "~"
		subfolders = ""
	} else if strings.HasPrefix(workdir, homedir) {
		root = "~"
		// Remove homedir from workdir and decorate the remaining folder path.
		subfolders = workdir[len(homedir)+1:]
	} else {
		// User is somewhere else in the system outside the $HOME directory.
		subfolders = subfolders[1:]
	}

	out <- Segment{Kind: FolderBox, Index: priority, Show: true, Fg: config.HomeFg, Bg: config.HomeBg, Text: u0020 + root + u0020}

	if subfolders != "" {
		// Plus one to account for the first characters in the entire folder
		// path that was removed in the conditions leading up to the creation
		// of the subfolders variable.
		nSections := strings.Count(subfolders, SEP) + 1
		if nSections > config.CwdN {
			// Path too long; replace parent folders with an ellipsis.
			sections := strings.Split(subfolders, SEP)
			sections = sections[nSections-config.CwdN : nSections]
			sections = append([]string{u2026}, sections...)
			subfolders = strings.Join(sections, SEP)
		}
		// Replace all folder separators (forward-slash) with light arrows.
		subfolders = strings.ReplaceAll(subfolders, SEP, u0020+uE0B1+u0020)
		out <- Segment{Kind: LockBox, Index: priority + 2, Show: true, Fg: config.CwdFg, Bg: config.CwdBg, Text: u0020 + subfolders + u0020}
	}

	if unix.Access(workdir, unix.W_OK) != nil {
		// Draw lock symbol if the current directory is read-only.
		out <- Segment{Kind: FolderBox, Index: priority + 4, Show: true, Fg: config.RodirFg, Bg: config.RodirBg, Text: u0020 + uE0A2 + u0020}
	}
}

// segmentRepoStatus prints the status of the current version control system.
func segmentRepoStatus(wg *sync.WaitGroup, sem chan struct{}, out chan Segment, priority int, config Config) {
	defer wg.Done()
	defer func() { <-sem }()
	if !config.RepoOn || slices.Contains(config.RepoExclude, os.Getenv("PWD")) {
		// Disabled globally or per-{git,hg}-repository.
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
		out <- Segment{Kind: RepoStatusBox, Index: priority, Show: true, Text: "hgrepo " + err.Error()}
		return
	}
	if len(status.Branch) == 0 {
		// hide as there is no information to show.
		return
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, " %s %s", uE0A0, status.Branch)
	if status.Ahead > 0 {
		fmt.Fprintf(&buf, " %s%d", u21E1, status.Ahead)
	}
	if status.Behind > 0 {
		fmt.Fprintf(&buf, " %s%d", u21E3, status.Behind)
	}
	if status.Added > 0 {
		fmt.Fprintf(&buf, " +%d", status.Added)
	}
	if status.Modified > 0 {
		fmt.Fprintf(&buf, " ~%d", status.Modified)
	}
	if status.Deleted > 0 {
		fmt.Fprintf(&buf, " -%d", status.Deleted)
	}
	fmt.Fprint(&buf, " ")
	out <- Segment{Kind: RepoStatusBox, Index: priority, Show: true, Fg: config.RepoFg, Bg: config.RepoBg, Text: buf.String()}
}

func segmentCallPlugins(wg *sync.WaitGroup, sem chan struct{}, out chan Segment, priority int, config Config) {
	defer wg.Done()
	defer func() { <-sem }()
	for i, command := range config.Plugins {
		wg.Add(1)
		sem <- struct{}{ /* lock */ }
		go segmentCallOnePlugin(wg, sem, out, i+101, config, command)
	}
}

func segmentCallOnePlugin(wg *sync.WaitGroup, sem chan struct{}, out chan Segment, priority int, config Config, cmd Plugin) {
	defer wg.Done()
	defer func() { <-sem }()
	start := time.Now()
	output, err := call(config.PluginTimeout, cmd.Name, cmd.Args...)
	runtime := time.Since(start)
	if config.Debug {
		fmt.Printf("%s ran in %s\n", cmd.Name, runtime)
	}
	if errors.Is(err, errEmptyOutput) {
		// hide as there is no output to show.
		out <- Segment{Kind: PluginBox, Index: priority, Show: false}
		return
	}
	if err != nil {
		// use error message instead.
		output = []byte(err.Error())
	}
	out <- Segment{Kind: PluginBox, Index: priority, Show: true, Fg: config.PluginFg, Bg: config.PluginBg, Text: u0020 + string(output) + u0020}
}

// segmentExitCode prints an indicator for root users.
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
func segmentExitCode(wg *sync.WaitGroup, sem chan struct{}, out chan Segment, _ int, config Config) {
	defer wg.Done()
	defer func() { <-sem }()
	var color int
	var symbol string
	status := config.StatusCode
	if os.Getuid() == 0 {
		symbol = config.SymbolRoot
	} else {
		symbol = config.SymbolUser
	}
	if status == 0 {
		color = config.StatusSuccess
	} else if status == 1 {
		color = config.StatusError
	} else if status == 2 {
		color = config.StatusMisuse
	} else if status == 126 {
		color = config.StatusCantExec
	} else if status == 127 {
		color = config.StatusNotFound
	} else if status == 128 {
		color = config.StatusInvalid
	} else if status > 128 && status != 130 && status < 255 {
		color = config.StatusErrSignal
	} else if status == 130 {
		color = config.StatusTerminated
	} else {
		color = config.StatusOutofrange
	}
	out <- Segment{Kind: ExitCodeBox, Index: 9999, Show: true, Fg: config.StatusFg, Bg: color, Text: u0020 + symbol + u0020}
}

func segmentInputSeparator(wg *sync.WaitGroup, sem chan struct{}, out chan Segment, _ int, config Config) {
	defer wg.Done()
	defer func() { <-sem }()
	// Add a whitespace at the end to separate the prompt from the user input.
	// This also guarantees the correct background color for the immediately
	// previous arrow segment, otherwise, the program panics with an index out
	// of bounds error.
	out <- Segment{Kind: LastBox, Index: 10000, Show: true, Fg: -1, Bg: -1, Text: u0020 /*+u000A*/}
}

// call executes an external command and returns the output.
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
