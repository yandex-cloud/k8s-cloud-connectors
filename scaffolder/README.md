# Scaffolder

Scaffolder is a customizable scaffolding generator. It takes in a value file, a scaffolding
directory, and a scheme to produce scaffolding with inserted values. All the scaffolding files
follow a standard [**Go template**](https://golang.org/pkg/text/template/) syntax and are populated with
values provided as a value file or command-line options.

### Scheme
Scheme file just explains how value-populated files would be located in the output. It can hold templates, 
too. Its syntax is quite simple:

```
# Every line that starts with an '#' is considered a comment and therefore is not processed

input/path/to/scaffolding/template => output/path/to/result

input/path/is/relative/to/template/sirectory => output/path/is/relative/to/output/directory

{{ .This.Path.Is.Templatable }} => somewhere

somewhere => {{ .And.This.One.Is.Templatable.Also }}

{{ .Both.Paths }} => {{ .Are.Templatable }}
```

### Value file 

Value file must be provided as either `.json` or `.yaml` file. It can be specified via 
`input` flag of the command line. If you provide values in command line, they must be
a valid `.json` string.

Currently, only string values are supported.