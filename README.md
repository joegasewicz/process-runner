# Process Runner
Runs many processes

### Config
Create a `prconfig.yaml` file. Name a process & add the relevant commands and arguments.
```yaml
processes:
  sass:
    directory: examples/frontend
    command: sass
    args: ["sass/index.scss", "css/index.css", "--watch"]
  bash:
    directory:  examples/bash
    command: ./test.sh
  
```
### `prconfig` Config
- **processes**
    
    Top level of the processes config file
- **[name]**
    
    This is the process name you want to assign (it can be any string)
- **directory**
    
    The relative directory the process will be run from
- **command**

    The command that will be run eg. go, ./<exc>, make ... .etc
- **args**
    
    The command arguments as an array e.g `["-t", "-o"]`

### Run multiple MUX Go Servers
```yaml
processes:
  server_1:
    directory: examples/servers/server_1
    command: go
    args: ["run", "serverOne.go"]
  server_2:
    directory: examples/servers/server_2
    command: go
    args: ["run", "serverTwo.go"]
  ... etc.
```

### Example Output
This example logs to std output a Go's mux server's logs / std output

![Process Runner Log Example](imgs/log_1.png?raw=true)