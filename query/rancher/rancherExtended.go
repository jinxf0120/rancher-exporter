package rancher

import (
	"context"
	"fmt"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	crdGVR               = schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}
	customResourceDomain = "cattle.io"
)

func (r Client) GetRancherCustomResourceCount() (map[string]int, error) {
	rancherCustomResources := make(map[string]int)
	var m sync.Mutex
	var wg sync.WaitGroup

	lister := r.InformerManager.GetLister("crds", "")
	if lister == nil {
		return nil, fmt.Errorf("crds lister not found")
	}
	objList, err := lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, customResource := range objList {
		crdObj, ok := customResource.(*unstructured.Unstructured)
		if !ok {
			continue
		}

		if !strings.Contains(crdObj.GetName(), customResourceDomain) {
			continue
		}

		wg.Add(1)
		go func(rancherCustomResource *unstructured.Unstructured) {
			defer wg.Done()
			resource, group, _ := strings.Cut(rancherCustomResource.GetName(), ".")
			version, _, err := unstructured.NestedSlice(rancherCustomResource.Object, "status", "storedVersions")
			if err != nil {
				log.Errorf("error retrieving version of Rancher CRD: %v", err)
				return
			}

			if len(version) == 0 {
				return
			}

			result, err := r.Client.Resource(schema.GroupVersionResource{
				Group:    group,
				Version:  version[0].(string),
				Resource: resource,
			}).List(context.Background(), v1.ListOptions{})

			if err != nil {
				log.Errorf("error retrieving count of Rancher CRD: %v,%s,%s,%s\n", err, group, version[0].(string), resource)
				return
			}

			m.Lock()
			rancherCustomResources[rancherCustomResource.GetName()] = len(result.Items)
			m.Unlock()
		}(crdObj)
	}
	wg.Wait()
	return rancherCustomResources, nil
}