package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/logrusorgru/aurora/v3"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

const (
	configFileName = "prconfig.yaml"
)

const (
	STATE_STARTING  = "starting"
	STATE_RUNNING   = "running"
	STATE_COMPLETED = "completed"
)

type Process struct {
	ID        int
	Name      string
	Directory string
	Command   string
	Args      []string
	State     string
	Env       map[string]string
}

type ProcessYAML struct {
	Directory string            `yaml:"directory"`
	Command   string            `yaml:"command"`
	Args      []string          `yaml:"args"`
	Port      int8              `yaml:"port"`
	Env       map[string]string `yaml:"env"`
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

// unMarshalYAML takes a pointer to []Process & generates values from a `prconfig.yaml` file
func unMarshalYAML(processes *[]Process, dir string) error {
	var err error
	configFile := fmt.Sprintf("%s/%s", dir, configFileName)
	yamlFile, err := ioutil.ReadFile(configFile)
	if string(yamlFile) == "" {
		panic("No prconfig.yaml found")
	}

	var p ProcessesYAML
	err = yaml.Unmarshal(yamlFile, &p)
	// Add processes
	for k, v := range p.Processes {
		newProcess := Process{
			Name:      k,
			Directory: v.Directory,
			Command:   v.Command,
			Args:      v.Args,
			Env:       v.Env,
		}
		*processes = append(*processes, newProcess)
	}
	return err
}

func logProcessStdOut(p Process, out string, err string) {
	fmt.Printf(
		"[%s] Process: (%d)\n\t   name: %s\n\t   directory: %s\n\t   output: %s\n\t   error: %s\n\t   state: %s\n",
		aurora.Green(time.Now().Format("15:01:05")),
		aurora.Blue(p.ID),
		aurora.Blue(p.Name),
		aurora.Blue(p.Directory),
		aurora.Blue(out),
		aurora.Red(err),
		aurora.Blue(p.State),
	)
}

func sendStdOutToChannel(c chan ProcessOutput, p *Process, i int, o, e string) {
	c <- ProcessOutput{
		Process: p,
		Out:     o,
		Index:   i,
		Error:   e,
	}
}

func createProcess(_wg *sync.WaitGroup, index int, processes []Process, dir string, logChan chan ProcessOutput, ctx context.Context) {
	var (
		out       string
		stderrMsg string
		err       error
		done      = make(chan struct{})
	)
	defer _wg.Done()
	processes[index].ID = index
	processes[index].State = STATE_STARTING
	sendStdOutToChannel(logChan, &processes[index], index, out, stderrMsg)
	// Command
	cmd := exec.CommandContext(ctx, processes[index].Command, processes[index].Args...)
	// Env
	envVars := os.Environ()
	for k, v := range processes[index].Env {
		envVars = append(envVars, fmt.Sprintf("%s=%v", k, v))
	}
	cmd.Env = envVars
	// Set dir
	if dir != "" {
		cmd.Dir = processes[index].Directory
	}
	// Output
	processes[index].State = STATE_RUNNING
	sendStdOutToChannel(logChan, &processes[index], index, out, stderrMsg)
	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()
	// stdout
	stdoutScanner := bufio.NewScanner(stdout)
	stdoutScanner.Split(bufio.ScanLines)
	// stderr
	stderrScanner := bufio.NewScanner(stderr)
	stderrScanner.Split(bufio.ScanLines)
	// Update log chan
	go func() {
		for stdoutScanner.Scan() {
			out += stdoutScanner.Text()
			if out != "" {
				sendStdOutToChannel(logChan, &processes[index], index, out, stderrMsg)
				out = ""
			}
		}
		done <- struct{}{}
	}()
	go func() {
		for stderrScanner.Scan() {
			stderrMsg += stderrScanner.Text()
			if stderrMsg != "" {
				sendStdOutToChannel(logChan, &processes[index], index, out, stderrMsg)
				stderrMsg = ""
			}
		}
		done <- struct{}{}
	}()
	// Start process command
	err = cmd.Start()
	<-done
	cmd.Wait()
	// Errors
	if err != nil {
		if err.Error() == "exit status 127" {
			errMsg := fmt.Sprintf(
				"[process-runner] Error: Unknown command: %s Is this file executable?\n",
				processes[index].Command,
			)
			log.Fatal(errMsg)
		}
		log.Fatalf(
			"[process-runner] Error for process %s\n[process-runner] Full Error: %s",
			processes[index].Name, err.Error())
		return
	}
}

func main() {
	var err error
	var dir string
	var processes []Process
	var wg sync.WaitGroup
	var ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	var logChan chan ProcessOutput = make(chan ProcessOutput)
	defer cancel()
	// Flags
	flag.StringVar(&dir, "dir", "", "relative path to directory that contains a prconfig.yaml file")
	flag.Parse()
	// Marshal yaml
	err = unMarshalYAML(&processes, dir)
	if err != nil {
		log.Fatal(err)
	}
	// Run processes
	wg.Add(len(processes) * 2)
	for i := 0; i < len(processes); i++ {
		go createProcess(&wg, i, processes, dir, logChan, ctx)
	}

	go func() {
		for {
			select {
			case l := <-logChan:
				logProcessStdOut(*l.Process, l.Out, l.Error)
			}
		}
	}()
	wg.Wait()
	// If we reach this point it means there are no running processes left
	log.Printf(
		"\n[%s] All running processes stopped\n",
		aurora.Green(time.Now().Format("15:01:05")),
	)
}
