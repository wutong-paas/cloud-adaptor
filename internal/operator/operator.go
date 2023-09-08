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

package operator

import (
	"context"
	"fmt"
	"time"

	wutongv1alpha1 "github.com/wutong-paas/wutong-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Operator operator
type Operator struct {
	Config
}

// Config operator config
type Config struct {
	WutongVersion         string
	Namespace             string
	ArchiveFilePath       string
	RuntimeClient         client.Client
	Wutongpackage         string
	WutongImageRepository string
	ImageHubUser          string
	ImageHubPass          string
	OnlyInstallRegion     bool
	// RegionArch            string
}

// NewOperator new operator
func NewOperator(c Config) (*Operator, error) {
	if c.RuntimeClient == nil {
		return nil, fmt.Errorf("config runtime client can not be nil")
	}
	return &Operator{
		Config: c,
	}, nil
}

// Install install
func (o *Operator) Install(cluster *wutongv1alpha1.WutongCluster) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	if err := o.RuntimeClient.Create(ctx, cluster); err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("create wutong cluster failure %s", err.Error())
		}
		var old wutongv1alpha1.WutongCluster
		if err := o.RuntimeClient.Get(ctx, types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, &old); err != nil {
			return fmt.Errorf("get wutong cluster failure %s", err.Error())
		}

		// Keep the image configuration
		if cluster.Spec.ImageHub == nil && old.Spec.ImageHub != nil {
			cluster.Spec.ImageHub = old.Spec.ImageHub
		}

		// Keep the database configuration
		if cluster.Spec.RegionDatabase == nil && old.Spec.RegionDatabase != nil {
			cluster.Spec.RegionDatabase = old.Spec.RegionDatabase
		}
		if cluster.Spec.UIDatabase == nil && old.Spec.UIDatabase != nil {
			cluster.Spec.UIDatabase = old.Spec.UIDatabase
		}
		old.Spec = cluster.Spec
		if err := o.RuntimeClient.Update(ctx, &old); err != nil {
			return fmt.Errorf("update wutong cluster failure %s", err.Error())
		}
		*cluster = old
	}
	return nil
}
