package rancher

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"prometheus-rancher-exporter/informer"

	log "github.com/sirupsen/logrus"
)

type Client struct {
	Client          dynamic.Interface
	Config          *rest.Config
	InformerManager *informer.InformerManager
}

type clusterVersion struct {
	Name    string
	Version string
}

type projectLabel struct {
	Projectid          string
	ProjectDisplayName string
	ProjectClusterName string
	LabelKey           string
	LabelValue         string
}

type projectAnnotation struct {
	Projectid          string
	ProjectDisplayName string
	ProjectClusterName string
	AnnotationKey      string
	AnnotationValue    string
}

type projectResource struct {
	Projectid          string
	ProjectDisplayName string
	ProjectClusterName string
	ResourceKey        string
	ResourceValue      float64
	ResourceType       string
}

type clusterLabel struct {
	ClusterId          string
	ClusterDisplayName string
	ClusterName        string
	LabelKey           string
	LabelValue         string
}

type Release struct {
	TagName string `json:"tag_name"`
}

type nodeInfo struct {
	Name                    string
	ParentCluster           string
	IsControlPlane          bool
	IsEtcd                  bool
	IsWorker                bool
	Architecture            string
	ContainerRuntimeVersion string
	KernelVersion           string
	OS                      string
	OSImage                 string
}

func (r *Client) InitInformerManager(resyncPeriod int, includeBackupResources bool) error {
	r.InformerManager = informer.NewInformerManager(r.Client, resyncPeriod)

	resources := []struct {
		name     string
		group    string
		version  string
		resource string
	}{
		{"settings", "management.cattle.io", "v3", "settings"},
		{"clusters", "management.cattle.io", "v3", "clusters"},
		{"nodes", "management.cattle.io", "v3", "nodes"},
		{"tokens", "management.cattle.io", "v3", "tokens"},
		{"users", "management.cattle.io", "v3", "users"},
		{"projects", "management.cattle.io", "v3", "projects"},
		{"provisioning_clusters", "provisioning.cattle.io", "v1", "clusters"},
		{"crds", "apiextensions.k8s.io", "v1", "customresourcedefinitions"},
	}

	if includeBackupResources {
		resources = append(resources,
			[]struct {
				name     string
				group    string
				version  string
				resource string
			}{
				{"backups", "resources.cattle.io", "v1", "backups"},
				{"restores", "resources.cattle.io", "v1", "restores"},
			}...,
		)
	}

	for _, res := range resources {
		if err := r.InformerManager.CreateInformer(res.name, res.group, res.version, res.resource); err != nil {
			log.Warnf("Skipping informer for %s: %v", res.name, err)
		}
	}

	return nil
}

func (r *Client) StartInformer() error {
	if r.InformerManager == nil {
		return fmt.Errorf("informer manager not initialized")
	}
	return r.InformerManager.Start()
}

func (r Client) GetInstalledRancherVersion() (string, error) {
	lister := r.InformerManager.GetLister("settings", "")
	if lister == nil {
		return "", fmt.Errorf("settings lister not found")
	}
	obj, err := lister.Get("server-version")
	if err != nil {
		return "", err
	}
	unstr, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return "", fmt.Errorf("failed to cast to unstructured")
	}
	version, _, err := unstructured.NestedString(unstr.Object, "value")
	return version, err
}

func (r Client) GetNumberOfManagedClusters() (int, error) {
	lister := r.InformerManager.GetLister("clusters", "")
	if lister == nil {
		return 0, fmt.Errorf("clusters lister not found")
	}
	objs, err := lister.List(labels.Everything())
	if err != nil {
		return 0, err
	}
	count := 0
	for _, obj := range objs {
		unstr, ok := obj.(*unstructured.Unstructured)
		if !ok {
			continue
		}
		clusterName, _, _ := unstructured.NestedString(unstr.Object, "spec", "displayName")
		if clusterName != "local" {
			count++
		}
	}
	return count, nil
}

func (r Client) GetK8sDistributions() (map[string]int, error) {
	distributions := make(map[string]int)

	lister := r.InformerManager.GetLister("clusters", "")
	if lister == nil {
		return nil, fmt.Errorf("clusters lister not found")
	}
	objs, err := lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		unstr, ok := obj.(*unstructured.Unstructured)
		if !ok {
			continue
		}
		clusterName, _, _ := unstructured.NestedString(unstr.Object, "spec", "displayName")
		if clusterName == "local" {
			continue
		}
		resourceLabels := unstr.GetLabels()
		distribution := resourceLabels["provider.cattle.io"]
		distributions[distribution] += 1
	}
	return distributions, nil
}

func (r Client) GetLatestRancherVersion() (string, error) {

	client := &http.Client{Timeout: 10 * time.Second}

	_, err := net.LookupHost("api.github.com")
	if err != nil {
		return "unavailable", nil
	}

	resp, err := client.Get("https://api.github.com/repos/rancher/rancher/releases")
	if err != nil {
		return "unavailable", nil
	}

	defer resp.Body.Close()

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "unavailable", nil
	}

	var latestVersion string
	re := regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
	for _, release := range releases {
		if re.MatchString(release.TagName) && release.TagName > latestVersion {
			latestVersion = release.TagName
		}
	}

	return latestVersion, nil
}

func (r Client) GetNumberOfManagedNodes() (int, error) {
	lister := r.InformerManager.GetLister("nodes", "")
	if lister == nil {
		return 0, fmt.Errorf("nodes lister not found")
	}
	objs, err := lister.List(labels.Everything())
	if err != nil {
		return 0, err
	}
	return len(objs), nil
}

func (r Client) GetManagedNodeInfo() ([]nodeInfo, error) {
	var nodes []nodeInfo

	lister := r.InformerManager.GetLister("nodes", "")
	if lister == nil {
		return nil, fmt.Errorf("nodes lister not found")
	}
	objList, err := lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, obj := range objList {
		nodeObj, ok := obj.(*unstructured.Unstructured)
		if !ok {
			continue
		}

		nodeValue := nodeInfo{}

		parentClusterID, _, err := unstructured.NestedString(nodeObj.Object, "metadata", "namespace")
		if err != nil {
			continue
		}

		clusterName, err := r.clusterIdToName(parentClusterID)
		if err != nil {
			continue
		}

		if clusterName != "local" {

			nodeValue.ParentCluster = clusterName

			name, _, err := unstructured.NestedString(nodeObj.Object, "spec", "requestedHostname")
			if err != nil {
				continue
			}

			nodeValue.Name = name

			cpl, _, err := unstructured.NestedBool(nodeObj.Object, "spec", "controlPlane")
			if err != nil {
				continue
			}

			nodeValue.IsControlPlane = cpl

			etcd, _, err := unstructured.NestedBool(nodeObj.Object, "spec", "etcd")
			if err != nil {
				continue
			}

			nodeValue.IsEtcd = etcd

			worker, _, err := unstructured.NestedBool(nodeObj.Object, "spec", "worker")
			if err != nil {
				continue
			}

			nodeValue.IsWorker = worker

			architecture, _, err := unstructured.NestedString(nodeObj.Object, "status", "internalNodeStatus", "nodeInfo", "architecture")
			if err != nil {
				continue
			}

			nodeValue.Architecture = architecture

			containerRunTime, _, err := unstructured.NestedString(nodeObj.Object, "status", "internalNodeStatus", "nodeInfo", "containerRuntimeVersion")
			if err != nil {
				continue
			}

			nodeValue.ContainerRuntimeVersion = containerRunTime

			kernelVersion, _, err := unstructured.NestedString(nodeObj.Object, "status", "internalNodeStatus", "nodeInfo", "kernelVersion")
			if err != nil {
				continue
			}

			nodeValue.KernelVersion = kernelVersion

			OS, _, err := unstructured.NestedString(nodeObj.Object, "status", "internalNodeStatus", "nodeInfo", "operatingSystem")
			if err != nil {
				continue
			}

			nodeValue.OS = OS

			osImage, _, err := unstructured.NestedString(nodeObj.Object, "status", "internalNodeStatus", "nodeInfo", "osImage")
			if err != nil {
				continue
			}

			nodeValue.OSImage = osImage

			nodes = append(nodes, nodeValue)
		}
	}

	return nodes, nil
}

func (r Client) GetClusterConnectedState() (map[string]bool, error) {
	clusterStatus := make(map[string]bool)

	lister := r.InformerManager.GetLister("clusters", "")
	if lister == nil {
		return nil, fmt.Errorf("clusters lister not found")
	}
	objList, err := lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, cluster := range objList {
		clusterObj, ok := cluster.(*unstructured.Unstructured)
		if !ok {
			continue
		}

		clusterName, _, err := unstructured.NestedString(clusterObj.Object, "spec", "displayName")
		if err != nil {
			continue
		}

		if clusterName != "local" {
			clusterStatus[clusterName] = false

			statusSlice, _, _ := unstructured.NestedSlice(clusterObj.Object, "status", "conditions")

			for _, value := range statusSlice {
				foundStatus := false
				foundType := false

				for k, v := range value.(map[string]interface{}) {
					if k == "type" && v.(string) == "Connected" {
						foundType = true
					}

					if k == "status" && v.(string) == "True" {
						foundStatus = true
					}

					if foundStatus == true && foundType == true {
						clusterStatus[clusterName] = true
					}
				}
			}
		}
	}

	return clusterStatus, nil
}

func (r Client) GetDownstreamClusterVersions() ([]clusterVersion, error) {
	var clusters []clusterVersion

	lister := r.InformerManager.GetLister("clusters", "")
	if lister == nil {
		return nil, fmt.Errorf("clusters lister not found")
	}
	objList, err := lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, cluster := range objList {
		clusterObj, ok := cluster.(*unstructured.Unstructured)
		if !ok {
			continue
		}

		clusterName, _, err := unstructured.NestedString(clusterObj.Object, "spec", "displayName")
		if err != nil {
			continue
		}

		clusterK8sVersion, _, err := unstructured.NestedString(clusterObj.Object, "status", "version", "gitVersion")
		if err != nil {
			continue
		}

		if clusterK8sVersion != "" {
			clusterInfo := clusterVersion{
				Name:    clusterName,
				Version: clusterK8sVersion,
			}
			clusters = append(clusters, clusterInfo)
		}
	}

	return clusters, nil
}

func (r Client) GetNumberOfTokens() (int, error) {
	lister := r.InformerManager.GetLister("tokens", "")
	if lister == nil {
		return 0, fmt.Errorf("tokens lister not found")
	}
	objs, err := lister.List(labels.Everything())
	if err != nil {
		return 0, err
	}
	return len(objs), nil
}

func (r Client) GetNumberOfUsers() (int, error) {
	lister := r.InformerManager.GetLister("users", "")
	if lister == nil {
		return 0, fmt.Errorf("users lister not found")
	}
	objs, err := lister.List(labels.Everything())
	if err != nil {
		return 0, err
	}
	count := 0
	for _, obj := range objs {
		unstr, ok := obj.(*unstructured.Unstructured)
		if !ok {
			continue
		}
		userName, _, _ := unstructured.NestedString(unstr.Object, "username")
		if userName != "" {
			count++
		}
	}
	return count, nil
}
