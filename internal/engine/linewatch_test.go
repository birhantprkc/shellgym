package engine

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/creack/pty"
	"golang.org/x/sys/unix"
)

func TestParseTraceLine(t *testing.T) {
	cases := []struct {
		in   string
		pid  int
		line string
		ok   bool
	}{
		{`            bash-1103    [002] DNZff    39.302011: readline: (0x55c510f35b7f <- 0x55c510fde690) line="sleep 1 && hostname"`, 1103, "sleep 1 && hostname", true},
		// A comm with a dash: the pid is anchored on the [cpu] column.
		{`         my-bash-77    [000] d..1.   100.5: readline: (0x1 <- 0x2) line="echo "a" ; b"`, 77, `echo "a" ; b`, true},
		// An empty line (Enter on an empty prompt) is a line too.
		{`            bash-1103    [002] DNZff    41.306734: readline: (0x1 <- 0x2) line=""`, 1103, "", true},
		// EOF: readline returned NULL, the kernel could not fetch the string.
		{`            bash-1103    [002] DNZff    41.306734: readline: (0x1 <- 0x2) line="(fault)"`, 0, "", false},
		{`CPU:2 [LOST 12 EVENTS]`, 0, "", false},
		{``, 0, "", false},
	}
	for _, c := range cases {
		ev, ok := parseTraceLine(c.in)
		if ok != c.ok {
			t.Errorf("%q: ok = %v, want %v", c.in, ok, c.ok)
			continue
		}
		if ok && (ev.PID != c.pid || ev.Line != c.line) {
			t.Errorf("%q: got pid=%d line=%q, want pid=%d line=%q", c.in, ev.PID, ev.Line, c.pid, c.line)
		}
	}
}

// bash exposes readline either from its own image (Debian/Ubuntu link it
// statically) or via libreadline (Fedora/Rocky); either way the symbol
// must resolve to a file offset inside a loadable segment.
func TestResolveReadline(t *testing.T) {
	if _, err := os.Stat("/bin/bash"); err != nil {
		t.Skip("no /bin/bash on this host")
	}
	target, off, err := resolveReadline("/bin/bash")
	if err != nil {
		t.Fatal(err)
	}
	if off == 0 || !strings.HasPrefix(target, "/") {
		t.Fatalf("bad resolution: target=%q offset=%#x", target, off)
	}
	t.Logf("readline at %s+%#x", target, off)

	if _, _, err := resolveReadline("/bin/sh"); err == nil {
		if real, _ := exec.Command("readlink", "-f", "/bin/sh").Output(); !strings.HasSuffix(strings.TrimSpace(string(real)), "/bash") {
			t.Fatalf("resolving a non-bash shell should fail")
		}
	}
}

// startLineWatcher installs the real readline uprobe, skipping the test
// when the host cannot (no root / no tracefs). Run the suite as root on
// the playground to exercise it.
func startLineWatcher(t *testing.T) *LineWatcher {
	t.Helper()
	w := NewLineWatcher()
	if err := w.Start("/bin/bash"); err != nil {
		t.Skipf("line watcher unavailable: %v", err)
	}
	t.Cleanup(w.Close)
	return w
}

// The whole point of the watcher: an interactive bash on a pty reads a
// line, and the daemon sees the line as typed - operator included - while
// a non-interactive bash running the same text produces nothing.
func TestLineWatcherSeesTypedLine(t *testing.T) {
	w := startLineWatcher(t)
	after := w.Seq()

	_ = exec.Command("bash", "-c", "true && echo shellgym-noninteractive-4711").Run()

	cmd := exec.Command("bash", "--norc", "-i")
	cmd.Env = append(os.Environ(), "PS1=$ ", "TERM=dumb")
	f, err := pty.Start(cmd)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill(); _, _ = cmd.Process.Wait(); _ = f.Close() }()
	go func() {
		buf := make([]byte, 4096)
		for {
			if _, err := f.Read(buf); err != nil {
				return
			}
		}
	}()
	time.Sleep(300 * time.Millisecond)
	if _, err := f.WriteString("true && echo shellgym-interactive-4711\n"); err != nil {
		t.Fatal(err)
	}

	ev, ok := w.WaitMatch(context.Background(), after, time.Now().Add(5*time.Second), func(ev LineEvent) bool {
		return strings.Contains(ev.Line, "shellgym-interactive-4711")
	})
	if !ok {
		t.Fatal("typed line not observed by the watcher")
	}
	if ev.Line != "true && echo shellgym-interactive-4711" || ev.PID != cmd.Process.Pid || ev.TTYNr <= 0 || ev.UID != os.Getuid() {
		t.Errorf("bad event: %+v", ev)
	}
	if _, ok := w.WaitMatch(context.Background(), after, time.Now().Add(200*time.Millisecond), func(ev LineEvent) bool {
		return strings.Contains(ev.Line, "noninteractive")
	}); ok {
		t.Error("a non-interactive bash must not produce line events")
	}
}

// A restarted daemon must recover from the probe and the tracefs instance
// a crashed one left behind (kernel state outlives the process), and a
// clean Close must leave nothing behind for the next one.
func TestLineWatcherRestartOverStaleProbe(t *testing.T) {
	w1 := startLineWatcher(t)
	// Simulate the crash: the process (and so its trace_pipe fd) is gone,
	// the probe and the instance are not.
	w1.eventRing.Close()
	<-w1.done
	_ = unix.Close(w1.fd)
	w1.fd = -1
	if _, err := os.Stat(w1.instance); err != nil {
		t.Fatalf("stale instance should still exist: %v", err)
	}

	w2 := NewLineWatcher()
	if err := w2.Start("/bin/bash"); err != nil {
		t.Fatalf("restart over stale probe: %v", err)
	}
	w2.Close()
	if _, err := os.Stat(w2.instance); err == nil {
		t.Error("tracefs instance left behind after Close")
	}
	if raw, _ := os.ReadFile(filepath.Join(w2.tracefs, "uprobe_events")); strings.Contains(string(raw), probeGroup+"/"+probeName) {
		t.Errorf("uprobe left behind after Close: %s", raw)
	}
}
