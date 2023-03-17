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

package repo

import (
	"github.com/wutong-paas/cloud-adaptor/internal/model"
	"gorm.io/gorm"
)

// CloudAccesskeyRepository
type CloudAccesskeyRepository interface {
	Create(ent *model.CloudAccessKey) error
	GetByProvider(providerName string) (*model.CloudAccessKey, error)
}

// CreateKubernetesTaskRepository
type CreateKubernetesTaskRepository interface {
	Transaction(tx *gorm.DB) CreateKubernetesTaskRepository
	Create(ent *model.CreateKubernetesTask) error
	GetLastTask(providerName string) (*model.CreateKubernetesTask, error)
	UpdateStatus(taskID string, status string) error
	GetTask(taskID string) (*model.CreateKubernetesTask, error)
	GetLatestOneByName(name string) (*model.CreateKubernetesTask, error)
	GetLatestOneByClusterID(clusterID string) (*model.CreateKubernetesTask, error)
}

// InitWutongTaskRepository init wutong region task
type InitWutongTaskRepository interface {
	Transaction(tx *gorm.DB) InitWutongTaskRepository
	Create(ent *model.InitWutongTask) error
	GetTaskByClusterID(providerName, clusterID string) (*model.InitWutongTask, error)
	UpdateStatus(taskID string, status string) error
	GetTask(taskID string) (*model.InitWutongTask, error)
	DeleteTask(providerName, clusterID string) error
	GetTaskRunningLists() ([]*model.InitWutongTask, error)
}

// UpdateKubernetesTaskRepository -
type UpdateKubernetesTaskRepository interface {
	Transaction(tx *gorm.DB) UpdateKubernetesTaskRepository
	Create(ent *model.UpdateKubernetesTask) error
	GetTaskByClusterID(clusterID string) (*model.UpdateKubernetesTask, error)
	UpdateStatus(taskID string, status string) error
	GetTask(taskID string) (*model.UpdateKubernetesTask, error)
	GetLastTask(providerName string) (*model.UpdateKubernetesTask, error)
}

// TaskEventRepository task event
type TaskEventRepository interface {
	Transaction(tx *gorm.DB) TaskEventRepository
	Create(ent *model.TaskEvent) error
	ListEvent(taskID string) ([]*model.TaskEvent, error)
	UpdateStatusInBatch(eventIDs []string, status string) error
}

// WutongClusterConfigRepository -
type WutongClusterConfigRepository interface {
	Create(ent *model.WutongClusterConfig) error
	Get(clusterID string) (*model.WutongClusterConfig, error)
}

// RKEClusterRepository -
type RKEClusterRepository interface {
	Create(te *model.RKECluster) error
	Update(te *model.RKECluster) error
	GetCluster(name string) (*model.RKECluster, error)
	ListCluster() ([]*model.RKECluster, error)
	DeleteCluster(name string) error
}

// CustomClusterRepository -
type CustomClusterRepository interface {
	Create(cluster *model.CustomCluster) error
	Update(cluster *model.CustomCluster) error
	GetCluster(name string) (*model.CustomCluster, error)
	ListCluster() ([]*model.CustomCluster, error)
	DeleteCluster(name string) error
}
