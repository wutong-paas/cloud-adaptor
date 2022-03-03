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
	"fmt"

	"github.com/pkg/errors"
	"github.com/wutong-paas/cloud-adaptor/internal/model"
	"github.com/wutong-paas/cloud-adaptor/pkg/bcode"
	"github.com/wutong-paas/cloud-adaptor/pkg/util/uuidutil"
	"gorm.io/gorm"
)

// InitWutongRegionTaskRepo enterprise create kubernetes task
type InitWutongRegionTaskRepo struct {
	DB *gorm.DB `inject:""`
}

// NewInitWutongRegionTaskRepo new Enterprise repoo
func NewInitWutongRegionTaskRepo(db *gorm.DB) InitWutongTaskRepository {
	return &InitWutongRegionTaskRepo{DB: db}
}

// Transaction -
func (c *InitWutongRegionTaskRepo) Transaction(tx *gorm.DB) InitWutongTaskRepository {
	return &InitWutongRegionTaskRepo{DB: tx}
}

//Create create a task
func (c *InitWutongRegionTaskRepo) Create(ck *model.InitWutongTask) error {
	var old model.InitWutongTask
	if ck.TaskID == "" {
		ck.TaskID = uuidutil.NewUUID()
	}
	if err := c.DB.Where("eid = ? and task_id=? and cluster_id=?", ck.EnterpriseID, ck.TaskID, ck.ClusterID).Take(&old).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// not found error, create new
			if err := c.DB.Save(ck).Error; err != nil {
				return err
			}
			return nil
		}
		return err
	}
	return fmt.Errorf("task is exit")
}

//GetTaskByClusterID get cluster task
func (c *InitWutongRegionTaskRepo) GetTaskByClusterID(eid string, providerName, clusterID string) (*model.InitWutongTask, error) {
	var old model.InitWutongTask
	if err := c.DB.Where("eid=? and provider_name=? and cluster_id=?", eid, providerName, clusterID).Last(&old).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.Wrap(bcode.ErrInitWutongTaskNotFound, "get init wutong task")
		}
		return nil, errors.Wrap(err, "get init wutong task")
	}
	return &old, nil
}

//UpdateStatus update status
func (c *InitWutongRegionTaskRepo) UpdateStatus(eid string, taskID string, status string) error {
	var old model.InitWutongTask
	if err := c.DB.Model(&old).Where("eid = ? and task_id=?", eid, taskID).Update("status", status).Error; err != nil {
		return err
	}
	return nil
}

//GetTask get task
func (c *InitWutongRegionTaskRepo) GetTask(eid string, taskID string) (*model.InitWutongTask, error) {
	var old model.InitWutongTask
	if err := c.DB.Where("eid = ? and task_id=?", eid, taskID).Take(&old).Error; err != nil {
		return nil, err
	}
	return &old, nil
}

//GetTaskRunningLists get not complete tasks
func (c *InitWutongRegionTaskRepo) GetTaskRunningLists(eid string) ([]*model.InitWutongTask, error) {
	var list []*model.InitWutongTask
	if err := c.DB.Where("eid = ? and status != ?", eid, "complete").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

//DeleteTask -
func (c *InitWutongRegionTaskRepo) DeleteTask(eid string, providerName, clusterID string) error {
	var old model.InitWutongTask
	if err := c.DB.Where("eid = ? and provider_name=? and cluster_id=?", eid, providerName, clusterID).Delete(&old).Error; err != nil {
		return err
	}
	return nil
}
