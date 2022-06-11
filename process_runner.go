package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/logrusorgru/aurora/v3"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os/exec"
	"sync"
	"time"
)

const (
	configFileName = "prconfig.yaml"
)

type Process struct {
	ID        int
	Name      string
	Directory string
	Command   string
	Args      []string
}

type ProcessYAML struct {
	Directory string   `yaml:"directory"`
	Command   string   `yaml:"command"`
	Args      []string `yaml:"args"`
	Port      int8     `yaml:"port"`
}

type ProcessesYAML struct {
	Processes map[string]ProcessYAML `yaml:"processes"`
}

type ProcessOutput struct {
	Process *Process
	Out     string
	Index   int
	Error   string
}

// UnMarshalYAML takes a pointer to []Process & generates values from a `prconfig.yaml` file
func UnMarshalYAML(processes *[]Process, dir string) error {
	var err error
	configFile := fmt.Sprintf("%s/%s", dir, configFileName)
	yamlFile, err := ioutil.ReadFile(configFile)
	var p ProcessesYAML
	err = yaml.Unmarshal(yamlFile, &p)
	// Add processes
	for k, v := range p.Processes {
		newProcess := Process{
			Name:      k,
			Directory: v.Directory,
			Command:   v.Command,
			Args:      v.Args,
		}
		*processes = append(*processes, newProcess)
	}
	return err
}

func logProcessStdOut(p Process, out string, err string) {
	fmt.Printf(
		"[%s] Process: (%d)\n\tname: %s\n\tdirectory: %s\n\toutput: %s\n\terror: %s\n",
		time.Now().Format("15:01:05"),
		aurora.Blue(p.ID),
		aurora.Blue(p.Name),
		aurora.Blue(p.Directory),
		aurora.Blue(out),
		aurora.Blue(err),
	)
}

// CreateProcess creates a single process
func CreateProcess(_wg *sync.WaitGroup, index int, processes []Process, dir string, logChan chan ProcessOutput) {
	var out string
	var stderrMsg string
	var err error
	defer _wg.Done()
	processes[index].ID = index
	// Command
	cmd := exec.Command(processes[index].Command, processes[index].Args...)
	if dir != "" {
		cmd.Dir = processes[index].Directory
	}
	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()

	err = cmd.Start()

	errScanner := bufio.NewScanner(stdout)
	errScanner.Split(bufio.ScanWords)
	stdoutScanner := bufio.NewScanner(stderr)
	stdoutScanner.Split(bufio.ScanWords)
	for stdoutScanner.Scan() {
		m := stdoutScanner.Text()
		out += m
	}
	for errScanner.Scan() {
		e := errScanner.Text()
		stderrMsg += e

	}
	// Output
	logChan <- ProcessOutput{
		Process: &processes[index],
		Out:     out,
		Index:   index,
		Error:   stderrMsg,
	}
	err = cmd.Wait()
	// Errors
	if err != nil {
		if err.Error() == "exit status 127" {
			errMsg := fmt.Sprintf("Unknown command: %s Is this file executable?\n", processes[index].Command)
			log.Fatal(errMsg)
		}
		log.Fatal(err)
		return
	}

}

func main() {
	var err error
	var dir string
	var processes []Process
	var wg sync.WaitGroup
	//var hasQuit bool
	//var userOption string
	var logChan chan ProcessOutput = make(chan ProcessOutput)
	// Flags
	flag.StringVar(&dir, "dir", "", "directory to run the process from")
	flag.Parse()
	// Marshal yaml
	err = UnMarshalYAML(&processes, dir)
	if err != nil {
		log.Fatal(err)
	}
	// Run processes
	wg.Add(len(processes))
	for i := 0; i < len(processes); i++ {
		go CreateProcess(&wg, i, processes, dir, logChan)
	}

	go func() {
		for {
			select {
			case l := <-logChan:
				logProcessStdOut(*l.Process, l.Out, l.Error)
			}
		}
	}()
	// Wait for user inputs
	//for !hasQuit {
	//	fmt.Println("Ready:")
	//	fmt.Scanln(&userOption)
	//	fmt.Println("Entered: ", userOption)
	//	time.Sleep(100 * time.Millisecond)
	//}
	wg.Wait()
	// If we reach this point it means there are no running processes left
	log.Printf("\n[%s] All running processes started...\n", time.Now().Format("15:01:05"))
}
