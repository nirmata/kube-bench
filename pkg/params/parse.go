package params

import (
	"flag"
	"path/filepath"
	"time"

	"k8s.io/client-go/util/homedir"
)

type KubeBenchArgs struct {
	Namespace string

	Name string

	Kubeconfig         string
	KubebenchImg       string
	KubebenchTargets   string
	KubebenchVersion   string
	KubebenchBenchmark string
	CustomJobFile      string
	ClusterType        string

	Timeout time.Duration

	NodeSelectorKey   string
	NodeSelectorValue string
}

var Params KubeBenchArgs

func ParseArguments() {
	flag.StringVar(&Params.Name, "name", "kube-bench", "name of policy report")
	flag.StringVar(&Params.Namespace, "namespace", "default", "namespace of kube-bench job")
	flag.StringVar(&Params.KubebenchTargets, "kube-bench-targets", "master,node,etcd,policies", "targets for benchmark of kube-bench job")
	flag.StringVar(&Params.KubebenchVersion, "kube-bench-version", "", "specify the Kubernetes version for kube-bench job")
	flag.StringVar(&Params.KubebenchBenchmark, "kube-bench-benchmark", "cis-1.7", "specify the benchmark for kube-bench job")
	flag.StringVar(&Params.CustomJobFile, "custom-job-file", "", "specify a custom kubebench job file if any")
	flag.StringVar(&Params.ClusterType, "cluster-type-override", "", "use non default cluster type in kube-bench, E.g aks, eks, eks-asff, gke, etc.")

	flag.StringVar(&Params.KubebenchImg, "kube-bench-image", "aquasec/kube-bench:v0.6.17", "kube-bench image used as part of this test")
	flag.DurationVar(&Params.Timeout, "timeout", 10*time.Minute, "Test Timeout")
	flag.StringVar(&Params.NodeSelectorKey, "nodeSelectorKey", "", "Job nameSelector Key")
	flag.StringVar(&Params.NodeSelectorValue, "nodeSelectorValue", "", "Job nameSelector value")

	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&Params.Kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&Params.Kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()
}
