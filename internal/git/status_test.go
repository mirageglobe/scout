package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// initRepo creates a temp git repo with one committed file and returns its dir.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "test")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "a.txt")
	run("commit", "-m", "initial commit")
	return dir
}

func TestDiffAndLog(t *testing.T) {
	dir := initRepo(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// no changes yet -> empty diff
	if out, err := Diff(ctx, dir, "a.txt"); err != nil || strings.TrimSpace(out) != "" {
		t.Errorf("clean Diff = %q, err=%v; want empty", out, err)
	}

	// modify the file -> diff should surface the change
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := Diff(ctx, dir, "a.txt")
	if err != nil {
		t.Fatalf("Diff after edit: %v", err)
	}
	if !strings.Contains(out, "-one") || !strings.Contains(out, "+two") {
		t.Errorf("Diff = %q, want it to show -one/+two", out)
	}

	// log should list the initial commit
	logOut, err := Log(ctx, dir, "a.txt")
	if err != nil {
		t.Fatalf("Log: %v", err)
	}
	if !strings.Contains(logOut, "initial commit") {
		t.Errorf("Log = %q, want it to mention the initial commit", logOut)
	}
}
