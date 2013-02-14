package main

import "fmt"
import "os"
import "os/exec"
import "strconv"

func setOsReadahead(readahead int64) {
	readahead /= 4096;
	BLOCK_DEVS := []string {
		"/dev/sda",
		"/dev/sdb",
		"/dev/sdc",
		"/dev/sdd",
		"/dev/sde",
		"/dev/sdf",
		"/dev/sdg",
		"/dev/sdh",
		"/dev/sdi",
		"/dev/sdj",
		"/dev/sdk",
		"/dev/sdl",
		"/dev/fioa" }
	for i := 0; i < len(BLOCK_DEVS); i++ {
		cmd := exec.Command("blockdev", "--setra",
			strconv.FormatInt(readahead, 10), BLOCK_DEVS[i])
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			panic(err.Error())
		}
	}
}

func main() {
	if (len(os.Args) < 2) {
		fmt.Printf("setReadahead: you must supply one argument: the amount of readahead to set\n")
		os.Exit(1)
	}
	readahead, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		fmt.Printf("setReadahead: error parsing argument: " + err.Error() + "\n")
		os.Exit(1)
	}
	setOsReadahead(readahead)
}
