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

// WutongClusterConfigRepo
type WutongClusterConfigRepo struct {
	DB *gorm.DB `inject:""`
}

// NewWutongClusterConfigRepo
func NewWutongClusterConfigRepo(db *gorm.DB) WutongClusterConfigRepository {
	return &WutongClusterConfigRepo{DB: db}
}

// Create create an event
func (t *WutongClusterConfigRepo) Create(te *model.WutongClusterConfig) error {
	var old model.WutongClusterConfig
	if err := t.DB.Where("clusterID=?", te.ClusterID).Take(&old).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := t.DB.Save(te).Error; err != nil {
				return err
			}
			return nil
		}
		return err
	}
	old.Config = te.Config
	return t.DB.Save(old).Error
}

// Get -
func (t *WutongClusterConfigRepo) Get(clusterID string) (*model.WutongClusterConfig, error) {
	var rcc model.WutongClusterConfig
	if err := t.DB.Where("clusterID=?", clusterID).Take(&rcc).Error; err != nil {
		return nil, err
	}
	return &rcc, nil
}
