package kubebench

import (
	"embed"
)

//go:embed jobs/*.yaml
var yamlDir embed.FS

func embedYAMLs(benchmark, clusterType, customJobFile string) ([]byte, error) {
	var data []byte
	var err error

	jobFile := "jobs/job.yaml"
	if customJobFile != "" {
		jobFile = customJobFile
	} else if clusterType != "" {
		jobFile = "jobs/job-" + clusterType + ".yaml"
	}
	data, err = yamlDir.ReadFile(jobFile)
	if err != nil {
		return nil, err
	}

	return data, nil
}
