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
}

type ProcessesYAML struct {
	Processes map[string]ProcessYAML `yaml:"processes"`
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

// CreateProcess creates a single process
func CreateProcess(_wg *sync.WaitGroup, index int, processes []Process, dir string) {
	defer _wg.Done()
	processes[index].ID = index
	testCmd := exec.Command(processes[index].Command, processes[index].Args...)
	if dir != "" {
		testCmd.Dir = processes[index].Directory
	}
	out, err := testCmd.Output()
	if err != nil {
		if err.Error() == "exit status 127" {
			errMsg := fmt.Sprintf("Unknown command: %s Is this file executable?", processes[index].Command)
			log.Fatal(errMsg)
		}
		log.Fatal(err)
		return
	}
	if len(out) < 1 {
		out = []byte("\n")
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
	//var cmd string
	var processes []Process
	var wg sync.WaitGroup
	var hasQuit bool
	var userOption string
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
		go CreateProcess(&wg, i, processes, dir)
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
