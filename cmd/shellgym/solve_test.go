package main

import "testing"

func TestExpandVars(t *testing.T) {
	vars := map[string]string{"PAUSE": "2", "PAUSE_MAX": "9", "BOGUS": "carrots"}
	cases := map[string]string{
		"sleep $PAUSE && hostname":         "sleep 2 && hostname",
		"sleep ${PAUSE}s; echo $PAUSE_MAX": "sleep 2s; echo 9",
		"date --$BOGUS || whoami":          "date --carrots || whoami",
		"echo $PAUSEX $? $HOME":            "echo $PAUSEX $? $HOME",
		"#!type sleep $PAUSE":              "#!type sleep 2",
	}
	for in, want := range cases {
		if got := expandVars(in, vars); got != want {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
}
