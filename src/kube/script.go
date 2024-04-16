package kube

import (
	"embed"
	"strings"
)

//go:embed help.sh
var content embed.FS

func Generate(path, namespace string) string {
	data, _ := content.ReadFile("help.sh")
	script := string(data)
	script = strings.ReplaceAll(script, "GL_KUBE_CONF=", "GL_KUBE_CONF="+path)
	script = strings.ReplaceAll(script, "GL_KUBE_NAMESPACE=", "GL_KUBE_NAMESPACE="+namespace)
	return script
}
