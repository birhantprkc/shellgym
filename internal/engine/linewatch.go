package engine

import (
	"bufio"
	"bytes"
	"debug/elf"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// LineEvent is one command line read by an interactive shell - the text
// the student typed at the prompt, before the shell parsed it. This is
// the only place the shell's own syntax (`&&` vs `;`, quotes, pipes,
// builtins) is observable: by the time a command execs, the line is gone.
type LineEvent struct {
	Seq  uint64    `json:"seq"`
	Time time.Time `json:"time"`
	PID  int       `json:"pid"` // the shell that read the line
	UID  int       `json:"uid"` // -1 = unknown (the shell was gone before /proc could be read)
	// TTYNr is the shell's controlling terminal (0 = none, -1 = unknown).
	TTYNr int    `json:"ttyNr"`
	Comm  string `json:"comm"`
	Line  string `json:"line"`
}

// LineWatcher records the command lines interactive bash shells read,
// sourced from a kernel uretprobe on bash's readline() (the same trick as
// bpftrace's bashreadline): the probe fires when readline returns, the
// kernel copies the returned string into the trace buffer, and the daemon
// streams that buffer from a private tracefs instance. Nothing in the
// student's shell is touched - no prompt hooks, no wrappers.
//
// The capability is optional: it needs root, tracefs with uprobe events
// (CONFIG_UPROBE_EVENTS), and a bash that exposes the readline symbol
// (statically linked or via libreadline). When Start fails the daemon
// runs without it, and units that `requires: [readline]` are marked
// unsupported.
type LineWatcher struct {
	*eventRing[LineEvent]

	// Source is "uprobe" when the probe is active, "" when line watching
	// is unavailable.
	Source string
	// Target is the probed binary and the readline offset in it (for logs).
	Target string

	tracefs  string
	instance string
	fd       int
	done     chan struct{} // closed when readLoop has exited
}

// probeGroup/probeName identify the uprobe event in tracefs
// (<tracefs>/events/<group>/<name>). Fixed names let a restarted daemon
// find and remove a probe a crashed one left behind.
const (
	probeGroup = "shellgym"
	probeName  = "readline"
)

func NewLineWatcher() *LineWatcher {
	return &LineWatcher{eventRing: newEventRing(ringSize,
		func(ev LineEvent) uint64 { return ev.Seq },
		func(ev LineEvent, seq uint64, t time.Time) LineEvent { ev.Seq, ev.Time = seq, t; return ev }), fd: -1}
}

// Start installs the readline uretprobe for the given shell binary and
// begins streaming lines. It returns an error, leaving nothing behind in
// tracefs, when any prerequisite is missing.
func (w *LineWatcher) Start(shellPath string) error {
	tracefs, err := findTracefs()
	if err != nil {
		return err
	}
	target, offset, err := resolveReadline(shellPath)
	if err != nil {
		return err
	}
	w.tracefs = tracefs
	w.instance = filepath.Join(tracefs, "instances", probeGroup)
	w.Target = fmt.Sprintf("%s+0x%x", target, offset)

	// A previous daemon may have died without cleaning up: the probe and
	// the instance are kernel state and outlive processes.
	w.teardown()

	spec := fmt.Sprintf("r:%s/%s %s:0x%x line=+0($retval):string\n", probeGroup, probeName, target, offset)
	if err := appendFile(filepath.Join(tracefs, "uprobe_events"), spec); err != nil {
		return fmt.Errorf("register uprobe: %w", err)
	}
	if err := os.Mkdir(w.instance, 0o755); err != nil && !errors.Is(err, os.ErrExist) {
		w.teardown()
		return fmt.Errorf("create tracefs instance: %w", err)
	}
	if err := writeFile(filepath.Join(w.instance, "events", probeGroup, probeName, "enable"), "1\n"); err != nil {
		w.teardown()
		return fmt.Errorf("enable uprobe: %w", err)
	}
	// Non-blocking + poll so Close can stop the reader: a blocking read on
	// trace_pipe is uninterruptible from Go's side.
	fd, err := unix.Open(filepath.Join(w.instance, "trace_pipe"), unix.O_RDONLY|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		w.teardown()
		return fmt.Errorf("open trace_pipe: %w", err)
	}
	w.fd = fd
	w.Source = "uprobe"
	w.done = make(chan struct{})
	go w.readLoop()
	return nil
}

// Close stops the reader and removes the probe and the tracefs instance.
// It waits for the reader to leave its poll first: a thread inside
// poll(2) keeps the instance's trace_pipe referenced, and while it is,
// the kernel refuses both removing the instance and opening the pipe
// again (EBUSY) - a restarted daemon would then start without the
// capability.
func (w *LineWatcher) Close() {
	w.eventRing.Close()
	if w.Source == "" {
		return
	}
	w.Source = ""
	select {
	case <-w.done:
	case <-time.After(2 * time.Second):
	}
	w.teardown()
}

// teardown removes every trace of the watcher from tracefs, in the order
// the kernel requires (a probe cannot be deleted while enabled anywhere).
// Every step is best-effort: it runs on failed starts and on stale state.
func (w *LineWatcher) teardown() {
	if w.fd >= 0 {
		_ = unix.Close(w.fd)
		w.fd = -1
	}
	enable := filepath.Join("events", probeGroup, probeName, "enable")
	_ = writeFile(filepath.Join(w.instance, enable), "0\n")
	_ = writeFile(filepath.Join(w.tracefs, enable), "0\n")
	_ = os.Remove(w.instance)
	_ = appendFile(filepath.Join(w.tracefs, "uprobe_events"), fmt.Sprintf("-:%s/%s\n", probeGroup, probeName))
}

func (w *LineWatcher) readLoop() {
	defer close(w.done)
	buf := make([]byte, 64*1024)
	var pending []byte
	for !w.isClosed() {
		fds := []unix.PollFd{{Fd: int32(w.fd), Events: unix.POLLIN}}
		n, err := unix.Poll(fds, 200)
		if err != nil && err != unix.EINTR {
			log.Printf("linewatch: poll: %v; line-based checks disabled", err)
			return
		}
		if n <= 0 {
			continue
		}
		n, err = unix.Read(w.fd, buf)
		if err != nil {
			if err == unix.EAGAIN || err == unix.EINTR {
				continue
			}
			if !w.isClosed() {
				log.Printf("linewatch: read: %v; line-based checks disabled", err)
			}
			return
		}
		pending = append(pending, buf[:n]...)
		for {
			i := bytes.IndexByte(pending, '\n')
			if i < 0 {
				break
			}
			line := string(pending[:i])
			pending = pending[i+1:]
			if ev, ok := parseTraceLine(line); ok {
				harvestShell(&ev)
				w.publish(ev)
			}
		}
	}
}

// traceLineRe matches one trace_pipe record of the readline probe, e.g.
//
//	bash-1103  [002] DNZff  39.302011: readline: (0x55c5... <- 0x55c5...) line="sleep 1 && hostname"
//
// The comm is at most 16 bytes and may itself contain dashes, so the
// pid is anchored on the "[cpu]" column that follows it.
var traceLineRe = regexp.MustCompile(`^\s*(.{1,16}?)-(\d+)\s+\[\d+\]\s+\S+\s+[\d.]+: ` + probeName + `: \([^)]*\) line="(.*)"$`)

// parseTraceLine extracts the shell pid and the typed line from one trace
// record. Records of other kinds (lost-event markers, other probes) and
// unreadable lines (a NULL return on EOF shows up as "(fault)") yield
// false.
func parseTraceLine(s string) (LineEvent, bool) {
	m := traceLineRe.FindStringSubmatch(s)
	if m == nil {
		return LineEvent{}, false
	}
	pid, err := strconv.Atoi(m[2])
	if err != nil || m[3] == "(fault)" {
		return LineEvent{}, false
	}
	return LineEvent{PID: pid, UID: -1, TTYNr: -1, Comm: m[1], Line: m[3]}, true
}

// harvestShell fills uid and tty from /proc. Shells are long-lived, so
// unlike fast exec'ed commands they are practically always still there.
func harvestShell(ev *LineEvent) {
	if status, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", ev.PID)); err == nil {
		for _, line := range strings.Split(string(status), "\n") {
			if v, ok := strings.CutPrefix(line, "Uid:"); ok {
				if f := strings.Fields(v); len(f) > 0 {
					ev.UID, _ = strconv.Atoi(f[0])
				}
				break
			}
		}
	}
	if stat, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", ev.PID)); err == nil {
		if f := statFields(string(stat)); len(f) > 6 {
			if v, err := strconv.Atoi(f[6]); err == nil {
				ev.TTYNr = v
			}
		}
	}
}

// findTracefs locates a tracefs mount with uprobe support.
func findTracefs() (string, error) {
	for _, dir := range []string{"/sys/kernel/tracing", "/sys/kernel/debug/tracing"} {
		if _, err := os.Stat(filepath.Join(dir, "uprobe_events")); err == nil {
			return dir, nil
		}
	}
	return "", errors.New("tracefs with uprobe_events not found (needs root and CONFIG_UPROBE_EVENTS)")
}

// resolveReadline finds the binary that carries the shell's readline()
// and the symbol's file offset in it: bash itself when readline is linked
// in statically (Debian, Ubuntu), or the libreadline shared object bash
// loads (Fedora, Rocky - built with --with-installed-readline).
func resolveReadline(shellPath string) (target string, offset uint64, err error) {
	shellPath, err = filepath.EvalSymlinks(shellPath)
	if err != nil {
		return "", 0, fmt.Errorf("shell %s: %w", shellPath, err)
	}
	if filepath.Base(shellPath) != "bash" {
		return "", 0, fmt.Errorf("shell %s is not bash (only bash exposes readline)", shellPath)
	}
	if off, ok, err := elfSymbolOffset(shellPath, "readline"); err != nil {
		return "", 0, err
	} else if ok {
		return shellPath, off, nil
	}
	lib, err := sharedLib(shellPath, "libreadline")
	if err != nil {
		return "", 0, fmt.Errorf("%s has no readline symbol and loads no libreadline: %w", shellPath, err)
	}
	if off, ok, err := elfSymbolOffset(lib, "readline"); err != nil {
		return "", 0, err
	} else if ok {
		return lib, off, nil
	}
	return "", 0, fmt.Errorf("no readline symbol in %s or %s", shellPath, lib)
}

// elfSymbolOffset returns the file offset of a defined function symbol,
// looking at the dynamic symbol table first (all that a stripped bash
// keeps) and the full symbol table second.
func elfSymbolOffset(path, name string) (uint64, bool, error) {
	f, err := elf.Open(path)
	if err != nil {
		return 0, false, fmt.Errorf("%s: %w", path, err)
	}
	defer f.Close()
	var syms []elf.Symbol
	if ds, err := f.DynamicSymbols(); err == nil {
		syms = append(syms, ds...)
	}
	if ss, err := f.Symbols(); err == nil {
		syms = append(syms, ss...)
	}
	for _, s := range syms {
		if s.Name != name || s.Value == 0 || s.Section == elf.SHN_UNDEF {
			continue
		}
		for _, p := range f.Progs {
			if p.Type == elf.PT_LOAD && s.Value >= p.Vaddr && s.Value < p.Vaddr+p.Filesz {
				return s.Value - p.Vaddr + p.Off, true, nil
			}
		}
	}
	return 0, false, nil
}

// sharedLib resolves the path of a shared object the binary loads, by
// name prefix (e.g. "libreadline"), via the dynamic loader's own view.
func sharedLib(binary, prefix string) (string, error) {
	out, err := exec.Command("ldd", binary).Output()
	if err != nil {
		return "", fmt.Errorf("ldd: %w", err)
	}
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		// "libreadline.so.8 => /lib/x86_64-linux-gnu/libreadline.so.8 (0x...)"
		if len(f) >= 3 && strings.HasPrefix(f[0], prefix) && f[1] == "=>" && strings.HasPrefix(f[2], "/") {
			return filepath.EvalSymlinks(f[2])
		}
	}
	return "", fmt.Errorf("no %s among %s's shared libraries", prefix, binary)
}

func writeFile(path, s string) error {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(s)
	return err
}

func appendFile(path, s string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(s)
	return err
}
