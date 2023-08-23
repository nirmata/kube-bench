// package
package main

import (
	"fmt"
	"os"

	"github.com/nirmata/kube-bench/pkg/kubebench"
	"github.com/nirmata/kube-bench/pkg/params"
	"github.com/nirmata/kube-bench/pkg/report"
)

func main() {
	//parse arguments
	params.ParseArguments()

	//run kube-bench job
	cis, err := kubebench.RunJob(&params.Params)
	if err != nil {
		fmt.Printf("failed to run job of kube-bench: %v \n", err)
		os.Exit(-1)
	}

	// create policy report
	r, err := report.New(cis, params.Params.Name, params.Params.KubebenchBenchmark, params.Params.Category)
	if err != nil {
		fmt.Printf("failed to create policy reports: %v \n", err)
		os.Exit(-1)
	}

	// write policy report
	r, err = report.Write(r, params.Params.Kubeconfig)
	if err != nil {
		fmt.Printf("failed to create policy reports: %v \n", err)
		os.Exit(-1)
	}

	fmt.Printf("wrote policy report %s \n", r.Name)
}
