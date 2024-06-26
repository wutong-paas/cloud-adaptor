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

package datastore

import (
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"github.com/wutong-paas/cloud-adaptor/cmd/cloud-adaptor/config"
	"github.com/wutong-paas/cloud-adaptor/internal/model"
	gmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var gdb *gorm.DB

// NewDB creates a new gorm.DB
func NewDB() *gorm.DB {
	var db *gorm.DB
	gormCfg := &gorm.Config{
		NamingStrategy: &schema.NamingStrategy{
			TablePrefix: "adaptor_",
		},
	}

	mySQLConfig := &mysql.Config{
		User:                 config.C.DB.User,
		Passwd:               config.C.DB.Pass,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", config.C.DB.Host, config.C.DB.Port),
		DBName:               config.C.DB.Name,
		AllowNativePasswords: true,
		ParseTime:            true,
		Loc:                  time.Local,
		Params:               map[string]string{"charset": "utf8"},
		Timeout:              time.Second * 5,
	}

	retry := 10
	for retry > 0 {
		var err error

		db, err = gorm.Open(gmysql.Open(mySQLConfig.FormatDSN()), gormCfg)
		if err != nil {
			logrus.Errorf("open db connection failure %s, will retry", err.Error())
			time.Sleep(time.Second * 3)
			retry--
			continue
		}
		break
	}
	gdb = db
	return db
}

// GetGDB -
func GetGDB() *gorm.DB {
	return gdb
}

// AutoMigrate run auto migration for given models
func AutoMigrate(db *gorm.DB) error {
	models := map[string]interface{}{
		"CloudAccessKey":       model.CloudAccessKey{},
		"CreateKubernetesTask": model.CreateKubernetesTask{},
		"InitWutongTask":       model.InitWutongTask{},
		"RKECluster":           model.RKECluster{},
		"CustomCluster":        model.CustomCluster{},
		"UpdateKubernetesTask": model.UpdateKubernetesTask{},
		"WutongClusterConfig":  model.WutongClusterConfig{},
		"AppStore":             model.AppStore{},
		"TaskEvent":            model.TaskEvent{},
	}

	for name, mod := range models {
		if err := db.AutoMigrate(mod); err != nil {
			return fmt.Errorf("auto migrate %s: %v", name, err)
		}
	}

	return nil
}
