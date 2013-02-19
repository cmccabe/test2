package main

import (
	"fmt"
	"testing"
	"time"
)

func Test_NewTimedSubprocess(t *testing.T) {
	timedSubProc := NewTimedSubprocess([]string { "sleep", "2" }, false, 1)
	timedSubProc.Run()
	if (timedSubProc.Elapsed == 0) {
		t.Error("elapsed time was 0\n")
	}
	if (timedSubProc.Tries != 1) {
		t.Error("tries should be 1\n")
	}
}

func tryFailedNewTimedSubprocess() (timedSubProc *TimedSubprocess,
		failed bool) {
	failed = true
	defer func() { recover(); }()
	timedSubProc = NewTimedSubprocess([]string { "false" }, false, 3)
	timedSubProc.RetryTime, _ = time.ParseDuration("1s")
	timedSubProc.Run()
	failed = false
	return
}

func Test_FailedNewTimedSubprocess(t *testing.T) {
	timedSubProc, failed := tryFailedNewTimedSubprocess();
	if !failed {
		t.Error("expected failure of subprocess.\n")
	}
	if (timedSubProc.Tries != 3) {
		t.Error(fmt.Sprintf("tries should be 3, but it is %d\n",
			timedSubProc.Tries))
	}
}
