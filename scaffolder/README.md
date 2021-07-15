# Scaffolder

Scaffolder is a customizable scaffolding generator. It takes in a value file, a scaffolding
directory, and a scheme to produce scaffolding with inserted values. All the scaffolding files
follow a standard [**Go template**](https://golang.org/pkg/text/template/) syntax and are populated with
values provided as a value file or command-line options.

### Scheme
Scheme file just explains how value-populated files would be located in the output. It can hold templates, 
too. It is a simple `yaml` file:

```
scheme:
  - source: path/to/{{ .template }}/file
    destination: path/to/output/file

  - source: path/to/{{ .template }}/directory
    destination: path/to/output/directory
    recursive: true
```

### Value file 

Value file must be provided as either `.json` or `.yaml` file. It can be specified via 
`input` flag of the command line. If you provide values in command line, they must be
a valid `.json` string.

Currently, only string values are supported.

### Templating

Scaffolder uses standard [**Go template**](https://golang.org/pkg/text/template/) syntax and functionality
as well as [**Sprig library**](https://github.com/Masterminds/sprig) to provide additional
functions. All the documentation about them can be found [here](https://masterminds.github.io/sprig/).