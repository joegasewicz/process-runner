package main

import (
	"flag"
	"fmt"
	"github.com/logrusorgru/aurora/v3"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os/exec"
	"sync"
)

type Process struct {
	ID        int
	Name      string
	Directory string
	Command   string
}

type ProcessYAML struct {
	Directory string `yaml:"directory"`
	Command   string `yaml:"command"`
}

type ProcessesYAML struct {
	Processes map[string]ProcessYAML `yaml:"processes"`
}

// UnMarshalYAML takes a pointer to []Process & generates values from a `prconfig.yaml` file
func UnMarshalYAML(processes *[]Process) error {
	var err error
	yamlFile, err := ioutil.ReadFile("examples/go_routines/prconfig.yaml")
	var p ProcessesYAML
	err = yaml.Unmarshal(yamlFile, &p)
	// Add processes
	for k, v := range p.Processes {
		newProcess := Process{
			Name:      k,
			Directory: v.Directory,
			Command:   v.Command,
		}
		*processes = append(*processes, newProcess)
	}
	return err
}

func CreateProcess(_wg *sync.WaitGroup, index int, processes []Process) {
	defer _wg.Done()
	processes[index].ID = index
	testCmd := exec.Command(processes[index].Command)
	testCmd.Dir = processes[index].Directory
	out, err := testCmd.Output()
	if err != nil {
		log.Fatal(err)
		return
	}
	// logging TODO use templates
	fmt.Printf(
		"Process: (%d)\n\tname: %s\n\tdirectory: %s\n\toutput: %s",
		aurora.Blue(processes[index].ID),
		aurora.Blue(processes[index].Name),
		aurora.Blue(processes[index].Directory),
		aurora.Blue(out),
	)
}

func main() {
	var err error
	var dir string
	var cmd string
	var processes []Process
	var wg sync.WaitGroup
	var hasQuit bool
	var userOption string

	flag.StringVar(&dir, "dir", "", "directory to run the process from")
	flag.StringVar(&cmd, "cmd", "", "command to run")
	flag.Parse()

	err = UnMarshalYAML(&processes)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	// Run processes
	wg.Add(len(processes))
	for i := 0; i < len(processes); i++ {
		go CreateProcess(&wg, i, processes)
	}

	// Wait for user inputs
	for !hasQuit {
		fmt.Println("Ready:")
		fmt.Scanln(&userOption)
		fmt.Println("Entered: ", userOption)

	}
	wg.Wait()

	// If we reach this point it means there are no running processes left
	log.Printf("All running processes complete\n")
}
