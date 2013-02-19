package main

import "io/ioutil"
import "fmt"
import "os"

const TEMP_SIZE = 85899345920 //858993459200

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
	NewSubprocess([]string { "dropCache" }, true, 1).Run()
	copyFromLocal := NewTimedSubprocess([]string { "/home/cmccabe/h/bin/hadoop",
		"fs", "-copyFromLocal", tempFile.Name(), "/t"}, false, 30)
	copyFromLocal.Run()
	NewSubprocess([]string { "dropCache" }, true, 1).Run()
	hdfsCat := NewTimedSubprocess([]string {"/home/cmccabe/h/bin/hadoop",
		"fs", "-cat", "/t"}, false, 30)
	hdfsCat.Run()

	var inputRate float64 = TEMP_SIZE
	inputRate /= copyFromLocal.Elapsed

	var outputRate float64 = TEMP_SIZE
	outputRate /= hdfsCat.Elapsed

	fmt.Printf("*** input: %f MB/s\n", toMegabytesPerSecond(inputRate))
	fmt.Printf("*** output: %f MB/s\n", toMegabytesPerSecond(outputRate))
	// remove this so it doesn't disrupt the next test
	NewSubprocess([]string { "/home/cmccabe/h/bin/hadoop", "fs", "-rm", "/t"},
		false, 30).Run()
}
