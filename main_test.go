package main

import (
	"bytes"
	"testing"
)

func compareRepoStatus(t *testing.T, actual RepoStatus, expected RepoStatus) {
	if !bytes.Equal(actual.Branch, expected.Branch) {
		t.Fatalf("unexpected status.Branch `%s` != `%s`", actual.Branch, expected.Branch)
	}

	if actual.Ahead != expected.Ahead {
		t.Fatalf("unexpected status.Ahead %d != %d", actual.Ahead, expected.Ahead)
	}

	if actual.Behind != expected.Behind {
		t.Fatalf("unexpected status.Behind %d != %d", actual.Behind, expected.Behind)
	}

	if actual.Added != expected.Added {
		t.Fatalf("unexpected status.Added %d != %d", actual.Added, expected.Added)
	}

	if actual.Deleted != expected.Deleted {
		t.Fatalf("unexpected status.Deleted %d != %d", actual.Deleted, expected.Deleted)
	}

	if actual.Modified != expected.Modified {
		t.Fatalf("unexpected status.Modified %d != %d", actual.Modified, expected.Modified)
	}
}

func TestStatusGit(t *testing.T) {
	lines := [][]byte{
		[]byte("## master...origin/master"),
		[]byte("D  deleted.txt"),
		[]byte(" D missing.txt"),
		[]byte("M  patches.go"),
		[]byte(" M changes.go"),
		[]byte("A  newfile.sh"),
		[]byte("?? isadded.json"),
	}

	status, err := repoStatusGitParse(lines)

	if err != nil {
		t.Fatalf("repoStatusGitParse %s", err)
	}

	compareRepoStatus(t, status, RepoStatus{
		Branch:   []byte("master"),
		Modified: 2,
		Deleted:  2,
		Added:    2,
	})
}

func TestStatusGitAhead(t *testing.T) {
	lines := [][]byte{
		[]byte("## master...origin/master [ahead 5]"),
		[]byte("D  deleted.txt"),
		[]byte(" D missing.txt"),
		[]byte("M  patches.go"),
		[]byte(" M changes.go"),
		[]byte("A  newfile.sh"),
		[]byte("?? isadded.json"),
	}

	status, err := repoStatusGitParse(lines)

	if err != nil {
		t.Fatalf("repoStatusGitParse %s", err)
	}

	compareRepoStatus(t, status, RepoStatus{
		Branch:   []byte("master"),
		Modified: 2,
		Deleted:  2,
		Added:    2,
		Ahead:    5,
	})
}

func TestStatusGitBehind(t *testing.T) {
	lines := [][]byte{
		[]byte("## master...origin/master [behind 8]"),
		[]byte("D  deleted.txt"),
		[]byte(" D missing.txt"),
		[]byte("M  patches.go"),
		[]byte(" M changes.go"),
		[]byte("A  newfile.sh"),
		[]byte("?? isadded.json"),
	}

	status, err := repoStatusGitParse(lines)

	if err != nil {
		t.Fatalf("repoStatusGitParse %s", err)
	}

	compareRepoStatus(t, status, RepoStatus{
		Branch:   []byte("master"),
		Modified: 2,
		Deleted:  2,
		Added:    2,
		Behind:   8,
	})
}

func TestStatusGitAheadBehind(t *testing.T) {
	lines := [][]byte{
		[]byte("## master...origin/master [ahead 5, behind 8]"),
		[]byte("D  deleted.txt"),
		[]byte(" D missing.txt"),
		[]byte("M  patches.go"),
		[]byte(" M changes.go"),
		[]byte("A  newfile.sh"),
		[]byte("?? isadded.json"),
	}

	status, err := repoStatusGitParse(lines)

	if err != nil {
		t.Fatalf("repoStatusGitParse %s", err)
	}

	compareRepoStatus(t, status, RepoStatus{
		Branch:   []byte("master"),
		Modified: 2,
		Deleted:  2,
		Added:    2,
		Ahead:    5,
		Behind:   8,
	})
}

func TestStatusMercurial(t *testing.T) {
	lines := [][]byte{
		[]byte("R deleted.txt"),
		[]byte("! missing.txt"),
		[]byte("M patches.go"),
		[]byte("M changes.go"),
		[]byte("A newfile.sh"),
		[]byte("? isadded.json"),
	}

	status, err := repoStatusMercurialParse(lines)

	if err != nil {
		t.Fatalf("repoStatusMercurialParse %s", err)
	}

	compareRepoStatus(t, status, RepoStatus{
		Branch:   []byte("default"),
		Modified: 2,
		Deleted:  2,
		Added:    2,
	})
}

func TestStatusGitNoOrigin(t *testing.T) {
	lines := [][]byte{
		[]byte("## develop"),
	}

	status, err := repoStatusGitParse(lines)

	if err != nil {
		t.Fatalf("repoStatusGitParse %s", err)
	}

	compareRepoStatus(t, status, RepoStatus{
		Branch: []byte("develop"),
	})
}

func compareRootSymbol(t *testing.T, status string, color string) {
	var buf bytes.Buffer

	a := []byte("\\[\\e[48;5;")
	b := []byte("m\\] r \\[\\e[0m\\]\\[\\e[38;5;")
	c := []byte("m\\]\ue0b0\\[\\e[0m\\] \n")

	p := NewPowergoline(Config{
		Symbol: StatusSymbol{
			Regular:   "r",
			SuperUser: "s",
		},
		Status: StatusCode{
			Success:     "c000", // 0 - Operation success and generic status code.
			Failure:     "c111", // 1 - Catchall for general errors and failures.
			Misuse:      "c222", // 2 - Misuse of shell builtins, missing command or permission problem.
			Permission:  "c126", // 126 - Cannot execute command, permission problem, or not an executable.
			NotFound:    "c127", // 127 - Command not found, illegal path, or possible typo.
			InvalidExit: "c128", // 128 - Invalid argument to exit, only use range 0-255.
			Terminated:  "c130", // 130 - Script terminated by Control-C.
		},
	})

	var expected []byte

	p.RootSymbol(status)
	p.PrintSegments(&buf)

	expected = append(expected, a...)
	expected = append(expected, []byte(color)...)
	expected = append(expected, b...)
	expected = append(expected, []byte(color)...)
	expected = append(expected, c...)

	if !bytes.Equal(buf.Bytes(), expected) {
		t.Fatalf("invalid root symbol output:\nExpected: `%q`\nActual:   `%q`", buf.Bytes(), expected)
	}
}

func TestRootSymbol000(t *testing.T) { compareRootSymbol(t, "0", "c000") }

func TestRootSymbol111(t *testing.T) { compareRootSymbol(t, "1", "c111") }

func TestRootSymbol222(t *testing.T) { compareRootSymbol(t, "2", "c222") }

func TestRootSymbol126(t *testing.T) { compareRootSymbol(t, "126", "c126") }

func TestRootSymbol127(t *testing.T) { compareRootSymbol(t, "127", "c127") }

func TestRootSymbol128(t *testing.T) { compareRootSymbol(t, "128", "c128") }

func TestRootSymbol129(t *testing.T) { compareRootSymbol(t, "129", "c222") }

func TestRootSymbol130(t *testing.T) { compareRootSymbol(t, "130", "c130") }

func TestRootSymbol256(t *testing.T) { compareRootSymbol(t, "256", "c222") }

func TestRootSymbolABC(t *testing.T) { compareRootSymbol(t, "abc", "c222") }

func BenchmarkAll(b *testing.B) {
	var buf bytes.Buffer

	p := NewPowergoline(Config{
		Datetime:   SimpleConfig{On: true},
		Username:   SimpleConfig{On: true},
		Hostname:   SimpleConfig{On: true},
		Repository: SimpleConfig{On: true},
	})

	for i := 0; i < b.N; i++ {
		p.PrintSegments(&buf)
	}
}

func BenchmarkTermTitle(b *testing.B) {
	var buf bytes.Buffer

	p := NewPowergoline(Config{})

	for i := 0; i < b.N; i++ {
		p.TermTitle()
		p.PrintSegments(&buf)
	}
}

func BenchmarkDatetime(b *testing.B) {
	var buf bytes.Buffer

	p := NewPowergoline(Config{Datetime: SimpleConfig{On: true}})

	for i := 0; i < b.N; i++ {
		p.Datetime()
		p.PrintSegments(&buf)
	}
}

func BenchmarkUsername(b *testing.B) {
	var buf bytes.Buffer

	p := NewPowergoline(Config{Username: SimpleConfig{On: true}})

	for i := 0; i < b.N; i++ {
		p.Username()
		p.PrintSegments(&buf)
	}
}

func BenchmarkHostname(b *testing.B) {
	var buf bytes.Buffer

	p := NewPowergoline(Config{Hostname: SimpleConfig{On: true}})

	for i := 0; i < b.N; i++ {
		p.Hostname()
		p.PrintSegments(&buf)
	}
}

func BenchmarkDirectories(b *testing.B) {
	var buf bytes.Buffer

	p := NewPowergoline(Config{})

	for i := 0; i < b.N; i++ {
		p.Directories()
		p.PrintSegments(&buf)
	}
}

func BenchmarkRepoStatus(b *testing.B) {
	var buf bytes.Buffer

	p := NewPowergoline(Config{Repository: SimpleConfig{On: true}})

	for i := 0; i < b.N; i++ {
		p.RepoStatus()
		p.PrintSegments(&buf)
	}
}

func BenchmarkCallPlugins(b *testing.B) {
	var buf bytes.Buffer

	p := NewPowergoline(Config{
		Plugins: []Plugin{
			{Command: "echo"},
		},
	})

	for i := 0; i < b.N; i++ {
		p.CallPlugins()
		p.PrintSegments(&buf)
	}
}

func BenchmarkRootSymbol(b *testing.B) {
	var buf bytes.Buffer

	p := NewPowergoline(Config{})

	for i := 0; i < b.N; i++ {
		p.RootSymbol("0")
		p.PrintSegments(&buf)
	}
}
