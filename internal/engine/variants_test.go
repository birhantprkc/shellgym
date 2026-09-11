package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iximiuz/labs-content/tools/shellgym/internal/bus"
	"github.com/iximiuz/labs-content/tools/shellgym/internal/content"
	"github.com/iximiuz/labs-content/tools/shellgym/internal/state"
)

const trivialUnit = `---
title: Trivial
tasks:
  t:
    check: |
      true
---
::task
Waiting...
::
`

// variantPath loads a path with two variant keys and opens its state dir.
func variantPath(t *testing.T) (*content.Path, string) {
	t.Helper()
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "content", "path.yaml"), "id: vtest\ntitle: vtest\n")
	for rel, extra := range map[string]string{
		"010.m/010.always/unit.md": "",
		"010.m/020.forest/unit.md": "variant: scene=forest",
		"010.m/030.meadow/unit.md": "variant: scene=meadow",
		"010.m/040.night/unit.md":  "variant: time=night",
		"010.m/050.day/unit.md":    "variant: time=day",
	} {
		mustWrite(t, filepath.Join(dir, "content", rel), strings.Replace(trivialUnit, "title: Trivial", "title: Trivial\n"+extra, 1))
	}
	p, err := content.Load(filepath.Join(dir, "content"), "ubuntu", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return p, filepath.Join(dir, "state")
}

func openEngine(t *testing.T, p *content.Path, stateDir string) *Engine {
	t.Helper()
	st, err := state.Open(stateDir, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	eng := New(p, st, bus.New(), nil, Options{ChecksDir: stateDir, SockPath: filepath.Join(stateDir, "x.sock")})
	t.Cleanup(eng.Shutdown)
	return eng
}

func TestVariantDrawIsPersistedAndApplied(t *testing.T) {
	p, stateDir := variantPath(t)
	eng := openEngine(t, p, stateDir)

	keys, picks := eng.Variants()
	if len(keys) != 2 || len(picks) != 2 {
		t.Fatalf("keys=%v picks=%v", keys, picks)
	}
	for k, vals := range keys {
		if !contains(vals, picks[k]) {
			t.Fatalf("pick %s=%s not in pool %v", k, picks[k], vals)
		}
	}
	// Exactly one value per key is visible; unconditional units always are.
	if p.Unit("m/always").Hidden {
		t.Fatal("unconditional unit hidden")
	}
	for _, pair := range [][2]string{{"m/forest", "m/meadow"}, {"m/night", "m/day"}} {
		a, b := p.Unit(pair[0]).Hidden, p.Unit(pair[1]).Hidden
		if a == b {
			t.Fatalf("%v: hidden=%v,%v - want exactly one hidden", pair, a, b)
		}
	}
	statuses := eng.UnitStatuses()
	hiddenCount := 0
	for _, st := range statuses {
		if st == "hidden" {
			hiddenCount++
		}
	}
	if hiddenCount != 2 {
		t.Fatalf("unit statuses: %v", statuses)
	}

	// The draw is in progress.json and a fresh engine on the same state
	// keeps it (no re-roll on daemon restart).
	var persisted map[string]string
	eng.Store.View(func(d *state.Data) { persisted = d.Variants })
	if persisted["scene"] != picks["scene"] || persisted["time"] != picks["time"] {
		t.Fatalf("persisted %v, picks %v", persisted, picks)
	}
	eng.Shutdown()
	eng2 := openEngine(t, p, stateDir)
	_, picks2 := eng2.Variants()
	if picks2["scene"] != picks["scene"] || picks2["time"] != picks["time"] {
		t.Fatalf("restart re-drew: %v -> %v", picks, picks2)
	}

	// A stale pick (value removed from the content) is re-drawn.
	eng2.Shutdown()
	st, _ := state.Open(stateDir, p.ID)
	_ = st.Update(func(d *state.Data) { d.Variants["scene"] = "gone" })
	eng3 := openEngine(t, p, stateDir)
	if _, picks3 := eng3.Variants(); !contains(keys["scene"], picks3["scene"]) {
		t.Fatalf("stale pick kept: %v", picks3)
	}
}

func TestSelectVariantSwitchesTheDraw(t *testing.T) {
	p, stateDir := variantPath(t)
	eng := openEngine(t, p, stateDir)
	_, picks := eng.Variants()
	other := "forest"
	if picks["scene"] == "forest" {
		other = "meadow"
	}
	if err := eng.SelectVariant("scene", other); err != nil {
		t.Fatal(err)
	}
	if p.Unit("m/"+other).Hidden || !p.Unit("m/"+picks["scene"]).Hidden {
		t.Fatalf("switch not applied: %s hidden=%v", other, p.Unit("m/"+other).Hidden)
	}
	_, picks2 := eng.Variants()
	if picks2["scene"] != other || picks2["time"] != picks["time"] {
		t.Fatalf("picks after switch: %v", picks2)
	}
	if err := eng.SelectVariant("scene", "desert"); err == nil {
		t.Fatal("unknown value accepted")
	}
	if err := eng.SelectVariant("weather", "rain"); err == nil {
		t.Fatal("unknown key accepted")
	}
}

func TestHiddenUnitsStayActivatable(t *testing.T) {
	// Authoring tools (shellgym solve) walk hidden units too: they must
	// activate and complete like any other unit.
	p, stateDir := variantPath(t)
	eng := openEngine(t, p, stateDir)
	_, picks := eng.Variants()
	hidden := "m/forest"
	if picks["scene"] == "forest" {
		hidden = "m/meadow"
	}
	if !p.Unit(hidden).Hidden {
		t.Fatalf("%s expected hidden", hidden)
	}
	ch, unsub := eng.Bus.Subscribe()
	defer unsub()
	if err := eng.ActivateUnit(hidden); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(10 * time.Second)
	for {
		select {
		case ev := <-ch:
			if d, ok := ev.Data.(UnitEvent); ok && d.Unit == hidden && d.Status == "completed" {
				return
			}
		case <-deadline:
			t.Fatal("hidden unit did not complete")
		}
	}
}

func TestPathWithoutVariantsWritesNoDraw(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "content", "path.yaml"), "id: novar\ntitle: x\n")
	mustWrite(t, filepath.Join(dir, "content", "010.m", "010.u", "unit.md"), trivialUnit)
	p, err := content.Load(filepath.Join(dir, "content"), "ubuntu", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	eng := openEngine(t, p, filepath.Join(dir, "state"))
	keys, picks := eng.Variants()
	if len(keys) != 0 || len(picks) != 0 {
		t.Fatalf("keys=%v picks=%v", keys, picks)
	}
	if _, err := os.Stat(filepath.Join(dir, "state", "novar", "progress.json")); err == nil {
		t.Fatal("progress.json written for a path without variants (constructor must stay side-effect free)")
	}
}
