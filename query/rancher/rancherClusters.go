package rancher

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (r Client) GetNumberofClusters() (int, error) {
	lister := r.InformerManager.GetLister("provisioning_clusters", "")
	if lister == nil {
		return 0, fmt.Errorf("provisioning_clusters lister not found")
	}
	objs, err := lister.List(labels.Everything())
	if err != nil {
		return 0, err
	}
	return len(objs), nil
}

func (r Client) GetClusterLabels() ([]clusterLabel, error) {
	lister := r.InformerManager.GetLister("provisioning_clusters", "")
	if lister == nil {
		return nil, fmt.Errorf("provisioning_clusters lister not found")
	}
	objList, err := lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var clusterLabelsArray []clusterLabel

	for _, clusterValue := range objList {
		clusterObj, ok := clusterValue.(*unstructured.Unstructured)
		if !ok {
			continue
		}

		clusterLabels := clusterObj.GetLabels()
		clusterName, _, err := unstructured.NestedString(clusterObj.Object, "spec", "rkeConfig", "chartValues", "harvester-cloud-provider", "global", "cattle", "clusterName")
		if err != nil {
			continue
		}
		clusterID, _, err := unstructured.NestedString(clusterObj.Object, "status", "clusterName")
		if err != nil {
			continue
		}

		if clusterName != "" {

			for labelKey, labelValue := range clusterLabels {

				cluster := clusterLabel{
					ClusterId:          clusterID,
					ClusterDisplayName: clusterObj.GetName(),
					ClusterName:        "",
					LabelKey:           labelKey,
				LabelValue:         labelValue,
				}

				clusterLabelsArray = append(clusterLabelsArray, cluster)

			}
		} else {
			for labelKey, labelValue := range clusterLabels {

				cluster := clusterLabel{
					ClusterId:          clusterID,
					ClusterDisplayName: clusterObj.GetName(),
					ClusterName:        clusterName,
					LabelKey:           labelKey,
					LabelValue:         labelValue,
				}

				clusterLabelsArray = append(clusterLabelsArray, cluster)
			}
		}
	}
	return clusterLabelsArray, nil
}
