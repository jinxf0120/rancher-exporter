package rancher

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (r Client) GetNumberofProjects() (int, error) {
	lister := r.InformerManager.GetLister("projects", "")
	if lister == nil {
		return 0, fmt.Errorf("projects lister not found")
	}
	objs, err := lister.List(labels.Everything())
	if err != nil {
		return 0, err
	}
	return len(objs), nil
}

func (r Client) GetProjectLabels() ([]projectLabel, error) {
	lister := r.InformerManager.GetLister("projects", "")
	if lister == nil {
		return nil, fmt.Errorf("projects lister not found")
	}
	objList, err := lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var projectLabelsArray []projectLabel

	for _, projectValue := range objList {
		projectObj, ok := projectValue.(*unstructured.Unstructured)
		if !ok {
			continue
		}

		projectLabels := projectObj.GetLabels()

		projectDisplayName, _, err := unstructured.NestedString(projectObj.Object, "spec", "displayName")
		if err != nil {
			continue
		}

		projectClusterID, _, err := unstructured.NestedString(projectObj.Object, "spec", "clusterName")
		if err != nil {
			continue
		}

		projectClusterName, _ := r.clusterIdToName(projectClusterID)

		if projectClusterName != "" {
			for labelKey, labelValue := range projectLabels {
				project := projectLabel{
					Projectid:          projectObj.GetName(),
					ProjectDisplayName: projectDisplayName,
					ProjectClusterName: projectClusterName,
					LabelKey:           labelKey,
					LabelValue:         labelValue,
				}
				projectLabelsArray = append(projectLabelsArray, project)
			}
		}
	}
	return projectLabelsArray, nil
}

func (r Client) GetProjectAnnotations() ([]projectAnnotation, error) {
	lister := r.InformerManager.GetLister("projects", "")
	if lister == nil {
		return nil, fmt.Errorf("projects lister not found")
	}
	objList, err := lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var projectAnnotationsArray []projectAnnotation

	for _, projectValue := range objList {
		projectObj, ok := projectValue.(*unstructured.Unstructured)
		if !ok {
			continue
		}

		projectDisplayName, _, err := unstructured.NestedString(projectObj.Object, "spec", "displayName")
		if err != nil {
			continue
		}

		projectClusterID, _, err := unstructured.NestedString(projectObj.Object, "spec", "clusterName")
		if err != nil {
			continue
		}

		projectClusterName, _ := r.clusterIdToName(projectClusterID)

		projectAnnotations := projectObj.GetAnnotations()

		if projectClusterName != "" {
			for annotationKey, annotationValue := range projectAnnotations {
				annotation := projectAnnotation{
					Projectid:          projectObj.GetName(),
					ProjectDisplayName: projectDisplayName,
					ProjectClusterName: projectClusterName,
					AnnotationKey:      annotationKey,
					AnnotationValue:    annotationValue,
				}
				projectAnnotationsArray = append(projectAnnotationsArray, annotation)
			}
		}
	}
	return projectAnnotationsArray, nil
}

func (r Client) GetProjectResourceQuota() ([]projectResource, error) {
	lister := r.InformerManager.GetLister("projects", "")
	if lister == nil {
		return nil, fmt.Errorf("projects lister not found")
	}
	objList, err := lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var projectResourceArray []projectResource

	for _, projectValue := range objList {
		projectObj, ok := projectValue.(*unstructured.Unstructured)
		if !ok {
			continue
		}

		projectDisplayName, _, err := unstructured.NestedString(projectObj.Object, "spec", "displayName")
		if err != nil {
			continue
		}

		projectClusterID, _, err := unstructured.NestedString(projectObj.Object, "spec", "clusterName")
		if err != nil {
			continue
		}

		projectClusterName, _ := r.clusterIdToName(projectClusterID)

		projectResourceQuotas, _, err := unstructured.NestedMap(projectObj.Object, "spec", "resourceQuota", "limit")

		if err != nil {
			continue
		}

		if projectClusterName != "" {
			for key, value := range projectResourceQuotas {
				var convertedValue float64
				quantity, err := resource.ParseQuantity(value.(string))
				if err != nil {
					continue
				}
				convertedValue = float64(quantity.Value())

				resource := projectResource{
					Projectid:          projectObj.GetName(),
					ProjectDisplayName: projectDisplayName,
					ProjectClusterName: projectClusterName,
					ResourceKey:        key,
					ResourceValue:      convertedValue,
					ResourceType:       "hard",
				}

				projectResourceArray = append(projectResourceArray, resource)
			}

			projectResourceQuotas, _, err = unstructured.NestedMap(projectObj.Object, "spec", "resourceQuota", "usedLimit")

			if err != nil {
				continue
			}

			for key, value := range projectResourceQuotas {
				var convertedValue float64
				quantity, err := resource.ParseQuantity(value.(string))
				if err != nil {
					continue
				}
				convertedValue = float64(quantity.Value())
				resource := projectResource{
					Projectid:          projectObj.GetName(),
					ProjectDisplayName: projectDisplayName,
					ProjectClusterName: projectClusterName,
					ResourceKey:        key,
					ResourceValue:      convertedValue,
					ResourceType:       "used",
				}

				projectResourceArray = append(projectResourceArray, resource)
			}
		}
	}

	return projectResourceArray, nil
}

func (r Client) clusterIdToName(id string) (string, error) {
	lister := r.InformerManager.GetLister("clusters", "")
	if lister == nil {
		return "", fmt.Errorf("clusters lister not found")
	}
	obj, err := lister.Get(id)
	if err != nil {
		return "", err
	}
	unstr, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return "", fmt.Errorf("failed to cast to unstructured")
	}
	clusterName, _, err := unstructured.NestedString(unstr.Object, "spec", "displayName")
	return clusterName, err
}