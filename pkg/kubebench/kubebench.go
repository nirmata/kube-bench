package kubebench

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	kubebench "github.com/aquasecurity/kube-bench/check"
	"github.com/nirmata/kube-bench/pkg/params"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	yaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

func getClientSet(kubeconfigPath string) (*kubernetes.Clientset, error) {
	var kubeconfig *rest.Config

	cfg, err := rest.InClusterConfig()
	if err != nil {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			klog.Fatalf("Error building kubeconfig: %s", err.Error())
			return nil, err
		}
	}
	kubeconfig = cfg

	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	return clientset, nil

}

func RunJob(params *params.KubeBenchArgs) (*kubebench.OverallControls, error) {
	clientset, err := getClientSet(params.Kubeconfig)
	if err != nil {
		return nil, err
	}

	if params.Registry != "aquasec" && params.RegistryUsername != "" && params.RegistryPassword != "" {
		secretName, err := deploySecret(context.Background(), clientset, params)
		if err != nil {
			return nil, err
		}

		_, err = clientset.CoreV1().Secrets(params.Namespace).Get(context.Background(), secretName, metav1.GetOptions{})

		if err != nil {
			fmt.Printf("failed to find secret for job %s\n", err)
			return nil, err
		}
	} else {
		requiredParam := make([]string, 2)
		if params.Registry != "aquasec" && params.RegistryUsername == "" {
			requiredParam = append(requiredParam, "registry-username")
		}

		if params.Registry != "aquasec" && params.RegistryPassword == "" {
			requiredParam = append(requiredParam, "registry-password")
		}
		fmt.Printf("failed to create imagePullSecret, pls specify required params %s\n", requiredParam)
	}

	var jobName string
	jobName, err = deployJob(context.Background(), clientset, params)
	if err != nil {
		return nil, err
	}

	p, err := findPodForJob(context.Background(), clientset, params, jobName, params.Timeout)
	if err != nil {
		fmt.Printf("failed to find pod for job %s\n", err)
		return nil, err
	}

	output, err := getPodLogs(context.Background(), clientset, jobName, p)
	if err != nil {
		fmt.Printf("error getting pod logs for job, %s. Error %v\n", jobName, err)
		return nil, err
	}

	err = clientset.BatchV1().Jobs(params.Namespace).Delete(context.Background(), jobName, metav1.DeleteOptions{})
	if err != nil {
		return nil, err
	}

	err = clientset.CoreV1().Pods(params.Namespace).Delete(context.Background(), p.Name, metav1.DeleteOptions{})
	if err != nil {
		return nil, err
	}

	controls, err := convert(output)
	if err != nil {
		return nil, err
	}

	return controls, nil
}

func deploySecret(ctx context.Context, clientset *kubernetes.Clientset, params *params.KubeBenchArgs) (string, error) {
	// Create a new secret object
	secret := &apiv1.Secret{
		Type: apiv1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "docker-image-pull-secret",
			Namespace: params.Namespace,
		},
		Data: map[string][]byte{
			".dockerconfigjson": []byte(`{"auths":{"` + params.Registry + `":{"username":"` + params.RegistryUsername + `","password":"` + params.RegistryPassword + `"}}}`),
		},
	}

	// Create the secret in the Kubernetes cluster
	_, err := clientset.CoreV1().Secrets(params.Namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}

	// Return the name of the created secret
	return secret.Name, nil
}

func deployJob(ctx context.Context, clientset *kubernetes.Clientset, params *params.KubeBenchArgs) (string, error) {
	jobYAML, err := embedYAMLs(params.KubebenchBenchmark, params.ClusterType, params.CustomJobFile)
	if err != nil {
		return "", err
	}

	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(jobYAML), len(jobYAML))
	job := &batchv1.Job{}
	if err := decoder.Decode(job); err != nil {
		return "", err
	}
	kubebenchImg := params.Registry + "/kube-bench:" + params.KubebenchTag
	jobName := job.GetName()
	job.Spec.Template.Spec.Containers[0].Image = kubebenchImg
	job.Spec.Template.Spec.Containers[0].Args = []string{"--json"}
	if params.KubebenchBenchmark != "" {
		job.Spec.Template.Spec.Containers[0].Args = append(job.Spec.Template.Spec.Containers[0].Args, "--benchmark="+params.KubebenchBenchmark)
	}
	if params.KubebenchTargets != "" {
		job.Spec.Template.Spec.Containers[0].Args = append(job.Spec.Template.Spec.Containers[0].Args, "--targets="+params.KubebenchTargets)
	}
	if params.NodeSelectorKey != "" {
		job.Spec.Template.Spec.NodeSelector = map[string]string{params.NodeSelectorKey: params.NodeSelectorValue}
	}

	_, err = clientset.BatchV1().Jobs(params.Namespace).Create(ctx, job, metav1.CreateOptions{})

	return jobName, err
}

func findPodForJob(ctx context.Context, clientset *kubernetes.Clientset, params *params.KubeBenchArgs, jobName string, duration time.Duration) (*apiv1.Pod, error) {
	failedPods := make(map[string]struct{})
	selector := fmt.Sprintf("job-name=%s", jobName)
	timeout := time.After(duration)
	for {
		time.Sleep(3 * time.Second)
	podfailed:
		select {
		case <-timeout:
			return nil, fmt.Errorf("podList - timed out: no Pod found for Job %s", jobName)
		default:
			pods, err := clientset.CoreV1().Pods(params.Namespace).List(ctx, metav1.ListOptions{
				LabelSelector: selector,
			})
			if err != nil {
				fmt.Printf("job listing error %v\n", err)
				return nil, err
			}
			fmt.Printf("Found (%d) pods\n", len(pods.Items))
			for _, cp := range pods.Items {
				if _, found := failedPods[cp.Name]; found {
					fmt.Printf("skip failed pod %s\n", cp.Name)
					continue
				}

				if strings.HasPrefix(cp.Name, jobName) {
					fmt.Printf("pod (%s) - %#v\n", cp.Name, cp.Status.Phase)
					if cp.Status.Phase == apiv1.PodSucceeded {
						fmt.Printf("pod %s succeeded\n", cp.Name)
						return &cp, nil
					}

					if cp.Status.Phase == apiv1.PodFailed {
						fmt.Printf("pod (%s) - %s - retrying...\n", cp.Name, cp.Status.Phase)
						fmt.Print(getPodLogs(ctx, clientset, jobName, &cp))
						failedPods[cp.Name] = struct{}{}
						break podfailed
					}
				}
			}
		}
	}
}

func getPodLogs(ctx context.Context, clientset *kubernetes.Clientset, jobName string, pod *apiv1.Pod) (string, error) {
	podLogOpts := apiv1.PodLogOptions{}
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}

	lines := strings.Split(buf.String(), "\n")
	for _, line := range lines {
		if len(line) > 0 && line[0] == '{' {
			return line, nil
		}
	}

	return "", errors.New("Couldn't find the JSON")
}

func convert(jsonString string) (*kubebench.OverallControls, error) {
	jsonDataReader := strings.NewReader(jsonString)
	decoder := json.NewDecoder(jsonDataReader)

	var controls kubebench.OverallControls
	if err := decoder.Decode(&controls); err != nil {
		return nil, err
	}

	return &controls, nil
}
