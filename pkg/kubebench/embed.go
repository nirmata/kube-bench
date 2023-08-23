package kubebench

import (
	"embed"
	"strings"
)

//go:embed jobs/*.yaml
var yamlDir embed.FS

func getJobYamlName(benchmark string) (fname string) {
	benchmark = strings.Split(benchmark, "-")[0]
	switch benchmark {
	case "eks":
		fname = "job-eks.yaml"
	case "gke":
		fname = "job-gke.yaml"
	default:
		fname = "job.yaml"
	}

	return
}

func embedYAMLs(benchmark string) ([]byte, error) {
	var data []byte
	var err error

	data, err = yamlDir.ReadFile("jobs/" + getJobYamlName(benchmark))
	if err != nil {
		return nil, err
	}

	return data, nil
}
