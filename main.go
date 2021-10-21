// Swagrag is an extremely simple **swag**ge**r ag**gregator. It was built to
// meet a single need; to combine multiple yaml swagger files with varying
// server definitions into a single, [oats](https://github.com/influxdata/oats/)
// consumable one. If your needs differ, do not use this.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// strSlc is needed for accepting a slice of strings as a flag.
type strSlc []string

func (s *strSlc) String() string {
	return strings.Join(*s, ",")
}

func (s *strSlc) Set(val string) error {
	*s = append(*s, val)
	return nil
}

type config struct {
	files       strSlc
	oapiVersion string
	apiTitle    string
	apiVersion  string
}

// Info defines the 'info' section of an openapi document.
type Info struct {
	Title       string `yaml:"title,omitempty"`       // Title represents the title of the api being defined.
	Version     string `yaml:"version,omitempty"`     // Version represents the version of the defined api.
	Description string `yaml:"description,omitempty"` // Description describes the api.
}

// Servers defines the 'servers' section of an openapi document.
type Servers struct {
	Description string `yaml:"description,omitempty"` // Description optionally describes the host at URL.
	URL         string `yaml:"url"`                   // URL is required and defines the url to the target host.
}

// Components defines the 'components' section of an openapi document.
type Components struct {
	Parameters map[string]interface{} `yaml:"parameters,omitempty"` // Parameters define request parameters.
	Schemas    map[string]interface{} `yaml:"schemas,omitempty"`    // Schemas define re-usable data types.
	Responses  map[string]interface{} `yaml:"responses,omitempty"`  // Responses define server responses.
}

// Swagger defines the complete swagger definition.
type Swagger struct {
	Openapi    string                 `yaml:"openapi,omitempty"`
	Info       Info                   `yaml:"info,omitempty"`
	Servers    []Servers              `yaml:"servers,omitempty"`
	Paths      map[string]interface{} `yaml:"paths,omitempty"`
	Components Components             `yaml:"components,omitempty"`
}

var cfg config

func init() {
	flag.Var(&cfg.files, "file", "location of yaml swagger file (comma separated or define multiple times)")
	flag.StringVar(&cfg.oapiVersion, "openapi-version", "3.0.0", "version of openapi to print in output")
	flag.StringVar(&cfg.apiTitle, "api-title", "", "api title to print in output")
	flag.StringVar(&cfg.apiVersion, "api-version", "", "api version to print in output")
}

func main() {
	flag.Parse()

	if len(cfg.files) == 1 && strings.Contains(cfg.files[0], ",") {
		cfg.files = strings.Split(cfg.files[0], ",")
	}

	if len(cfg.files) < 2 {
		fmt.Fprintln(os.Stderr, "at least two files must be specified")
		flag.Usage()
		os.Exit(1)
	}

	paths := map[string]interface{}{}
	components := Components{
		Parameters: map[string]interface{}{},
		Schemas:    map[string]interface{}{},
		Responses:  map[string]interface{}{},
	}

	for _, file := range cfg.files {
		swagger := &Swagger{}
		dat, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read %s - %s\n", file, err.Error())
			continue
		}

		err = yaml.Unmarshal(dat, swagger)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to unmarshal - %s\n", err.Error())
			continue
		}

		switch {
		case len(swagger.Servers) == 0:
			swagger.Servers = []Servers{{URL: ""}}
		case len(swagger.Servers) > 1:
			fmt.Fprintf(os.Stderr, "warn: multiple servers defined in %s\n", file)
		}

		for k, v := range swagger.Paths {
			if _, ok := paths[swagger.Servers[0].URL+k]; ok {
				fmt.Fprintf(os.Stderr, "warn: path already exists for %q\n", swagger.Servers[0].URL+k)
				continue
			}
			paths[swagger.Servers[0].URL+k] = v
		}

		for k, v := range swagger.Components.Parameters {
			if _, ok := components.Parameters[k]; ok {
				fmt.Fprintf(os.Stderr, "warn: parameter already exists for %q\n", k)
				continue
			}
			components.Parameters[k] = v
		}

		for k, v := range swagger.Components.Schemas {
			if _, ok := components.Schemas[k]; ok {
				fmt.Fprintf(os.Stderr, "warn: schema already exists for %q\n", k)
				continue
			}
			components.Schemas[k] = v
		}

		for k, v := range swagger.Components.Responses {
			if _, ok := components.Responses[k]; ok {
				fmt.Fprintf(os.Stderr, "warn: response already exists for %q\n", k)
				continue
			}
			components.Responses[k] = v
		}
	}

	info := Info{}
	switch {
	case cfg.apiTitle != "":
		info.Title = cfg.apiTitle
	case cfg.apiTitle == "":
		info.Title = fmt.Sprintf("Combined API from %s", cfg.files)
	case cfg.apiVersion != "":
		info.Version = cfg.apiVersion
	}

	swagger := &Swagger{
		Openapi:    cfg.oapiVersion,
		Servers:    []Servers{{URL: ""}},
		Info:       info,
		Paths:      paths,
		Components: components,
	}

	d, err := yaml.Marshal(swagger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal - %s\n", err.Error())
		return
	}

	os.Stdout.Write(d)
}
