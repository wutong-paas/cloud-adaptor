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

	"github.com/wutong-paas/cloud-adaptor/internal/model"
	"github.com/wutong-paas/cloud-adaptor/pkg/util/uuidutil"
	"gorm.io/gorm"
)

// CustomClusterRepo -
type CustomClusterRepo struct {
	DB *gorm.DB `inject:""`
}

// NewCustomClusterRepo
func NewCustomClusterRepo(db *gorm.DB) *CustomClusterRepo {
	return &CustomClusterRepo{DB: db}
}

// NewCustomClusterRepository
func NewCustomClusterRepository(db *gorm.DB) CustomClusterRepository {
	return &CustomClusterRepo{DB: db}
}

// Create create an event
func (t *CustomClusterRepo) Create(te *model.CustomCluster) error {
	if te.Name == "" {
		return fmt.Errorf("custom cluster name can not be empty")
	}
	var old model.CustomCluster
	if err := t.DB.Where("name = ?", te.Name).Take(&old).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// not found error, create new
			if te.ClusterID == "" {
				te.ClusterID = uuidutil.NewUUID()
			}
			if err := t.DB.Save(te).Error; err != nil {
				return err
			}
			return nil
		}
		return err
	}
	return fmt.Errorf("rke cluster %s is exist", te.Name)
}

// Update -
func (t *CustomClusterRepo) Update(te *model.CustomCluster) error {
	return t.DB.Save(te).Error
}

// GetCluster -
func (t *CustomClusterRepo) GetCluster(name string) (*model.CustomCluster, error) {
	var rc model.CustomCluster
	if err := t.DB.Where("name=? or clusterID=?", name, name).Take(&rc).Error; err != nil {
		return nil, err
	}
	return &rc, nil
}

// ListCluster -
func (t *CustomClusterRepo) ListCluster() ([]*model.CustomCluster, error) {
	var list []*model.CustomCluster
	if err := t.DB.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// DeleteCluster delete cluster
func (t *CustomClusterRepo) DeleteCluster(name string) error {
	var rc model.CustomCluster
	if err := t.DB.Where("name=? or clusterID=?", name, name).Delete(&rc).Error; err != nil {
		return err
	}
	return nil
}
