package main

import "os"
import "syscall"

func clearPageCache() {
	syscall.Sync()
	fo, err := os.OpenFile("/proc/sys/vm/drop_caches", syscall.O_WRONLY, 0666)
	if err != nil { panic(err) }
	defer fo.Close()
	if n2, err := fo.Write([]byte{ 0x33 }); err != nil {
		panic(err)
	} else if n2 != 1 {
		panic("error writing")
	}
}

func main() {
	clearPageCache()
}
