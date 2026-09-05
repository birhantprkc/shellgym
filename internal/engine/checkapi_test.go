package engine

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/iximiuz/labs-content/tools/shellgym/internal/bus"
	"github.com/iximiuz/labs-content/tools/shellgym/internal/content"
	"github.com/iximiuz/labs-content/tools/shellgym/internal/state"
)

// execWait runs one /exec/wait request against a synthetic watcher.
func execWaitOnce(t *testing.T, api *checkAPI, req ExecWaitRequest) ExecWaitResponse {
	t.Helper()
	body, _ := json.Marshal(req)
	r := httptest.NewRequest("POST", "http://gym/exec/wait", bytes.NewReader(body))
	w := httptest.NewRecorder()
	api.handleExecWait(w, r)
	if w.Code != 200 {
		t.Fatalf("exec/wait returned %d: %s", w.Code, w.Body.String())
	}
	var out ExecWaitResponse
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

// The joined-argv regex cannot tell `date '+%A %d'` (one quoted argument)
// from `date +%A %d` (two arguments) - both join to the same string. The
// argc filter can, and that is what quoting reps rely on.
func TestExecWaitArgcDistinguishesQuoting(t *testing.T) {
	w := NewExecWatcher()
	api := &checkAPI{watcher: w, shellUID: 1000}

	// Unquoted form: three argv elements.
	w.publish(ExecEvent{PID: 1, UID: 1000, TTYNr: 3, Argv: []string{"date", "+%A", "%d"}})

	req := ExecWaitRequest{Regex: `^date \+%A %d$`, Argc: 2, TimeoutSec: 0.05}
	if execWaitOnce(t, api, req).Matched {
		t.Fatal("argc=2 matched a 3-element argv")
	}
	// Without the argc constraint the same event matches.
	if !execWaitOnce(t, api, ExecWaitRequest{Regex: `^date \+%A %d$`, TimeoutSec: 0.05}).Matched {
		t.Fatal("regex alone should match the unquoted form")
	}

	// Quoted form: two argv elements, same joined string.
	w.publish(ExecEvent{PID: 2, UID: 1000, TTYNr: 3, Argv: []string{"date", "+%A %d"}})
	if !execWaitOnce(t, api, req).Matched {
		t.Fatal("argc=2 did not match the quoted form")
	}
}

// A check that branches on WHICH command matched needs two things: the
// response must carry the matched argv, and --latest must prefer the
// newest buffered answer - otherwise the first (wrong) answer since
// activation would keep winning after every hint_exit restart.
func TestExecWaitLatestPrefersNewestMatch(t *testing.T) {
	w := NewExecWatcher()
	api := &checkAPI{watcher: w, shellUID: 1000}
	w.publish(ExecEvent{PID: 1, UID: 1000, TTYNr: 3, Argv: []string{"whoami"}})
	w.publish(ExecEvent{PID: 2, UID: 1000, TTYNr: 3, Argv: []string{"hostname"}})

	re := `(^|/)(hostname|whoami)$`
	oldest := execWaitOnce(t, api, ExecWaitRequest{Regex: re, TimeoutSec: 0.05})
	if !oldest.Matched || oldest.Event.Argv[0] != "whoami" {
		t.Fatalf("default should match the oldest event, got %+v", oldest.Event)
	}
	newest := execWaitOnce(t, api, ExecWaitRequest{Regex: re, Latest: true, TimeoutSec: 0.05})
	if !newest.Matched || newest.Event.Argv[0] != "hostname" {
		t.Fatalf("--latest should match the newest event, got %+v", newest.Event)
	}
}

// readSSEEvent scans lines from an SSE stream up to the next blank line and
// returns the "event:" and "data:" fields.
func readSSEEvent(t *testing.T, scanner *bufio.Scanner) (event, data string) {
	t.Helper()
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "event: "):
			event = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			data = strings.TrimPrefix(line, "data: ")
		case line == "" && event != "":
			return event, data
		}
	}
	t.Fatal("SSE stream ended before a full event was received")
	return "", ""
}

// A supervisor watching /units/watch must never miss a unit completion, even
// across a reconnect - the snapshot on connect and one event per subsequent
// change are what make that possible.
func TestUnitsWatchStreamsSnapshotThenUpdates(t *testing.T) {
	path := &content.Path{ID: "p", Modules: []*content.Module{{ID: "m", Units: []*content.Unit{
		{ID: "m/a", ModuleID: "m"},
		{ID: "m/b", ModuleID: "m", Unsupported: true},
	}}}}
	st, err := state.Open(t.TempDir(), "p")
	if err != nil {
		t.Fatal(err)
	}
	b := bus.New()
	eng := New(path, st, b, NewExecWatcher(), Options{})
	api := &checkAPI{eng: eng}

	srv := httptest.NewServer(http.HandlerFunc(api.handleUnitsWatch))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)

	event, data := readSSEEvent(t, scanner)
	if event != "snapshot" {
		t.Fatalf("first event = %q, want %q", event, "snapshot")
	}
	var snap unitsSnapshot
	if err := json.Unmarshal([]byte(data), &snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if snap.Path != "p" {
		t.Fatalf("snapshot path = %q, want %q", snap.Path, "p")
	}
	wantUnits := map[string]string{"m/a": "pending", "m/b": "unsupported"}
	if !reflect.DeepEqual(snap.Units, wantUnits) {
		t.Fatalf("snapshot units = %+v, want %+v", snap.Units, wantUnits)
	}

	// A non-unit event must not be forwarded; the unit event that follows
	// must be the very next thing the client sees.
	b.Publish(bus.Event{Type: "task", Data: TaskEvent{Unit: "m/a", Task: "x", Status: "completed"}})
	b.Publish(bus.Event{Type: "unit", Data: UnitEvent{Unit: "m/a", Status: "completed"}})

	event, data = readSSEEvent(t, scanner)
	if event != "unit" {
		t.Fatalf("second event = %q, want %q", event, "unit")
	}
	var us unitStatus
	if err := json.Unmarshal([]byte(data), &us); err != nil {
		t.Fatalf("decode unit status: %v", err)
	}
	if us.ID != "m/a" || us.Status != "completed" {
		t.Fatalf("unit status = %+v, want {m/a completed}", us)
	}

	cancel()
	for scanner.Scan() {
		// drain until the cancelled context closes the body
	}
}
