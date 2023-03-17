// WUTONG, Application Management Platform
// Copyright (C) 2020-2020 Wutong Co., Ltd.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. For any non-GPL usage of Wutong,
// one or multiple Commercial Licenses authorized by Wutong Co., Ltd.
// must be obtained first.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package model

// CloudAccessKey cloud access key
type CloudAccessKey struct {
	Model
	ProviderName string `gorm:"column:provider_name" json:"provider_name"`
	AccessKey    string `gorm:"column:access_key" json:"access_key"`
	SecretKey    string `gorm:"column:secret_key" json:"secret_key"`
}

// CreateKubernetesTask create kubernetes task model
type CreateKubernetesTask struct {
	Model
	Name               string `gorm:"column:name" json:"name"`
	WorkerResourceType string `gorm:"column:resource_type" json:"resourceType"`
	WorkerNum          int    `gorm:"column:worker_num" json:"workerNum"`
	Provider           string `gorm:"column:provider_name" json:"providerName"`
	Region             string `gorm:"column:region" json:"region"`
	TaskID             string `gorm:"column:task_id" json:"taskID"`
	Status             string `gorm:"column:status" json:"status"`
	ClusterID          string `gorm:"column:cluster_id" json:"clusterID"`
}

// InitWutongTask init wutong task
type InitWutongTask struct {
	Model
	TaskID    string `gorm:"column:task_id" json:"taskID"`
	ClusterID string `gorm:"column:cluster_id" json:"clusterID"`
	Provider  string `gorm:"column:provider_name" json:"providerName"`
	Status    string `gorm:"column:status" json:"status"`
}

// UpdateKubernetesTask -
type UpdateKubernetesTask struct {
	Model
	TaskID    string `gorm:"column:task_id" json:"taskID"`
	ClusterID string `gorm:"column:cluster_id;uniqueIndex:version;type:varchar(64)" json:"clusterID"`
	// Version for optimistic lock
	Version    int    `gorm:"column:version;uniqueIndex:version;" json:"version"`
	Provider   string `gorm:"column:provider_name" json:"providerName"`
	NodeNumber int    `gorm:"column:node_number" json:"nodeNumber"`
	Status     string `gorm:"column:status" json:"status"`
}

// TaskEvent task event
type TaskEvent struct {
	Model
	TaskID   string `gorm:"column:task_id" json:"taskID"`
	StepType string `gorm:"column:step_type" json:"type"`
	Message  string `gorm:"column:message;size:512" json:"message"`
	Status   string `gorm:"column:status" json:"status"`
	EventID  string `gorm:"column:event_id" json:"eventID"`
	Reason   string `gorm:"column:reason" json:"reason"`
}

// BackupListModelData list all model data
type BackupListModelData struct {
	CloudAccessKeys       []CloudAccessKey       `json:"cloud_access_keys"`
	CreateKubernetesTasks []CreateKubernetesTask `json:"create_kubernetes_tasks"`
	InitWutongTasks       []InitWutongTask       `json:"init_wutong_tasks"`
	TaskEvents            []TaskEvent            `json:"task_events"`
	UpdateKubernetesTasks []UpdateKubernetesTask `json:"update_kubernetes_tasks"`
	CustomClusters        []CustomCluster        `json:"custom_clusters"`
	RKEClusters           []RKECluster           `json:"rke_clusters"`
	WutongClusterConfigs  []WutongClusterConfig  `json:"wutong_cluster_configs"`
	AppStores             []AppStore             `json:"app_stores"`
}
