package utils

import (
	"context"
	"prometheus-rancher-exporter/query/rancher"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	crdGVR = schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}
)

// CheckInstalledRancherApps Determines if both Rancher and the Backup operator are installed
func CheckInstalledRancherApps(r rancher.Client) (bool, bool, error) {

	var rancherInstalled = false
	var rancherBackupInstalled = false

	res, err := r.Client.Resource(crdGVR).List(context.Background(), v1.ListOptions{})
	if err != nil {
		return false, false, err
	}

	for _, customResource := range res.Items {

		if customResource.GetName() == "catalogs.management.cattle.io" {
			rancherInstalled = true
		}

		if customResource.GetName() == "backups.resources.cattle.io" {
			rancherBackupInstalled = true
		}
	}

	return rancherInstalled, rancherBackupInstalled, nil
}
