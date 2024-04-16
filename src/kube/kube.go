package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Kube struct {
	*kubernetes.Clientset
	Path    string
	Output  string
	Ignores []string
}

func New(path, configPath, output string) (*Kube, error) {
	// 读取Kubernetes 配置
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, err
	}
	restClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	// 读取程序配置路径
	// TODO

	var ignores = []string{"kube-root-ca.crt"}
	return &Kube{Clientset: restClient, Path: path, Output: output, Ignores: ignores}, nil
}
