package main

import "io/ioutil"
import "fmt"
import "os"
import "os/exec"
import "time"

const TEMP_SIZE = 858993459200

func timedCommand(args []string) time.Duration {
	before := time.Now()
	cmd := exec.Command(args[0])
	//cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Args = args
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
	after := time.Now()
	duration := after.Sub(before)
	return duration
}

func toMegabytesPerSecond(bytesPerSec float64) float64 {
	return bytesPerSec / (1024 * 1024)
}

func main() {
	tempFile, err := ioutil.TempFile("", "hdfsCat")
	if err != nil {
		panic(err)
	}
	os.Truncate(tempFile.Name(), TEMP_SIZE)
	fmt.Println("*** inputting file to hdfs ***")
	timedCommand([]string {"/home/cmccabe/test2/dropCache"})
	inputDur := timedCommand([]string {"/home/cmccabe/h/bin/hadoop", "fs",
		"-copyFromLocal", tempFile.Name(), "/t"})
	fmt.Println("*** finished in " + inputDur.String() + " ***")
	timedCommand([]string {"/home/cmccabe/test2/dropCache"})
	outputDur := timedCommand([]string {"/home/cmccabe/h/bin/hadoop", "fs",
		"-cat", "/t"})

	var inputRate float64 = TEMP_SIZE
	inputRate /= inputDur.Seconds()

	var outputRate float64 = TEMP_SIZE
	outputRate /= outputDur.Seconds()

	fmt.Printf("*** input: %f MB/s\n", toMegabytesPerSecond(inputRate))
	fmt.Printf("*** output: %f MB/s\n", toMegabytesPerSecond(outputRate))
	// remove this so it doesn't disrupt the next test
	timedCommand([]string {"/home/cmccabe/h/bin/hadoop", "fs", "-rm", "/t"})
}
