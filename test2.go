package main

import "flag"
import "fmt"
import "os"
import "os/exec"
import "path"
import "path/filepath"
import "runtime"
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
}

/////////////////// Execute local /////////////////// 
func execLocal(params []string) {
	_, filename, _, _ := runtime.Caller(1)
	params[0] = path.Dir(filename) + "/" + params[0]
	cmd := exec.Command(params[0])
	cmd.Args = params
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
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
	err = exec.Command(start_dfs).Run()
	if err != nil {
		panic(err)
	}
}

func formatHdfs() {
	exec.Command("bash", "-c", "rm -rf /data/*/cmccabe/data*/*").Run()
	exec.Command("bash", "-c", "rm -rf /data/*/cmccabe/name*/*").Run()
	err := exec.Command("bash", "-c", "yes Y | " +
		"/home/cmccabe/h/bin/hadoop namenode -format").Run()
	if (err != nil) {
		panic(err)
	}
}

func waitForSafeModeOff() {
	const MAX_RETRIES = 1000
	for i := 0; i < MAX_RETRIES; i++ {
		output, err := exec.Command("/home/cmccabe/h/bin/hadoop", "dfsadmin",
			"-safemode", "get").Output()
		if (err != nil) {
			panic(err)
		}
		outputStr := string(output)
		if (strings.Contains(outputStr, "OFF")) {
			return
		}
		d, _ := time.ParseDuration("30s")
		time.Sleep(d)
	}
}

func (c *Config) startCluster() {
	fmt.Println("** shutting down Java...")
	shutdownJava()
	fmt.Println("** re-arranging symlinks, checking out code, setting readahead...")
	os.Remove(HADOOP_HOME_BASE)
	err := os.Symlink(c.hadoop, HADOOP_HOME_BASE)
	if err != nil {
		panic(err)
	}
	execLocal([]string { "setReadahead", strconv.FormatInt(c.readahead, 10) })
	checkoutConfig(c.confBranch)
	fmt.Println("** starting cluster for " + c.toString() + "...")
	if (c.shouldReformat) {
		fmt.Println("** reformatting...")
		formatHdfs()
	}
	dfsStart()
	waitForSafeModeOff()
	fmt.Println("** cluster started for " + c.toString())
	execLocal([]string { "dropCache" })
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
	curArgs := append(args, testRun.toString())
	curArgs = append(curArgs, testRun.directory)
	cmd := exec.Command(curArgs[0])
	cmd.Args = curArgs
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

var skipClusterSetup = flag.Bool("skipClusterSetup", false, "whether to skip cluster setup (useful only for trivial testing)")

var nonceDir = flag.String("nonce", "RANDOM", "the directory to put test outputs into.")

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
	var err error
	n.directory, err = filepath.Abs(baseDir)
	err = os.Mkdir("./" + baseDir, 0755)
	if err != nil {
		fmt.Println("** failed to create nonce directory " + baseDir)
		return err
	}
	return nil
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
	for i := 0; i < len(CONFIGS); i++ {
		var testRun TestRun
		testRun.init(&CONFIGS[i], &nonce)
		if (!*skipClusterSetup) {
			testRun.startCluster()
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
