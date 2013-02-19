package main

import (
	"errors"
	"fmt"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strings"
	"strconv"
	"time"
)

type Subprocess struct {
	params []string
	CombinedOutput string
	Tries int
	MaxTries int
	RetryTime time.Duration
}

func NewSubprocess(params []string, useRel bool, maxTries int) *Subprocess {
	subProc := Subprocess{}
	if (useRel) {
		_, filename, _, _ := runtime.Caller(1)
		params[0] = path.Dir(filename) + "/" + params[0]
	}
	subProc.params = params
	subProc.Tries = 0
	subProc.MaxTries = maxTries
	subProc.RetryTime, _ = time.ParseDuration("30s")
	return &subProc
}

func (subProc *Subprocess) Run() {
	var err error
	first := true
	for {
		if (subProc.Tries >= subProc.MaxTries) {
			break
		}
		if (!first) {
			time.Sleep(subProc.RetryTime)
		}
		subProc.Tries++;
		first = false
		cmd := exec.Command(subProc.params[0])
		cmd.Args = subProc.params
		var out []byte
		out, err = cmd.CombinedOutput()
		subProc.CombinedOutput = string(out)
		if err == nil {
			return
		}
		fmt.Printf("Subprocess %v failed: %s\n",
			subProc.params, subProc.CombinedOutput)
	}
	panic(errors.New(fmt.Sprintf("failed to run command %v after %d tries. " +
			"CombinedOutput: '%v'",
			subProc.params, subProc.Tries, subProc.CombinedOutput)))
}

type TimedSubprocess struct {
	*Subprocess
	User float64
	System float64
	Elapsed float64
}

func (subProc *TimedSubprocess) Run() {
	timeDataRegex, err := regexp.Compile("TIME_DATA: user=([0-9.]*), " +
		"system=([0-9.]*), elapsed=([0-9.]*), ")
	if err != nil {
		panic(err)
	}
	subProc.Subprocess.Run()
	for _, line := range strings.Split(subProc.CombinedOutput, "\n") {
		arr := timeDataRegex.FindStringSubmatch(line)
		if arr != nil {
			subProc.User, _ = strconv.ParseFloat(arr[1], 64)
			subProc.System, _ = strconv.ParseFloat(arr[2], 64)
			subProc.Elapsed, _ = strconv.ParseFloat(arr[3], 64)
		}
	}
}

func NewTimedSubprocess(params []string, useRel bool, maxTries int) *TimedSubprocess {
	subProc := TimedSubprocess{}
	timedParams := make([]string, len(params) + 3)
	timedParams[0] = "/usr/bin/time"
	timedParams[1] = "-f"
	timedParams[2] = "TIME_DATA: user=%U, system=%S, " +
		"elapsed=%e, CPU=%P, (%Xtext+%Ddata %Mmax)k, " +
		"inputs=%I, outputs=%O, (%Fmajor+%Rminor)pagefaults, swaps=%W "
	copy(timedParams[3:], params)
	subProc.Subprocess = NewSubprocess(timedParams, useRel, maxTries)
	return &subProc
}
