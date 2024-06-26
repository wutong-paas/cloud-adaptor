// WUTONG, Application Management Platform
// Copyright (C) 2020-2021 Wutong Co., Ltd.

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

package domain

// ClusterTaskType -
type ClusterTaskType string

// ClusterTaskType -
var (
	ClusterTaskTypeInitWutong       ClusterTaskType = "init-wutong"
	ClusterTaskTypeCreateKubernetes ClusterTaskType = "create-kubernetes"
	ClusterTaskTypeUpdateKubernetes ClusterTaskType = "update-kubernetes"
)

// Cluster -
type Cluster struct {
	Name       string `json:"name"`
	ClusterID  string `json:"clusterID"`
	KubeConfig string `json:"kubeConfig"`
	Provider   string `json:"provider"`
}

// ClusterTask -
type ClusterTask struct {
	ClusterID    string          `json:"clusterID"`
	ProviderName string          `json:"providerName"`
	TaskID       string          `json:"taskID"`
	TaskType     ClusterTaskType `json:"taskType"`
}
