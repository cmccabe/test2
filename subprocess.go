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
	PrintableParams []string
	CombinedOutput string
	Tries int
	MaxTries int
	RetryTime time.Duration
}

func NewSubprocess(params []string, useRel bool, maxTries int) *Subprocess {
	subp := Subprocess{}
	if (useRel) {
		_, filename, _, _ := runtime.Caller(1)
		params[0] = path.Dir(filename) + "/" + params[0]
	}
	subp.params = params
	subp.PrintableParams = params
	subp.Tries = 0
	subp.MaxTries = maxTries
	subp.RetryTime, _ = time.ParseDuration("30s")
	return &subp
}

func (subp *Subprocess) Run() {
	var err error
	first := true
	for {
		if (subp.Tries >= subp.MaxTries) {
			break
		}
		if (!first) {
			time.Sleep(subp.RetryTime)
		}
		subp.Tries++;
		first = false
		cmd := exec.Command(subp.params[0])
		cmd.Args = subp.params
		var out []byte
		out, err = cmd.CombinedOutput()
		subp.CombinedOutput = string(out)
		if err == nil {
			fmt.Printf("SUCCESS: '%v': CombinedOutput: '%v'\n",
				subp.PrintableParams, subp.CombinedOutput)
			return
		}
		fmt.Printf("FAILED: '%v': CombinedOutput: '%v', err:'%v'\n",
			subp.PrintableParams, subp.CombinedOutput, err)
	}
	panic(errors.New(fmt.Sprintf("failed to run command after %d tries.",
			subp.Tries)))
}

type TimedSubprocess struct {
	*Subprocess
	User float64
	System float64
	Elapsed float64
}

func (subp *TimedSubprocess) Run() {
	timeDataRegex, err := regexp.Compile("TIME_DATA: user=([0-9.]*), " +
		"system=([0-9.]*), elapsed=([0-9.]*), ")
	if err != nil {
		panic(err)
	}
	subp.Subprocess.Run()
	for _, line := range strings.Split(subp.CombinedOutput, "\n") {
		arr := timeDataRegex.FindStringSubmatch(line)
		if arr != nil {
			subp.User, _ = strconv.ParseFloat(arr[1], 64)
			subp.System, _ = strconv.ParseFloat(arr[2], 64)
			subp.Elapsed, _ = strconv.ParseFloat(arr[3], 64)
		}
	}
}

func NewTimedSubprocess(params []string, useRel bool, maxTries int) *TimedSubprocess {
	subp := TimedSubprocess{}
	timedParams := make([]string, len(params) + 3)
	timedParams[0] = "/usr/bin/time"
	timedParams[1] = "-f"
	timedParams[2] = "TIME_DATA: user=%U, system=%S, " +
		"elapsed=%e, CPU=%P, (%Xtext+%Ddata %Mmax)k, " +
		"inputs=%I, outputs=%O, (%Fmajor+%Rminor)pagefaults, swaps=%W "
	copy(timedParams[3:], params)
	subp.Subprocess = NewSubprocess(timedParams, useRel, maxTries)
	subp.Subprocess.PrintableParams = timedParams[3:]
	return &subp
}
