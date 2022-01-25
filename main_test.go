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

func compareExitCode(t *testing.T, status int, color string) {
	var buf bytes.Buffer

	a := []byte("\\[\\e[38;5;000;48;5;")
	b := []byte("m\\] r \\[\\e[0m\\]\\[\\e[38;5;")
	c := []byte("m\\]\ue0b0\\[\\e[0m\\] ")

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
		t.Fatalf("invalid exit code output:\nExpected: `%q`\nActual:   `%q`", buf.Bytes(), expected)
	}
}

func TestExitCode(t *testing.T) {
	testCases := []struct {
		Name   string
		Status int
		Color  string
	}{
		{Name: "ExitCode000", Status: 0, Color: "000"},
		{Name: "ExitCode111", Status: 1, Color: "111"},
		{Name: "ExitCode222", Status: 2, Color: "222"},
		{Name: "ExitCode126", Status: 126, Color: "126"},
		{Name: "ExitCode127", Status: 127, Color: "127"},
		{Name: "ExitCode128", Status: 128, Color: "128"},
		{Name: "ExitCode129", Status: 129, Color: "333"},
		{Name: "ExitCode130", Status: 130, Color: "130"},
		{Name: "ExitCode133", Status: 133, Color: "333"},
		{Name: "ExitCode256", Status: 256, Color: "999"},
		{Name: "ExitCodeABC", Status: 300, Color: "999"},
	}

	for _, tx := range testCases {
		t.Run(tx.Name, func(t *testing.T) {
			compareExitCode(t, tx.Status, tx.Color)
		})
	}
}

func BenchmarkAll(b *testing.B) {
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		NewPowergoline(Config{
			TimeOn:     true,
			UserOn:     true,
			HostOn:     true,
			CwdN:       3,
			RepoOn:     true,
			Plugins:    []Plugin{{Name: "echo"}},
			StatusCode: 0,
		}).Render(&buf, []SegmentFunc{
			segmentDatetime,
			segmentUsername,
			segmentHostname,
			segmentDirectories,
			segmentRepoStatus,
			segmentCallPlugins,
			segmentExitCode,
		})
	}
}

func BenchmarkDatetime(b *testing.B) {
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		NewPowergoline(Config{TimeOn: true}).Render(&buf, []SegmentFunc{segmentDatetime})
	}
}

func BenchmarkUsername(b *testing.B) {
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		NewPowergoline(Config{UserOn: true}).Render(&buf, []SegmentFunc{segmentUsername})
	}
}

func BenchmarkHostname(b *testing.B) {
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		NewPowergoline(Config{HostOn: true}).Render(&buf, []SegmentFunc{segmentHostname})
	}
}

func BenchmarkDirectories(b *testing.B) {
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		NewPowergoline(Config{CwdN: 3}).Render(&buf, []SegmentFunc{segmentDirectories})
	}
}

func BenchmarkRepoStatus(b *testing.B) {
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		NewPowergoline(Config{RepoOn: true}).Render(&buf, []SegmentFunc{segmentRepoStatus})
	}
}

func BenchmarkCallPlugins(b *testing.B) {
	var buf bytes.Buffer

	for i := 0; i < b.N; i++ {
		NewPowergoline(Config{
			Plugins: []Plugin{
				{Name: "echo"},
			},
		}).Render(&buf, []SegmentFunc{
			segmentCallPlugins,
		})
	}
}

func BenchmarkExitCode(b *testing.B) {
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		NewPowergoline(Config{StatusCode: 0}).Render(&buf, []SegmentFunc{segmentExitCode})
	}
}
