package kube

func Generate(path, namespace string) string {
	return "kubectl --kubeconfig=" + path + " apply -n " + namespace + " -f ${1} "
}
