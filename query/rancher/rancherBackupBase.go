package rancher

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type backup struct {
	Name            string
	ResourceSetName string
	RetentionCount  int64
	BackupType      string
	Message         string
	Filename        string
	LastSnapshot    string
	NextSnapshot    string
	StorageLocation string
}

type restore struct {
	Name                 string
	Filename             string
	Prune                bool
	StorageLocation      string
	Message              string
	ResoreCompletionTime string
}

func (r Client) GetNumberOfBackups() (int, error) {
	lister := r.InformerManager.GetLister("backups", "")
	if lister == nil {
		return 0, fmt.Errorf("backups lister not found")
	}
	objs, err := lister.List(labels.Everything())
	if err != nil {
		return 0, err
	}
	return len(objs), nil
}

func (r Client) GetNumberOfRestores() (int, error) {
	lister := r.InformerManager.GetLister("restores", "")
	if lister == nil {
		return 0, fmt.Errorf("restores lister not found")
	}
	objs, err := lister.List(labels.Everything())
	if err != nil {
		return 0, err
	}
	return len(objs), nil
}

func (r Client) GetBackups() ([]backup, error) {
	var backups []backup

	lister := r.InformerManager.GetLister("backups", "")
	if lister == nil {
		return nil, fmt.Errorf("backups lister not found")
	}
	objList, err := lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, backupJob := range objList {
		backupObj, ok := backupJob.(*unstructured.Unstructured)
		if !ok {
			continue
		}

		backupName, _, err := unstructured.NestedString(backupObj.Object, "metadata", "name")
		if err != nil {
			continue
		}

		backupResourceSetName, _, err := unstructured.NestedString(backupObj.Object, "spec", "resourceSetName")
		if err != nil {
			continue
		}

		backupRetentionCount, _, err := unstructured.NestedInt64(backupObj.Object, "spec", "retentionCount")
		if err != nil {
			continue
		}

		backupType, _, err := unstructured.NestedString(backupObj.Object, "status", "backupType")
		if err != nil {
			continue
		}

		var backupMessage string
		var backupNextSnapshot string
		var backupLastSnapshot string

		if backupType == "One-time" {
			backupNextSnapshot = "N/A - One-time Backup"
		} else {
			backupNextSnapshot, _, err = unstructured.NestedString(backupObj.Object, "status", "nextSnapshotAt")
			if err != nil {
				continue
			}
		}

		backupLastSnapshot, _, err = unstructured.NestedString(backupObj.Object, "status", "lastSnapshotTs")
		if err != nil {
			continue
		}

		backupStorageLocation, _, err := unstructured.NestedString(backupObj.Object, "status", "storageLocation")
		if err != nil {
			continue
		}

		backupFileName, _, err := unstructured.NestedString(backupObj.Object, "status", "filename")
		if err != nil {
			continue
		}

		statusSlice, _, _ := unstructured.NestedSlice(backupObj.Object, "status", "conditions")
		for _, value := range statusSlice {
			for k, v := range value.(map[string]interface{}) {
				if k == "message" {
					backupMessage = v.(string)
				}
			}
		}

		backupInfo := backup{
			Name:            backupName,
			ResourceSetName: backupResourceSetName,
			RetentionCount:  backupRetentionCount,
			BackupType:      backupType,
			Message:         backupMessage,
			Filename:        backupFileName,
			NextSnapshot:    backupNextSnapshot,
			LastSnapshot:    backupLastSnapshot,
			StorageLocation: backupStorageLocation,
		}

		backups = append(backups, backupInfo)
	}
	return backups, nil
}

func (r Client) GetRestores() ([]restore, error) {
	var restores []restore

	lister := r.InformerManager.GetLister("restores", "")
	if lister == nil {
		return nil, fmt.Errorf("restores lister not found")
	}
	objList, err := lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, restoreJob := range objList {
		restoreObj, ok := restoreJob.(*unstructured.Unstructured)
		if !ok {
			continue
		}

		restoreName, _, err := unstructured.NestedString(restoreObj.Object, "metadata", "name")
		if err != nil {
			continue
		}

		fileName, _, err := unstructured.NestedString(restoreObj.Object, "spec", "backupFilename")
		if err != nil {
			continue
		}

		prune, _, err := unstructured.NestedBool(restoreObj.Object, "spec", "prune")
		if err != nil {
			continue
		}

		restoreStorageLocation, _, err := unstructured.NestedString(restoreObj.Object, "status", "backupSource")
		if err != nil {
			continue
		}

		var restoreMessage string
		statusSlice, _, err := unstructured.NestedSlice(restoreObj.Object, "status", "conditions")
		if err == nil {
			for _, value := range statusSlice {
				for k, v := range value.(map[string]interface{}) {
					if k == "message" {
						restoreMessage = v.(string)
					}
				}
			}
		}

		restoreTime, _, err := unstructured.NestedString(restoreObj.Object, "status", "restoreCompletionTs")
		if err != nil {
			continue
		}

		restoreInfo := restore{
			Name:                 restoreName,
			Filename:             fileName,
			Prune:                prune,
			StorageLocation:      restoreStorageLocation,
			Message:              restoreMessage,
			ResoreCompletionTime: restoreTime,
		}
		restores = append(restores, restoreInfo)
	}
	return restores, nil
}
