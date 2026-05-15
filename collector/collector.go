package collector

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promauto"

	"prometheus-rancher-exporter/query/rancher"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type rancherMetrics struct {
	installedRancherVersion prometheus.GaugeVec
	latestRancherVersion    prometheus.GaugeVec
	managedClusterCount     prometheus.Gauge
	managedK3sClusterCount  prometheus.Gauge
	managedRKEClusterCount  prometheus.Gauge
	managedRKE2ClusterCount prometheus.Gauge
	managedEKSClusterCount  prometheus.Gauge
	managedAKSClusterCount  prometheus.Gauge
	managedGKEClusterCount  prometheus.Gauge
	managedNodeCount        prometheus.Gauge

	// Cluster level rancherMetrics
	clusterConditionConnected    prometheus.GaugeVec
	clusterConditionNotConnected prometheus.GaugeVec

	// Downstream cluster rancherMetrics
	downstreamClusterVersion prometheus.GaugeVec

	// User related
	tokenCount prometheus.Gauge
	userCount  prometheus.Gauge

	// Project related
	projectCount       prometheus.Gauge
	projectLabels      prometheus.GaugeVec
	projectAnnotations prometheus.GaugeVec
	projectResources   prometheus.GaugeVec

	//Cluster related
	clusterLabels prometheus.GaugeVec

	// Extended rancherMetrics for Rancher CR's
	rancherCustomResources prometheus.GaugeVec

	// Additional Node Information
	managedNodeInfo prometheus.GaugeVec
}

type rancherBackupMetrics struct {
	backupCount     prometheus.Gauge
	restoreCount    prometheus.Gauge
	restoreSetCount prometheus.Gauge
	backup          prometheus.GaugeVec
	restore         prometheus.GaugeVec
}

func initRancherMetrics() rancherMetrics {
	m := rancherMetrics{

		installedRancherVersion: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "installed_rancher_version",
			Help: "version of the installed Rancher instance",
		}, []string{"version"},
		),
		latestRancherVersion: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "latest_rancher_version",
			Help: "version of the most recent Rancher release",
		}, []string{"version"},
		),
		managedClusterCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "rancher_managed_clusters",
			Help: "number of clusters this Rancher instance is currently managing",
		}),
		managedRKEClusterCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "rancher_managed_rke_clusters",
			Help: "number of RKE clusters this Rancher instance is currently managing",
		}),
		managedRKE2ClusterCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "rancher_managed_rke2_clusters",
			Help: "number of RKE2 clusters this Rancher instance is currently managing",
		}),
		managedK3sClusterCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "rancher_managed_k3s_clusters",
			Help: "number of K3s clusters this Rancher instance is currently managing",
		}),
		managedEKSClusterCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "rancher_managed_eks_clusters",
			Help: "number of RKE clusters this Rancher instance is currently managing",
		}),
		managedAKSClusterCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "rancher_managed_aks_clusters",
			Help: "number of RKE clusters this Rancher instance is currently managing",
		}),
		managedGKEClusterCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "rancher_managed_gke_clusters",
			Help: "number of RKE clusters this Rancher instance is currently managing",
		}),
		managedNodeCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "rancher_managed_nodes",
			Help: "number of managed nodes this Rancher instance is currently managing",
		}),
		clusterLabels: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "rancher_cluster_labels",
			Help: "labels associated with Rancher Clusters",
		}, []string{"cluster_id", "cluster_display_name", "cluster_name", "cluster_label_key", "cluster_label_value"},
		),
		clusterConditionConnected: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cluster_connected",
			Help: "identify if a downstream cluster is connected to Rancher",
		}, []string{"Name"},
		),
		clusterConditionNotConnected: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cluster_not_connected",
			Help: "identify if a downstream cluster is not connected to Rancher",
		}, []string{"Name"},
		),
		downstreamClusterVersion: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cluster_k8s_version",
			Help: "version of K8s running in the downstream cluster",
		}, []string{"Name", "Version"},
		),
		tokenCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "rancher_tokens",
			Help: "number of tokens issued by Rancher",
		}),
		userCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "rancher_users",
			Help: "number of users in this Rancher instance",
		}),
		projectCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "rancher_projects",
			Help: "number of Projects globally",
		}),
		projectLabels: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "rancher_project_labels",
			Help: "labels associated with Rancher Projects",
		}, []string{"cluster_name", "project_id", "project_display_name", "project_label_key", "project_label_value"},
		),
		projectAnnotations: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "rancher_project_annotations",
			Help: "annotations associated with Rancher Projects",
		}, []string{"cluster_name", "project_id", "project_display_name", "project_annotation_key", "project_annotation_value"},
		),
		projectResources: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "rancher_project_resourcequota",
			Help: "default namespace resource quota set the for project",
		}, []string{"cluster_name", "project_id", "project_display_name", "project_resource_key", "project_resource_type"},
		),
		rancherCustomResources: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "rancher_custom_resource_count",
			Help: "raw count of Rancher custom resources by name",
		}, []string{"resource_name"}),
		managedNodeInfo: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "rancher_managed_nodes_info",
			Help: "additional metadata about known downstream cluster nodes",
		}, []string{"name", "parent_cluster", "is_control_plane", "is_etcd", "is_worker", "architecture",
			"container_runtime_version", "kernel_version", "os", "os_image"},
		),
	}

	prometheus.MustRegister(m.installedRancherVersion)
	prometheus.MustRegister(m.latestRancherVersion)
	prometheus.MustRegister(m.managedClusterCount)
	prometheus.MustRegister(m.managedRKEClusterCount)
	prometheus.MustRegister(m.managedRKE2ClusterCount)
	prometheus.MustRegister(m.managedK3sClusterCount)
	prometheus.MustRegister(m.managedEKSClusterCount)
	prometheus.MustRegister(m.managedAKSClusterCount)
	prometheus.MustRegister(m.managedGKEClusterCount)
	prometheus.MustRegister(m.managedNodeCount)

	prometheus.MustRegister(m.clusterLabels)
	prometheus.MustRegister(m.clusterConditionConnected)
	prometheus.MustRegister(m.clusterConditionNotConnected)

	prometheus.MustRegister(m.downstreamClusterVersion)

	prometheus.MustRegister(m.tokenCount)
	prometheus.MustRegister(m.userCount)

	prometheus.MustRegister(m.projectCount)
	prometheus.MustRegister(m.projectLabels)
	prometheus.MustRegister(m.projectAnnotations)
	prometheus.MustRegister(m.projectResources)

	prometheus.MustRegister(m.rancherCustomResources)
	prometheus.MustRegister(m.managedNodeInfo)

	m.managedClusterCount.Set(0)
	m.managedRKEClusterCount.Set(0)
	m.managedRKE2ClusterCount.Set(0)
	m.managedK3sClusterCount.Set(0)
	m.managedEKSClusterCount.Set(0)
	m.managedAKSClusterCount.Set(0)
	m.managedGKEClusterCount.Set(0)
	m.managedNodeCount.Set(0)

	m.clusterConditionConnected.Reset()
	m.clusterConditionNotConnected.Reset()

	m.downstreamClusterVersion.Reset()

	m.tokenCount.Set(0)
	m.userCount.Set(0)

	m.projectCount.Set(0)
	m.managedNodeInfo.Reset()

	return m
}

func initRancherBackupMetrics(reg *prometheus.Registry) rancherBackupMetrics {

	rBackupMetrics := rancherBackupMetrics{
		backupCount: promauto.With(reg).NewGauge(prometheus.GaugeOpts{
			Name: "rancher_backups_count",
			Help: "number of rancher backups",
		}),
		restoreCount: promauto.With(reg).NewGauge(prometheus.GaugeOpts{
			Name: "rancher_restore_count",
			Help: "number of rancher restore",
		}),
		backup: *promauto.With(reg).NewGaugeVec(prometheus.GaugeOpts{
			Name: "rancher_backup",
			Help: "details regarding a specific backup operation",
		}, []string{"name", "resourceSetName", "retentionCount", "backupType", "status", "filename", "storageLocation", "nextSnapshot", "lastSnapshot"},
		),
		restore: *promauto.With(reg).NewGaugeVec(prometheus.GaugeOpts{
			Name: "rancher_restore",
			Help: "details regarding a specific restore operation",
		}, []string{"name", "fileName", "prune", "storageLocation", "status", "restoreTime"},
		),
	}

	rBackupMetrics.backupCount.Set(0)
	rBackupMetrics.restoreCount.Set(0)
	rBackupMetrics.backup.Reset()

	return rBackupMetrics
}

var baseMetrics rancherMetrics
var backupMetrics rancherBackupMetrics

func RegisterHandlers(client rancher.Client) {
	baseMetrics = initRancherMetrics()

	clusterHandler := func(obj interface{}) {
		log.Debug("Cluster changed, updating metrics")
		resetGaugeVecMetrics(baseMetrics)
		getNumberOfClusters(client, baseMetrics)
		getDistributions(client, baseMetrics)
		getClusterConnectedState(client, baseMetrics)
		getDownstreamClusterVersions(client, baseMetrics)
		getClusterLabels(client, baseMetrics)
	}

	nodeHandler := func(obj interface{}) {
		log.Debug("Node changed, updating metrics")
		getNumberOfNodes(client, baseMetrics)
		getNodeInfo(client, baseMetrics)
	}

	tokenHandler := func(obj interface{}) {
		log.Debug("Token changed, updating metrics")
		getNumberOfTokens(client, baseMetrics)
	}

	userHandler := func(obj interface{}) {
		log.Debug("User changed, updating metrics")
		getNumberOfUsers(client, baseMetrics)
	}

	projectHandler := func(obj interface{}) {
		log.Debug("Project changed, updating metrics")
		getNumberOfProjects(client, baseMetrics)
		getProjectLabels(client, baseMetrics)
		getProjectAnnotations(client, baseMetrics)
		getProjectResources(client, baseMetrics)
	}

	settingsHandler := func(obj interface{}) {
		log.Debug("Settings changed, updating metrics")
		getInstalledRancherVersion(client, baseMetrics)
	}

	client.InformerManager.AddEventHandler("clusters", clusterHandler)
	client.InformerManager.AddEventHandler("nodes", nodeHandler)
	client.InformerManager.AddEventHandler("tokens", tokenHandler)
	client.InformerManager.AddEventHandler("users", userHandler)
	client.InformerManager.AddEventHandler("projects", projectHandler)
	client.InformerManager.AddEventHandler("settings", settingsHandler)
}

func RunBaseMetricsLoop(client rancher.Client, Timer_GetLatestRancherVersion int) {
	ticker := time.NewTicker(time.Duration(Timer_GetLatestRancherVersion) * time.Minute)
	for ; ; <-ticker.C {
		baseMetrics.latestRancherVersion.Reset()
		latestVers, err := client.GetLatestRancherVersion()
		if err != nil {
			log.Errorf("error retrieving latest Rancher version: %v", err)
		}
		baseMetrics.latestRancherVersion.WithLabelValues(latestVers).Set(1)
	}
}

func RegisterBackupHandlers(client rancher.Client, reg *prometheus.Registry) {
	backupMetrics = initRancherBackupMetrics(reg)

	backupHandler := func(obj interface{}) {
		log.Debug("Backup changed, updating metrics")
		resetBackupGaugeVecMetrics(backupMetrics)
		getNumberOfBackups(client, backupMetrics)
		getBackups(client, backupMetrics)
	}

	restoreHandler := func(obj interface{}) {
		log.Debug("Restore changed, updating metrics")
		getNumberOfRestores(client, backupMetrics)
		getRestores(client, backupMetrics)
	}

	client.InformerManager.AddEventHandler("backups", backupHandler)
	client.InformerManager.AddEventHandler("restores", restoreHandler)
}



func getInstalledRancherVersion(client rancher.Client, m rancherMetrics) {

	start := time.Now()

	installedVersion, err := client.GetInstalledRancherVersion()
	if err != nil {
		log.Errorf("error retrieving the installed Rancher version: %v", err)
	}
	m.installedRancherVersion.WithLabelValues(installedVersion).Set(1)

	elapsed := time.Since(start)
	log.Debugf("getInstalledRancherVersion metric collection took %s", elapsed)
}

func getClusterConnectedState(client rancher.Client, m rancherMetrics) {
	state, err := client.GetClusterConnectedState()
	if err != nil {
		log.Errorf("error retrieving cluster connected states: %v", err)
	}
	for key, value := range state {
		if value == true {
			m.clusterConditionConnected.WithLabelValues(key).Set(1)
			m.clusterConditionNotConnected.WithLabelValues(key).Set(0)
		} else {
			m.clusterConditionNotConnected.WithLabelValues(key).Set(1)
			m.clusterConditionConnected.WithLabelValues(key).Set(0)
		}
	}
}

func getNumberOfClusters(client rancher.Client, m rancherMetrics) {
	numberOfClusters, err := client.GetNumberOfManagedClusters()

	if err != nil {
		log.Errorf("error retrieving number of managed clusters: %v", err)
	}
	m.managedClusterCount.Set(float64(numberOfClusters))
}

func getDistributions(client rancher.Client, m rancherMetrics) {
	distributions, err := client.GetK8sDistributions()
	if err != nil {
		log.Errorf("error retrieving cluster k8s distributions: %v", err)
	}
	m.managedRKEClusterCount.Set(float64(distributions["rke"]))
	m.managedRKE2ClusterCount.Set(float64(distributions["rke2"]))
	m.managedK3sClusterCount.Set(float64(distributions["k3s"]))
	m.managedEKSClusterCount.Set(float64(distributions["eks"]))
	m.managedAKSClusterCount.Set(float64(distributions["aks"]))
	m.managedGKEClusterCount.Set(float64(distributions["gke"]))
}

func getNumberOfNodes(client rancher.Client, m rancherMetrics) {
	numberOfNodes, err := client.GetNumberOfManagedNodes()

	if err != nil {
		log.Errorf("error retrieving number of managed nodes: %v", err)
	}

	m.managedNodeCount.Set(float64(numberOfNodes))
}

func getDownstreamClusterVersions(client rancher.Client, m rancherMetrics) {
	downstreamClusterVersions, err := client.GetDownstreamClusterVersions()

	if err != nil {
		log.Errorf("error retrieving downstream k8s cluster versions: %v", err)
	}

	for _, value := range downstreamClusterVersions {

		m.downstreamClusterVersion.WithLabelValues(value.Name, value.Version).Set(1)
	}
}

func getNumberOfUsers(client rancher.Client, m rancherMetrics) {
	users, err := client.GetNumberOfUsers()
	if err != nil {
		log.Errorf("error retrieving number of users: %v", err)
	}

	m.userCount.Set(float64(users))
}

func getNumberOfTokens(client rancher.Client, m rancherMetrics) {
	tokens, err := client.GetNumberOfTokens()
	if err != nil {
		log.Errorf("error retrieving number of tokens: %v", err)
	}

	m.tokenCount.Set(float64(tokens))
}

func getNumberOfProjects(client rancher.Client, m rancherMetrics) {
	projects, err := client.GetNumberofProjects()
	if err != nil {
		log.Errorf("error retrieving number of projects: %v", err)
	}

	m.projectCount.Set(float64(projects))
}

func getProjectLabels(client rancher.Client, m rancherMetrics) {
	projectLabels, err := client.GetProjectLabels()
	if err != nil {
		log.Errorf("error retrieving project labels: %v", err)
	}

	for _, value := range projectLabels {
		m.projectLabels.WithLabelValues(value.ProjectClusterName, value.Projectid, value.ProjectDisplayName, value.LabelKey, value.LabelValue).Set(1)

	}
}

func getProjectAnnotations(client rancher.Client, m rancherMetrics) {
	projectAnnotations, err := client.GetProjectAnnotations()
	if err != nil {
		log.Errorf("error retrieving project annotations: %v", err)
	}

	for _, value := range projectAnnotations {
		m.projectAnnotations.WithLabelValues(value.ProjectClusterName, value.Projectid, value.ProjectDisplayName, value.AnnotationKey, value.AnnotationValue).Set(1)
	}
}

func getProjectResources(client rancher.Client, m rancherMetrics) {
	projectResources, err := client.GetProjectResourceQuota()
	if err != nil {
		log.Errorf("error retrieving project resources: %v", err)
	}
	for _, value := range projectResources {
		m.projectResources.WithLabelValues(value.ProjectClusterName, value.Projectid, value.ProjectDisplayName, value.ResourceKey, value.ResourceType).Set(value.ResourceValue)
	}
}

func getClusterLabels(client rancher.Client, m rancherMetrics) {
	clusterLabels, err := client.GetClusterLabels()
	if err != nil {
		log.Errorf("error retrieving cluster labels: %v", err)
	}

	for _, value := range clusterLabels {
		m.clusterLabels.WithLabelValues(value.ClusterId, value.ClusterDisplayName, value.ClusterName, value.LabelKey, value.LabelValue).Set(1)

	}
}

func getRancherCustomResources(client rancher.Client, m rancherMetrics) {
	resources, err := client.GetRancherCustomResourceCount()
	if err != nil {
		return
	}
	for key, value := range resources {
		m.rancherCustomResources.WithLabelValues(key).Set(float64(value))
	}
}

func getNodeInfo(client rancher.Client, m rancherMetrics) {
	nodeResources, err := client.GetManagedNodeInfo()
	if err != nil {
		return
	}

	for _, value := range nodeResources {
		m.managedNodeInfo.WithLabelValues(value.Name, value.ParentCluster, strconv.FormatBool(value.IsControlPlane),
			strconv.FormatBool(value.IsEtcd), strconv.FormatBool(value.IsWorker), value.Architecture,
			value.ContainerRuntimeVersion, value.KernelVersion, value.OS, value.OSImage).Set(1)
	}
}

func getNumberOfBackups(client rancher.Client, m rancherBackupMetrics) {
	backups, err := client.GetNumberOfBackups()
	if err != nil {
		log.Errorf("error retrieving number of backups: %v", err)
	}

	m.backupCount.Set(float64(backups))
}

func getNumberOfRestores(client rancher.Client, m rancherBackupMetrics) {
	restores, err := client.GetNumberOfRestores()
	if err != nil {
		log.Errorf("error retrieving number of restores: %v", err)
	}

	m.restoreCount.Set(float64(restores))
}

func getBackups(client rancher.Client, m rancherBackupMetrics) {
	backups, err := client.GetBackups()
	if err != nil {
		log.Errorf("error retrieving backups: %v", err)
	}

	for _, value := range backups {
		m.backup.WithLabelValues(value.Name, value.ResourceSetName, strconv.FormatInt(value.RetentionCount, 10), value.BackupType, value.Message, value.Filename, value.StorageLocation, value.NextSnapshot, value.LastSnapshot).Set(1)
	}
}

func getRestores(client rancher.Client, m rancherBackupMetrics) {
	restores, err := client.GetRestores()
	if err != nil {
		log.Errorf("error retrieving backups: %v", err)
	}

	for _, value := range restores {
		m.restore.WithLabelValues(value.Name, value.Filename, strconv.FormatBool(value.Prune), value.StorageLocation, value.Message, value.ResoreCompletionTime).Set(1)
	}
}

// Reset GaugeVecs on each tick - facilitate state transition
func resetGaugeVecMetrics(m rancherMetrics) {
	m.installedRancherVersion.Reset()
	m.clusterLabels.Reset()
	m.clusterConditionConnected.Reset()
	m.clusterConditionNotConnected.Reset()
	m.downstreamClusterVersion.Reset()
	m.projectLabels.Reset()
	m.projectAnnotations.Reset()
	m.projectResources.Reset()
	m.managedNodeInfo.Reset()
}

func resetBackupGaugeVecMetrics(m rancherBackupMetrics) {
	m.backup.Reset()
	m.restore.Reset()
}
