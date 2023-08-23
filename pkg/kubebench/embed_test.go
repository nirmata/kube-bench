package kubebench

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestEmbedYAMLs(t *testing.T) {
	testCases := map[string]string{
		"cis-1.5":                  "job.yaml",
		"cis-1.6":                  "job.yaml",
		"cis-1.20":                 "job.yaml",
		"cis-1.23":                 "job.yaml",
		"cis-1.24":                 "job.yaml",
		"cis-1.7":                  "job.yaml",
		"gke-1.0":                  "job-gke.yaml",
		"gke-1.2.0":                "job-gke.yaml",
		"eks-1.0.1":                "job-eks.yaml",
		"eks-1.1.0":                "job-eks.yaml",
		"eks-1.2.0":                "job-eks.yaml",
		"ack-1.0":                  "job.yaml",
		"aks-1.0":                  "job-aks.yaml",
		"rh-0.7":                   "job.yaml",
		"rh-1.0":                   "job.yaml",
		"cis-1.6-k3s":              "job.yaml",
		"eks-stig-kubernetes-v1r6": "job-eks.yaml",
		"tkgi-1.2.53":              "job.yaml",
		"xyz":                      "job.yaml",
	}

	i := 0
	for k, v := range testCases {
		t.Log("Running test", i)
		assert.Equal(t, getJobYamlName(k), v)
		i++
	}
}
