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

package custom

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wutong-paas/cloud-adaptor/internal/adaptor"
	"github.com/wutong-paas/cloud-adaptor/internal/adaptor/v1alpha1"
	"github.com/wutong-paas/cloud-adaptor/internal/datastore"
	"github.com/wutong-paas/cloud-adaptor/internal/model"
	"github.com/wutong-paas/cloud-adaptor/internal/repo"
	"github.com/wutong-paas/cloud-adaptor/pkg/bcode"
	"github.com/wutong-paas/cloud-adaptor/pkg/util/versionutil"
	wutongv1alpha1 "github.com/wutong-paas/wutong-operator/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
)

type customAdaptor struct {
	Repo *repo.CustomClusterRepo
}

// Create create ack adaptor
func Create() (adaptor.WutongClusterAdaptor, error) {
	return &customAdaptor{
		Repo: repo.NewCustomClusterRepo(datastore.GetGDB()),
	}, nil
}

func (c *customAdaptor) ClusterList() ([]*v1alpha1.Cluster, error) {
	clusters, err := c.Repo.ListCluster()
	if err != nil {
		return nil, err
	}
	var re []*v1alpha1.Cluster
	var wait sync.WaitGroup
	for _, clu := range clusters {
		wait.Add(1)
		go func(clu *model.CustomCluster) {
			defer wait.Done()
			cluster, err := c.DescribeCluster(clu.ClusterID)
			if err != nil {
				logrus.Warningf("query kubernetes cluster failure %s", err.Error())
			}
			if cluster != nil {
				re = append(re, cluster)
			}
		}(clu)
	}
	wait.Wait()
	return re, nil
}

func (c *customAdaptor) DescribeCluster(clusterID string) (*v1alpha1.Cluster, error) {
	cc, err := c.Repo.GetCluster(clusterID)
	if err != nil {
		return nil, fmt.Errorf("query cluster meta info failure %s", err.Error())
	}
	cluster := &v1alpha1.Cluster{
		Name:        cc.Name,
		ClusterID:   cc.ClusterID,
		Created:     v1alpha1.NewTime(cc.CreatedAt),
		State:       v1alpha1.OfflineState,
		ClusterType: "custom",
		EIP: func() []string {
			if cc.EIP != "" {
				return strings.Split(cc.EIP, ",")
			}
			return nil
		}(),
		Parameters: make(map[string]interface{}),
	}
	kc := v1alpha1.KubeConfig{Config: cc.KubeConfig}
	client, _, err := kc.GetKubeClient()
	if err != nil {
		cluster.Parameters["DisableWutongInit"] = true
		cluster.Parameters["Message"] = "无法创建集群通信客户端"
		return cluster, fmt.Errorf("create kube client failure %s", err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	versionByte, err := client.RESTClient().Get().AbsPath("/version").DoRaw(ctx)
	if err != nil {
		cluster.Parameters["DisableWutongInit"] = true
		cluster.Parameters["Message"] = "无法直接与集群 KubeAPI 通信"
		return cluster, fmt.Errorf("get cluster version failure %s", err.Error())
	}
	var vinfo version.Info
	_ = json.Unmarshal(versionByte, &vinfo)
	cluster.KubernetesVersion = vinfo.String()
	cluster.CurrentVersion = vinfo.String()
	if !versionutil.CheckVersion(cluster.CurrentVersion) {
		cluster.Parameters["DisableWutongInit"] = true
		cluster.Parameters["Message"] = fmt.Sprintf("当前集群版本为 %s ，无法继续初始化，初始化Wutong支持的版本为1.16.x-1.22.x", cluster.CurrentVersion)
	}
	cluster.MasterURL.APIServerEndpoint, _ = kc.KubeServer()

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	nodes, err := client.CoreV1().Nodes().List(ctx, v1.ListOptions{})
	if err != nil {
		cluster.Parameters["DisableWutongInit"] = true
		cluster.Parameters["Message"] = "无法获取集群节点列表"
		return cluster, fmt.Errorf("query cluster node info failure %s", err.Error())
	}
	cluster.State = v1alpha1.RunningState
	cluster.Size = len(nodes.Items)
	_, err = client.CoreV1().ConfigMaps("wt-system").Get(ctx, "region-config", v1.GetOptions{})
	if err == nil {
		cluster.WutongInit = true
	}
	return cluster, nil
}

func (c *customAdaptor) GetKubeConfig(clusterID string) (*v1alpha1.KubeConfig, error) {
	cc, err := c.Repo.GetCluster(clusterID)
	if err != nil {
		return nil, fmt.Errorf("query cluster meta info failure %s", err.Error())
	}
	return &v1alpha1.KubeConfig{Config: cc.KubeConfig}, nil
}

// DeleteCluster delete cluster
func (c *customAdaptor) DeleteCluster(clusterID string) error {
	cluster, _ := c.DescribeCluster(clusterID)
	if cluster != nil && cluster.WutongInit {
		return bcode.ErrClusterNotAllowDelete
	}
	return c.Repo.DeleteCluster(clusterID)
}

func (c *customAdaptor) GetWutongInitConfig(cluster *v1alpha1.Cluster, gateway, chaos []*wutongv1alpha1.K8sNode, rollback func(step, message, status string)) *v1alpha1.WutongInitConfig {
	return &v1alpha1.WutongInitConfig{
		EnableHA:     cluster.Size > 3,
		ClusterID:    cluster.ClusterID,
		GatewayNodes: gateway,
		ChaosNodes:   chaos,
		EIPs: func() (re []string) {
			if len(cluster.EIP) > 0 {
				return cluster.EIP
			}
			for _, n := range gateway {
				if n.ExternalIP != "" {
					re = append(re, n.ExternalIP)
				}
			}
			if len(re) == 0 {
				for _, n := range gateway {
					if n.InternalIP != "" {
						re = append(re, n.InternalIP)
					}
				}
			}
			return
		}(),
	}
}

func (c *customAdaptor) CreateCluster(v1alpha1.CreateClusterConfig) (*v1alpha1.Cluster, error) {
	return nil, nil
}

func (c *customAdaptor) CreateWutongKubernetes(ctx context.Context, config *v1alpha1.KubernetesClusterConfig, rollback func(step, message, status string)) *v1alpha1.Cluster {
	rollback("CreateCluster", "", "success")
	return nil
}

func (c *customAdaptor) ExpansionNode(ctx context.Context, en *v1alpha1.ExpansionNode, rollback func(step, message, status string)) *v1alpha1.Cluster {
	return nil
}
