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

	NewPowergoline(Config{
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
	}).Render(&buf, []SegmentFunc{segmentExitCode})

	var expected []byte
	expected = append(expected, a...)
	expected = append(expected, []byte(color)...)
	expected = append(expected, b...)
	expected = append(expected, []byte(color)...)
	expected = append(expected, c...)

	if !bytes.Equal(buf.Bytes(), expected) {
		t.Fatalf("invalid root symbol output:\nExpected: `%q`\nActual:   `%q`", buf.Bytes(), expected)
	}
}

func TestRootSymbol(t *testing.T) {
	testCases := []struct {
		Name   string
		Status int
		Color  string
	}{
		{Name: "RootSymbol000", Status: 0, Color: "000"},
		{Name: "RootSymbol111", Status: 1, Color: "111"},
		{Name: "RootSymbol222", Status: 2, Color: "222"},
		{Name: "RootSymbol126", Status: 126, Color: "126"},
		{Name: "RootSymbol127", Status: 127, Color: "127"},
		{Name: "RootSymbol128", Status: 128, Color: "128"},
		{Name: "RootSymbol129", Status: 129, Color: "333"},
		{Name: "RootSymbol130", Status: 130, Color: "130"},
		{Name: "RootSymbol133", Status: 133, Color: "333"},
		{Name: "RootSymbol256", Status: 256, Color: "999"},
		{Name: "RootSymbolABC", Status: 300, Color: "999"},
	}

	for _, tx := range testCases {
		t.Run(tx.Name, func(t *testing.T) {
			compareRootSymbol(t, tx.Status, tx.Color)
		})
	}
}

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
