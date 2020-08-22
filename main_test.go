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

func compareRootSymbol(t *testing.T, status int, color string) {
	var buf bytes.Buffer

	a := []byte("\\[\\e[38;5;000;48;5;")
	b := []byte("m\\] r \\[\\e[0m\\]\\[\\e[38;5;")
	c := []byte("m\\]\ue0b0\\[\\e[0m\\] \n")

	p := NewPowergoline(Config{
		SymbolUser:       "r",
		SymbolRoot:       "s",
		StatusCode:       status,
		StatusSuccess:    0,
		StatusError:      111,
		StatusMisuse:     222,
		StatusCantExec:   126,
		StatusNotFound:   127,
		StatusInvalid:    128,
		StatusErrSignal:  333,
		StatusTerminated: 130,
		StatusOutofrange: 999,
	})

	var expected []byte

	p.RootSymbol()
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

func TestRootSymbol000(t *testing.T) { compareRootSymbol(t, 0, "000") }

func TestRootSymbol111(t *testing.T) { compareRootSymbol(t, 1, "111") }

func TestRootSymbol222(t *testing.T) { compareRootSymbol(t, 2, "222") }

func TestRootSymbol126(t *testing.T) { compareRootSymbol(t, 126, "126") }

func TestRootSymbol127(t *testing.T) { compareRootSymbol(t, 127, "127") }

func TestRootSymbol128(t *testing.T) { compareRootSymbol(t, 128, "128") }

func TestRootSymbol129(t *testing.T) { compareRootSymbol(t, 129, "333") }

func TestRootSymbol130(t *testing.T) { compareRootSymbol(t, 130, "130") }

func TestRootSymbol133(t *testing.T) { compareRootSymbol(t, 133, "333") }

func TestRootSymbol256(t *testing.T) { compareRootSymbol(t, 256, "999") }

func TestRootSymbolABC(t *testing.T) { compareRootSymbol(t, 300, "999") }

func BenchmarkAll(b *testing.B) {
	var buf bytes.Buffer

	for i := 0; i < b.N; i++ {
		p := NewPowergoline(Config{
			TimeOn:     true,
			UserOn:     true,
			HostOn:     true,
			CwdN:       3,
			RepoOn:     true,
			Plugins:    []Plugin{{Name: "echo"}},
			StatusCode: 0,
		})

		p.TermTitle()
		p.Datetime()
		p.Username()
		p.Hostname()
		p.Directories()
		p.RepoStatus()
		p.CallPlugins()
		p.RootSymbol()

		p.PrintSegments(&buf)
	}
}

func BenchmarkTermTitle(b *testing.B) {
	var buf bytes.Buffer

	for i := 0; i < b.N; i++ {
		p := NewPowergoline(Config{})
		p.TermTitle()
		p.PrintSegments(&buf)
	}
}

func BenchmarkDatetime(b *testing.B) {
	var buf bytes.Buffer

	for i := 0; i < b.N; i++ {
		p := NewPowergoline(Config{TimeOn: true})
		p.Datetime()
		p.PrintSegments(&buf)
	}
}

func BenchmarkUsername(b *testing.B) {
	var buf bytes.Buffer

	for i := 0; i < b.N; i++ {
		p := NewPowergoline(Config{UserOn: true})
		p.Username()
		p.PrintSegments(&buf)
	}
}

func BenchmarkHostname(b *testing.B) {
	var buf bytes.Buffer

	for i := 0; i < b.N; i++ {
		p := NewPowergoline(Config{HostOn: true})
		p.Hostname()
		p.PrintSegments(&buf)
	}
}

func BenchmarkDirectories(b *testing.B) {
	var buf bytes.Buffer

	for i := 0; i < b.N; i++ {
		p := NewPowergoline(Config{CwdN: 3})
		p.Directories()
		p.PrintSegments(&buf)
	}
}

func BenchmarkRepoStatus(b *testing.B) {
	var buf bytes.Buffer

	for i := 0; i < b.N; i++ {
		p := NewPowergoline(Config{RepoOn: true})
		p.RepoStatus()
		p.PrintSegments(&buf)
	}
}

func BenchmarkCallPlugins(b *testing.B) {
	var buf bytes.Buffer

	for i := 0; i < b.N; i++ {
		p := NewPowergoline(Config{
			Plugins: []Plugin{
				{Name: "echo"},
			},
		})
		p.CallPlugins()
		p.PrintSegments(&buf)
	}
}

func BenchmarkRootSymbol(b *testing.B) {
	var buf bytes.Buffer

	for i := 0; i < b.N; i++ {
		p := NewPowergoline(Config{StatusCode: 0})
		p.RootSymbol()
		p.PrintSegments(&buf)
	}
}
