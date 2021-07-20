# Scaffolder

Scaffolder is a scaffolding generator for creating new connectors. It takes in a scaffolding
directory, scheme, and some values to produce scaffolding with inserted values. All the scaffolding files
follow a standard [**Go template**](https://golang.org/pkg/text/template/) syntax with additional functions provided by
[**Sprig library**](https://github.com/Masterminds/sprig) and are populated with values provided as a command-line options.

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

### Example
To create scaffolding for **YetAnotherResource** with default scaffolding we can use the following command:
```shell
> scaffolder --group yet.another.group.com --name YetAnotherResource \
             --templates-dir templates --scheme scheme.yaml --output connector
```

For most use cases, you would not need `--short` and `--version` flags, as they are defaulted to their most common values.
However, if you would ever need to dig deeper, you can always use `scaffolder --help` command.

*__WARNING:__ do not forget to use `make generate` and `make manifest` after creating scaffolding. It will autogenerate rest of the files needed.*
