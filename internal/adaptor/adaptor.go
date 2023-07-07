// WUTONG, Application Management Platform
// Copyright (C) 2014-2017 Wutong Co., Ltd.

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

package adaptor

import (
	"context"
	"fmt"

	"github.com/wutong-paas/cloud-adaptor/internal/adaptor/v1alpha1"
	wutongv1alpha1 "github.com/wutong-paas/wutong-operator/api/v1alpha1"
)

var (
	//ErrNotSupportRDS not support rds
	ErrNotSupportRDS = fmt.Errorf("not support rds")
)

// CloudAdaptor cloud adaptor interface
type CloudAdaptor interface {
	WutongClusterAdaptor
	VPCList(regionID string) ([]*v1alpha1.VPC, error)
	CreateVPC(v *v1alpha1.VPC) error
	DeleteVPC(regionID, vpcID string) error
	DescribeVPC(regionID, vpcID string) (*v1alpha1.VPC, error)
	CreateVSwitch(v *v1alpha1.VSwitch) error
	DescribeVSwitch(regionID, vswitchID string) (*v1alpha1.VSwitch, error)
	DeleteVSwitch(regionID, vswitchID string) error
	ListZones(regionID string) ([]*v1alpha1.Zone, error)
	ListInstanceType(regionID string) ([]*v1alpha1.InstanceType, error)
	CreateDB(*v1alpha1.Database) error
}

// KubernetesClusterAdaptor -
type KubernetesClusterAdaptor interface {
	ClusterList() ([]*v1alpha1.Cluster, error)
	DescribeCluster(clusterID string) (*v1alpha1.Cluster, error)
	CreateCluster(config v1alpha1.CreateClusterConfig) (*v1alpha1.Cluster, error)
	GetKubeConfig(clusterID string) (*v1alpha1.KubeConfig, error)
	DeleteCluster(clusterID string) error
	ExpansionNode(ctx context.Context, en *v1alpha1.ExpansionNode, rollback func(step, message, status string)) *v1alpha1.Cluster
}

// WutongClusterAdaptor wutong init adaptor
type WutongClusterAdaptor interface {
	KubernetesClusterAdaptor
	CreateWutongKubernetes(ctx context.Context, config *v1alpha1.KubernetesClusterConfig, rollback func(step, message, status string)) *v1alpha1.Cluster
	GetWutongInitConfig(cluster *v1alpha1.Cluster, gateway, chaos []*wutongv1alpha1.K8sNode, rollback func(step, message, status string)) *v1alpha1.WutongInitConfig
}
