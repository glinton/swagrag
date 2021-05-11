# Swagrag

Swagrag is an extremely simple **swag**ge**r ag**gregator. It was built to meet a single need; to combine multiple yaml swagger files with varying server definitions into a single, [oats](https://github.com/influxdata/oats/) consumable one. If your needs differ, do not use this.

## Usage

```
  -api-title string
    	api title to print in output
  -api-version string
    	api version to print in output
  -file value
    	location of yaml swagger file (comma separated or define multiple times)
  -openapi-version string
    	version of openapi to print in output (default "3.0.0")
```

```sh
swagrag -file swagger1.yml -file swagger2.yml > combined.yml
```
