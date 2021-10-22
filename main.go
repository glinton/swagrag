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

	full := map[string]interface{}{}

	for _, file := range cfg.files {
		swagger := map[string]interface{}{}
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

		for k, v := range swagger {
			_, ok := full[k]
			if ok {
				fmt.Fprintf(os.Stderr, "warn: value already exists for %q\n", k)
				continue
			}
			// todo: prefix with "servers" definition so routes still match
			full[k] = v
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

	full["openapi"] = cfg.oapiVersion
	full["info"] = info
	full["servers"] = []Servers{{URL: ""}}

	d, err := yaml.Marshal(full)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal - %s\n", err.Error())
		os.Exit(1)
	}

	os.Stdout.Write(d)
}
