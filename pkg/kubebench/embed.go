package kubebench

import (
	"embed"
)

//go:embed jobs/*.yaml
var yamlDir embed.FS

func embedYAMLs(benchmark string) ([]byte, error) {
	var data []byte
	var err error

	data, err = yamlDir.ReadFile("jobs/job.yaml")
	if err != nil {
		return nil, err
	}

	return data, nil
}
