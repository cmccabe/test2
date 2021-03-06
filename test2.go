package main

import "flag"
import "fmt"
import "os"
import "os/exec"
import "path"
import "path/filepath"
import "strconv"
import "strings"
import "time"

// Configuration directory
const CONF_DIR = "/home/cmccabe/bench-conf"

// Where current hadoop is installed
const HADOOP_HOME_BASE = "/home/cmccabe/h"

type Config struct {
	// Whether we should reformat the HDFS install prior to running this 
	shouldReformat bool

	// Hadoop install directory
	hadoop string

	// Readahead setting
	readahead int64

	// Configuration branch to use
	confBranch string
}

var CONFIGS = []Config{
	// hard disk configs
	Config{
		shouldReformat:true,
		hadoop:"/home/cmccabe/cdh4",
		readahead:1048576,
		confBranch:"f_c_L_1mRA",
	},
	Config{
		shouldReformat:false,
		hadoop:"/home/cmccabe/cdh4",
		readahead:1048576,
		confBranch:"f_C_L_1mRA",
	},
	Config{
		shouldReformat:false,
		hadoop:"/home/cmccabe/cdh4",
		readahead:8388608,
		confBranch:"f_c_L_8mRA",
	},
	Config{
		shouldReformat:false,
		hadoop:"/home/cmccabe/cdh4",
		readahead:8388608,
		confBranch:"f_C_L_8mRA",
	},
	Config{
		shouldReformat:true,
		hadoop:"/home/cmccabe/cdh3",
		readahead:1048576,
		confBranch:"f_c_L_1mRA",
	},
	Config{
		shouldReformat:false,
		hadoop:"/home/cmccabe/cdh3",
		readahead:8388608,
		confBranch:"f_c_L_8mRA",
	},
	// fusion I/O configs
	Config{
		shouldReformat:true,
		hadoop:"/home/cmccabe/cdh4",
		readahead:1048576,
		confBranch:"F_c_L_1mRA",
	},
	Config{
		shouldReformat:true,
		hadoop:"/home/cmccabe/cdh4",
		readahead:1048576,
		confBranch:"F_C_L_1mRA",
	},
	Config{
		shouldReformat:true,
		hadoop:"/home/cmccabe/cdh4",
		readahead:8388608,
		confBranch:"F_c_L_8mRA",
	},
	Config{
		shouldReformat:true,
		hadoop:"/home/cmccabe/cdh4",
		readahead:8388608,
		confBranch:"F_C_L_8mRA",
	},
	Config{
		shouldReformat:true,
		hadoop:"/home/cmccabe/cdh3",
		readahead:1048576,
		confBranch:"F_c_L_1mRA",
	},
	Config{
		shouldReformat:true,
		hadoop:"/home/cmccabe/cdh3",
		readahead:8388608,
		confBranch:"F_c_L_8mRA",
	},
	// hard disk configs
//	Config{
//		shouldReformat:true,
//		hadoop:"/home/cmccabe/cdh4",
//		readahead:256,
//		confBranch:"f_c_L",
//	},
//	Config{
//		shouldReformat:false,
//		hadoop:"/home/cmccabe/cdh4",
//		readahead:256,
//		confBranch:"f_C_L",
//	},
//	Config{
//		shouldReformat:true,
//		hadoop:"/home/cmccabe/cdh3",
//		readahead:256,
//		confBranch:"f_c_L",
//	},
//	Config{
//		shouldReformat:false,
//		hadoop:"/home/cmccabe/cdh3",
//		readahead:256,
//		confBranch:"f_C_L",
//	},
//	// fusion I/O configs
//	Config{
//		shouldReformat:true,
//		hadoop:"/home/cmccabe/cdh4",
//		readahead:256,
//		confBranch:"F_c_L",
//	},
//	Config{
//		shouldReformat:false,
//		hadoop:"/home/cmccabe/cdh4",
//		readahead:256,
//		confBranch:"F_C_L",
//	},
//	Config{
//		shouldReformat:true,
//		hadoop:"/home/cmccabe/cdh3",
//		readahead:256,
//		confBranch:"F_c_L",
//	},
//	Config{
//		shouldReformat:false,
//		hadoop:"/home/cmccabe/cdh3",
//		readahead:256,
//		confBranch:"F_C_L",
//	},
}

/////////////////// Configuration management functions /////////////////// 
func checkoutConfig(branch string) {
	// TODO: verify that we're not creating the branch here
	err := os.Chdir(CONF_DIR)
	if err != nil {
		panic(err)
	}
	cmd := exec.Command("git", "checkout", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}

func javaLives() bool {
	err := exec.Command("ps", "-o", "pid", "-C", "java").Run()
	return err == nil
}

func shutdownJava() {
	for killall := false; javaLives(); killall = true {
		exec.Command("killall", "-q", "java")
		if (killall) {
			exec.Command("killall", "-9", "-q", "java").Run()
		}
	}
}

func dfsStart() {
	err := os.Chdir(HADOOP_HOME_BASE)
	if err != nil {
		panic(err)
	}
	start_dfs := "./sbin/start-dfs.sh"
	_, err = os.Stat(start_dfs)
	if (err != nil) {
		start_dfs = "./bin/start-dfs.sh"
	}
	startDfs := NewSubprocess([]string { start_dfs }, false, 30)
	startDfs.PrintOutputOnSuccess = false
	startDfs.Run()
}

func getDirectories(confKey string) ([]string, error) {
	envName := "OVERRIDE_" + strings.Replace(confKey, ".", "_", -1)
	out := os.Getenv(envName)
	if (len(out) == 0) {
		fmt.Printf(envName + " not found.  checking getconf.\n")
		cmd := exec.Command(HADOOP_HOME_BASE + "/bin/hdfs")
		cmd.Args = []string { HADOOP_HOME_BASE + "/bin/hdfs", "getconf",
			"-confKey", confKey }
		bytes, err := cmd.Output()
		if (err != nil) {
			return nil, err
		}
		out = string(bytes)
	} else {
		fmt.Printf(envName + " found as " + out + "\n")
	}
	// Strip off everything but the last line (i.e. everything before the last
	// newline).  This is to get rid of log4j spew.
	// Technically directory names could contain newlines.  But we refuse to
	// think about that.
	out = strings.TrimRight(out, "\n")
	lastNewline := strings.LastIndex(out, "\n")
	if (lastNewline > -1) {
		out = out[lastNewline+1:]
	}
	dirs := strings.Split(out, ",")
	return dirs, nil
}

func cleanDirectories(dirs []string) {
	for i := 0; i < len(dirs); i++ {
		fmt.Printf("removing directory %s\n", dirs[i])
		err := os.RemoveAll(dirs[i])
		if (err != nil) {
			fmt.Printf("error removing %s: %v\n", dirs[i], err)
		}
	}
}

func formatHdfs() {
	dirs, err := getDirectories("dfs.name.dir")
	if (err != nil) {
		fmt.Printf("error getting dfs.name.dir directories: %v\n", err)

	} else {
		cleanDirectories(dirs)
	}
	dirs, err = getDirectories("dfs.data.dir")
	if (err != nil) {
		fmt.Printf("error getting dfs.data.dir directories: %v\n", err)
	} else {
		cleanDirectories(dirs)
	}
	fmt := NewSubprocess([]string { "bash", "-c", "yes Y | " +
		"/home/cmccabe/h/bin/hadoop namenode -format" }, false, 100)
	fmt.PrintOutputOnSuccess = false
	fmt.Run()
}

func waitForSafeModeOff() {
	err := os.Chdir(HADOOP_HOME_BASE)
	if err != nil {
		panic(err)
	}
	subProc := NewSubprocess([]string { "./bin/hadoop", "dfsadmin",
			"-safemode", "get" }, false, 100)
	subProc.PrintOutputOnSuccess = false
	for ;; {
		subProc.Run()
		if (strings.Contains(subProc.CombinedOutput, "OFF")) {
			return
		}
		d, _ := time.ParseDuration("30s")
		time.Sleep(d)
	}
	panic("safe mode didn't turn off!")
}

func (c *Config) startCluster(clusterSetup string) {
	fmt.Println("** shutting down Java...")
	shutdownJava()
	fmt.Println("** re-arranging symlinks, checking out code, setting readahead...")
	os.Remove(HADOOP_HOME_BASE)
	err := os.Symlink(c.hadoop, HADOOP_HOME_BASE)
	if err != nil {
		panic(err)
	}
	NewSubprocess([]string { "setReadahead",
		strconv.FormatInt(c.readahead, 10) }, true, 1)
	checkoutConfig(c.confBranch)
	fmt.Println("** starting cluster for " + c.toString() + "...")
	if (c.shouldReformat || clusterSetup == "always") {
		fmt.Println("** reformatting...")
		formatHdfs()
	}
	dfsStart()
	waitForSafeModeOff()
	fmt.Println("** cluster started for " + c.toString())
	NewSubprocess([]string { "dopCache" }, true, 1)
	fmt.Println("** page cache dropped.")
}

func (c *Config) toString() string {
	return fmt.Sprintf("%s__%d__%s",
			path.Base(c.hadoop), c.readahead, c.confBranch)
}

type TestRun struct {
	*Config
	directory string
}

func (testRun *TestRun) init(c *Config, n *Nonce) error {
	testRun.Config = c
	testRun.directory = n.directory + "/" + c.toString()
	err := os.Mkdir(testRun.directory, 0755)
	if err != nil {
		fmt.Println("** failed to create test run directory ", testRun.directory)
		return err
	}
	return nil
}

func (testRun *TestRun) run(args []string) error {
	cmd := exec.Command(args[0])
	cmd.Args = args
	stdoutFile, err := os.Create(testRun.directory + "/stdout"); if err != nil {
		panic(err)
	}
	var stderrFile *os.File
	stderrFile, err = os.Create(testRun.directory + "/stderr"); if err != nil {
		panic(err)
	}
	// TODO: use bufio here?
	cmd.Stdout = stdoutFile
	defer stdoutFile.Close()
	cmd.Stderr = stderrFile
	defer stderrFile.Close()
	err = cmd.Run()
	return err
}

/////////////////// Main /////////////////// 
var ignoreFailure = flag.Bool("ignoreFailure", false, "whether to ignore the failure of tests and keep going")

var clusterSetup = flag.String("clusterSetup", "sometimes", "when to perform cluster setup (sometimes, always, never).")

var nonceDir = flag.String("nonce", "RANDOM", "the directory to put test outputs into.")

var startConfig = flag.String("startConfig", "", "skip all configurations up to this configuration.")

type Nonce struct {
	directory string
}

func (n *Nonce) init(dir string) error {
	baseDir := ""
	if (dir == "RANDOM") {
		baseDir = strconv.Itoa(os.Getpid())
	} else {
		baseDir = dir
	}
	n.directory, _ = filepath.Abs("./" + baseDir)
	if _, err := os.Stat(n.directory); err != nil {
		err = os.Mkdir(n.directory, 0755)
		if err != nil {
			fmt.Println("** failed to create nonce directory " + baseDir)
			return err
		}
	}
	return nil
}

func getStartConfigIdx(startConfig string) int {
	if (startConfig == "") {
		return 0
	}
	i := 0
	for i = 0; i < len(CONFIGS); i++ {
		if (CONFIGS[i].toString() == startConfig) {
			return i;
		}
	}
	fmt.Printf("There were no configurations matching %s\n", startConfig)
	fmt.Printf("valid configurations are: ");
	for j := 0; j < len(CONFIGS); j++ {
		fmt.Printf("\n%s", CONFIGS[j].toString())
	}
	fmt.Printf("\n")
	os.Exit(1)
	return -1
}

func main() {
	flag.Parse()
	args := flag.Args()
	nonce := Nonce{}
	err := nonce.init(*nonceDir); if err != nil {
		panic(err)
	}
	if (len(args) < 1) {
		fmt.Println("you must give at least one test command to run (example: echo)")
		os.Exit(1)
	}

	for i := getStartConfigIdx(*startConfig); i < len(CONFIGS); i++ {
		var testRun TestRun
		testRun.init(&CONFIGS[i], &nonce)
		if (*clusterSetup != "never") {
			testRun.startCluster(*clusterSetup)
		}
		fmt.Println("** running test " + args[0] + "...")
		err = testRun.run(args)
		if err != nil {
			fmt.Println("** test failed: " + err.Error())
			if (!*ignoreFailure) {
				panic(err)
			}
		} else {
			fmt.Println("** test succeeded.")
		}
	}
}
