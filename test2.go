package main

import "exec"
import "fmt"
import "os"
import "runtime"
import "strconv"
import "syscall"

// Configuration directory
const CONF_DIR = "/home/cmccabe/conf"

// Where current hadoop is installed
const HADOOP_HOME_BASE = "/home/cmccabe/h"

struct Config {
	// Whether we should reformat the HDFS install prior to running this 
	shouldReformat bool

	// Hadoop install directory
	hadoop string

	// Readahead setting
	readahead uint64

	// Configuration branch to use
	confBranch string
}

var CONFIGS := []Config{
	Config{
		shouldReformat:false, // true
		hadoop:"/home/cmccabe/cdh4",
		readahead:1048576,
		confBranch:"f_c_L_1MBR"
	},
	Config{
		shouldReformat:false,
		hadoop:"/home/cmccabe/cdh4",
		readahead:1048576,
		confBranch:"f_C_L_1MBR"
	},
	Config{
		shouldReformat:false,
		hadoop:"/home/cmccabe/cdh4",
		readahead:8388608,
		confBranch:"f_c_L_8MBR"
	},
	Config{
		shouldReformat:false,
		hadoop:"/home/cmccabe/cdh4",
		readahead:8388608,
		confBranch:"f_C_L_8MBR"
	},

	Config{
		shouldReformat:true,
		hadoop:"/home/cmccabe/cdh3",
		readahead:1048576,
		confBranch:"f_c_L_1MBR"
	},
	Config{
		shouldReformat:false,
		hadoop:"/home/cmccabe/cdh3",
		readahead:8388608,
		confBranch:"f_c_L_8MBR"
	},
}

/////////////////// Root permissions management functions /////////////////// 
var saved_uid : int;

func takeRoot() {
	saved_uid = syscall.Getuid()
	en := syscall.Setuid(0)
	if en != 0 {
		panic(fmt.Sprintf("setuid error:" , os.Errno(en)))
	}
	runtime.LockOSThread()
}

func releaseRoot() {
	en := syscall.Setuid(saved_uid)
	if en != 0 {
		panic(fmt.Sprintf("setuid error:" , os.Errno(en)))
	}
	runtime.UnlockOSThread()
}

func usage(retval int) {
	fmt.Printf("test2: performs HDFS tests.\n");
	os.Exit(retval)
}

/////////////////// Page cache management functions /////////////////// 
// Clear the page cache before our test. 
func clearPageCache() {
	takeRoot()
	defer releaseRoot()
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

func setOsReadahead(readahead uint64) {
	takeRoot()
	defer releaseRoot()
	readahead /= 4096;
	BLOCK_DEVS = []string {
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
		"/dev/fioa",
	}
	for (int i = 0; i < len(BLOCK_DEVS); i++) {
		err := exec.Command("blockdev", "--setra",
			strconv.Itoa(readahead), BLOCK_DEVS[i]).Run()
		if err != nil {
			panic(err)
		}
	}
}

/////////////////// Configuration management functions /////////////////// 
func checkoutConfig(branch string) {
	// TODO: verify that we're not creating the branch here
	err := os.Chdir(CONF_DIR)
	if err != nil {
		panic(err)
	}
	err = exec.Command("git", "checkout", branch).Run()
	if err != nil {
		panic(err)
	}
}

func shutdownJava() {
	err := exec.Command("killall", "java").Run()
	if err != nil {
		panic(err)
	}
	err = exec.Command("killall", "-9", "java").Run()
	if err != nil {
		panic(err)
	}
}

func dfsStart() {
	err := os.Chdir(HADOOP_HOME_BASE)
	if err != nil {
		panic(err)
	}
	start_dfs = "./sbin/start-dfs.sh"
	_, err := os.Stat(start_dfs)
	if (err != nil) {
		start_dfs = "./bin/start-dfs.sh"
	}
	err := exec.Command(start_dfs).Run()
	if err != nil {
		panic(err)
	}
}

func (c *Config) startCluster() {
	shutdownJava()
	os.Remove(HADOOP_HOME_BASE)
	err := os.Symlink(c.hadoop, HADOOP_HOME_BASE)
	if err != nil {
		panic(err)
	}
	setOsReadahead(c.readahead)
	checkoutConfig(c.confBranch)
	dfsStart()
	fmt.Println("cluster started for " + c.toString())
}

func (c *Config) toString() string {
	return fmt.Sprintf("%s__%d__%s",
			path.Base(c.hadoop), c.readahead, c.confBranch)
}

/////////////////// Test Code /////////////////// 

/////////////////// Main /////////////////// 
func main() {
	clearPageCache()
	for (int i = 0; i < len(CONFIGS); i++) {
		c.startCluster()
	}
}
