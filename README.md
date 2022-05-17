# Process Runner
Runs many processes

## Config
Create a `prconfig.yaml`
```yaml
processes:
  go1:
    directory: examples/go_routines
    command: ./go_routines
  go2:
    directory: examples/go_routines
    command: ./go_routines
  sass:
    directory: examples/frontend
    command: sass
    args: ["sass/index.scss", "css/index.css"]
```