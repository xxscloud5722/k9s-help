package kube

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
)
import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// RefreshProject 刷新工程.
func (kube *Kube) RefreshProject(ignore bool) error {
	namespaceItems := namespaces(kube)
	if namespaceItems == nil {
		return nil
	}
	for _, item := range namespaceItems {
		var script = Generate(kube.Path, item.ObjectMeta.Name)
		err := writer(script, path.Join(kube.Output, item.ObjectMeta.Name, item.ObjectMeta.Name+".sh"), 0755)
		if err != nil {
			if !ignore {
				return err
			}
		}
		err = deployment(kube, item.ObjectMeta.Name)
		if err != nil {
			if !ignore {
				return err
			}
		}
		err = statefulSet(kube, item.ObjectMeta.Name)
		if err != nil {
			if !ignore {
				return err
			}
		}
		err = daemonSet(kube, item.ObjectMeta.Name)
		if err != nil {
			if !ignore {
				return err
			}
		}
		err = job(kube, item.ObjectMeta.Name)
		if err != nil {
			if !ignore {
				return err
			}
		}
		err = cronJob(kube, item.ObjectMeta.Name)
		if err != nil {
			if !ignore {
				return err
			}
		}
		err = service(kube, item.ObjectMeta.Name)
		if err != nil {
			if !ignore {
				return err
			}
		}
		err = ingress(kube, item.ObjectMeta.Name)
		if err != nil {
			if !ignore {
				return err
			}
		}
		err = configmap(kube, item.ObjectMeta.Name)
		if err != nil {
			if !ignore {
				return err
			}
		}
		err = secret(kube, item.ObjectMeta.Name)
		if err != nil {
			if !ignore {
				return err
			}
		}
		err = persistentVolumeClaim(kube, item.ObjectMeta.Name)
		if err != nil {
			if !ignore {
				return err
			}
		}
	}
	err := persistentVolume(kube)
	if err != nil {
		if !ignore {
			return err
		}
	}
	return nil
}

func writer(value any, outputPath string, mode ...os.FileMode) error {
	directory := filepath.Dir(outputPath)
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		// 目录不存在，创建路径上的所有目录
		err = os.MkdirAll(directory, 0755)
		if err != nil {
			return err
		}
	}

	var content []byte
	var err error
	if reflect.TypeOf(value).Kind() == reflect.String {
		content = []byte(value.(string))
	} else {
		content, err = YAMLEncode(value)
		if err != nil {
			return err
		}
	}
	if len(mode) > 0 {
		err = os.WriteFile(outputPath, content, mode[0])
	} else {
		err = os.WriteFile(outputPath, content, 0644)
	}
	if err != nil {
		return err
	}
	return nil
}

func YAMLEncode(value any) ([]byte, error) {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var jsonContent any
	err = json.Unmarshal(jsonBytes, &jsonContent)
	if err != nil {
		return nil, err
	}

	jsonContent, err = cleanYAML(value, jsonContent)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	defer func(encoder *yaml.Encoder) {
		_ = encoder.Close()
	}(encoder)
	encoder.SetIndent(2)

	err = encoder.Encode(jsonContent)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func cleanYAML(source any, content any) (any, error) {
	var dataType = reflect.TypeOf(source).Name()
	var cMap = content.(map[string]interface{})
	if dataType == "Deployment" || dataType == "StatefulSet" || dataType == "DaemonSet" {
		cMap = cleanField(cMap, "metadata", "namespace", "annotations", "creationTimestamp", "generation", "managedFields", "uid", "resourceVersion")
		cMap = cleanField(cMap, "spec", "progressDeadlineSeconds", "replicas", "revisionHistoryLimit", "strategy")
		cMap = cleanField(cMap, "spec.template.metadata", "creationTimestamp", "annotations")
		cMap = cleanField(cMap, "spec.template.spec", "dnsPolicy", "schedulerName", "securityContext", "terminationGracePeriodSeconds")
		cMap = cleanField(cMap, "spec.template.spec.containers", "terminationMessagePath", "terminationMessagePolicy")
		cMap = cleanField(cMap, "status")
	}
	if dataType == "Service" {
		cMap = cleanField(cMap, "metadata", "namespace", "annotations", "creationTimestamp", "managedFields", "uid", "resourceVersion")
		cMap = cleanField(cMap, "spec", "clusterIP", "clusterIPs", "internalTrafficPolicy", "ipFamilies", "ipFamilyPolicy", "sessionAffinity")
		cMap = cleanField(cMap, "status")
	}
	if dataType == "Secret" || dataType == "ConfigMap" {
		cMap = cleanField(cMap, "metadata", "namespace", "annotations", "creationTimestamp", "managedFields", "uid", "resourceVersion")
	}
	if dataType == "Ingress" {
		cMap = cleanField(cMap, "metadata", "namespace", "annotations", "creationTimestamp", "generation", "managedFields", "uid", "resourceVersion")
		cMap = cleanField(cMap, "status")
	}
	if dataType == "PersistentVolume" {
		cMap = cleanField(cMap, "metadata", "namespace", "annotations", "creationTimestamp", "managedFields", "uid", "resourceVersion")
		cMap = cleanField(cMap, "spec.claimRef", "resourceVersion", "uid")
		cMap = cleanField(cMap, "status")
	}
	if dataType == "PersistentVolumeClaim" {
		cMap = cleanField(cMap, "metadata", "namespace", "annotations", "creationTimestamp", "managedFields", "uid", "resourceVersion")
		cMap = cleanField(cMap, "status")
	}
	return cMap, nil
}
func cleanField(source map[string]interface{}, keys string, items ...string) map[string]interface{} {
	if len(items) == 0 {
		delete(source, keys)
	} else {
		var b any = source
		for _, key := range strings.Split(keys, ".") {
			b = b.(map[string]interface{})[key]
		}
		dataType := reflect.TypeOf(b)
		if dataType.Kind() == reflect.Array || dataType.Kind() == reflect.Slice {
			c := b.([]interface{})
			for _, row := range c {
				for _, it := range items {
					delete(row.(map[string]interface{}), it)
				}
			}
		} else {
			for _, it := range items {
				delete(b.(map[string]interface{}), it)
			}
		}
	}
	return source
}

func namespaces(kube *Kube) []v1.Namespace {
	list, err := kube.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil
	}
	return list.Items
}
func deployment(kube *Kube, namespace string) error {
	items, err := kube.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, it := range items.Items {
		if lo.Contains(kube.Ignores, it.Name) {
			continue
		}
		it.APIVersion = "apps/v1"
		it.Kind = "Deployment"
		err = writer(it, path.Join(kube.Output, namespace, "deployment", it.Name+".yaml"))
		if err != nil {
			return err
		}
	}
	return nil
}
func statefulSet(kube *Kube, namespace string) error {
	items, err := kube.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, it := range items.Items {
		if lo.Contains(kube.Ignores, it.Name) {
			continue
		}
		it.APIVersion = "apps/v1"
		it.Kind = "StatefulSet"
		err = writer(it, path.Join(kube.Output, namespace, "statefulSet", it.Name+".yaml"))
		if err != nil {
			return err
		}
	}
	return nil
}
func daemonSet(kube *Kube, namespace string) error {
	items, err := kube.AppsV1().DaemonSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, it := range items.Items {
		if lo.Contains(kube.Ignores, it.Name) {
			continue
		}
		it.APIVersion = "apps/v1"
		it.Kind = "DaemonSet"
		err = writer(it, path.Join(kube.Output, namespace, "daemonSet", it.Name+".yaml"))
		if err != nil {
			return err
		}
	}
	return nil
}
func job(kube *Kube, namespace string) error {
	items, err := kube.BatchV1().Jobs(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, it := range items.Items {
		if lo.Contains(kube.Ignores, it.Name) {
			continue
		}
		it.APIVersion = "batch/v1"
		it.Kind = "Job"
		err = writer(it, path.Join(kube.Output, namespace, "job", it.Name+".yaml"))
		if err != nil {
			return err
		}
	}
	return nil
}
func cronJob(kube *Kube, namespace string) error {
	items, err := kube.BatchV1().CronJobs(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, it := range items.Items {
		if lo.Contains(kube.Ignores, it.Name) {
			continue
		}
		it.APIVersion = "batch/v1"
		it.Kind = "CronJob"
		err = writer(it, path.Join(kube.Output, namespace, "cronJob", it.Name+".yaml"))
		if err != nil {
			return err
		}
	}
	return nil
}
func service(kube *Kube, namespace string) error {
	items, err := kube.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, it := range items.Items {
		if lo.Contains(kube.Ignores, it.Name) {
			continue
		}
		it.APIVersion = "v1"
		it.Kind = "Service"
		err = writer(it, path.Join(kube.Output, namespace, "service", it.Name+".yaml"))
		if err != nil {
			return err
		}
	}
	return nil
}
func ingress(kube *Kube, namespace string) error {
	items, err := kube.NetworkingV1().Ingresses(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, it := range items.Items {
		if lo.Contains(kube.Ignores, it.Name) {
			continue
		}
		it.APIVersion = "networking.k8s.io/v1"
		it.Kind = "Ingress"
		err = writer(it, path.Join(kube.Output, namespace, "ingress", it.Name+".yaml"))
		if err != nil {
			return err
		}
	}
	return nil
}
func configmap(kube *Kube, namespace string) error {
	items, err := kube.CoreV1().ConfigMaps(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, it := range items.Items {
		if lo.Contains(kube.Ignores, it.Name) {
			continue
		}
		it.APIVersion = "v1"
		it.Kind = "ConfigMap"
		err = writer(it, path.Join(kube.Output, namespace, "configmap", it.Name+".yaml"))
		if err != nil {
			return err
		}
	}
	return nil
}
func secret(kube *Kube, namespace string) error {
	items, err := kube.CoreV1().Secrets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, it := range items.Items {
		if lo.Contains(kube.Ignores, it.Name) {
			continue
		}
		it.APIVersion = "v1"
		it.Kind = "Secret"
		err = writer(it, path.Join(kube.Output, namespace, "secret", it.Name+".yaml"))
		if err != nil {
			return err
		}
	}
	return nil
}
func persistentVolumeClaim(kube *Kube, namespace string) error {
	items, err := kube.CoreV1().PersistentVolumeClaims(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, it := range items.Items {
		if lo.Contains(kube.Ignores, it.Name) {
			continue
		}
		it.APIVersion = "v1"
		it.Kind = "PersistentVolumeClaim"
		err = writer(it, path.Join(kube.Output, namespace, "persistentVolumeClaim", it.Name+".yaml"))
		if err != nil {
			return err
		}
	}
	return nil
}
func persistentVolume(kube *Kube) error {
	items, err := kube.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, it := range items.Items {
		if lo.Contains(kube.Ignores, it.Name) {
			continue
		}
		it.APIVersion = "v1"
		it.Kind = "PersistentVolume"
		err = writer(it, path.Join(kube.Output, "persistentVolume", it.Name+".yaml"))
		if err != nil {
			return err
		}
	}
	return nil
}
