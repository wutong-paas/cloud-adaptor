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

package handler

import (
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "github.com/wutong-paas/cloud-adaptor/api/cloud-adaptor/v1"
	"github.com/wutong-paas/cloud-adaptor/internal/adaptor/v1alpha1"
	"github.com/wutong-paas/cloud-adaptor/internal/usecase"
	"github.com/wutong-paas/cloud-adaptor/pkg/bcode"
	"github.com/wutong-paas/cloud-adaptor/pkg/util/ginutil"
	"github.com/wutong-paas/cloud-adaptor/pkg/util/md5util"
)

// ClusterHandler -
type ClusterHandler struct {
	cluster *usecase.ClusterUsecase
}

// NewClusterHandler
func NewClusterHandler(clusterUsecase *usecase.ClusterUsecase) *ClusterHandler {
	return &ClusterHandler{
		cluster: clusterUsecase,
	}
}

// ListKubernetesClusters returns the information of .
//
// swagger:route GET /enterprise-server/api/v1/kclusters cloud kcluster
//
// # ListKubernetesCluster
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:KubernetesClustersResponse
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) ListKubernetesClusters(ctx *gin.Context) {
	var req v1.ListKubernetesCluster
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logrus.Errorf("bind query param failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	clusters, err := e.cluster.ListKubernetesCluster(req)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, v1.KubernetesClustersResponse{Clusters: clusters}, nil)
}

// AddKubernetesCluster returns the information of .
//
// swagger:route GET /enterprise-server/api/v1/kclusters cloud kcluster
//
// # CreateKubernetesReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:CreateKubernetesRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) AddKubernetesCluster(ctx *gin.Context) {
	var req v1.CreateKubernetesReq
	if err := ginutil.ShouldBindJSON(ctx, &req); err != nil {
		ginutil.Error(ctx, err)
		return
	}
	if req.Provider == "rke" {
		if req.EncodedRKEConfig == "" {
			ginutil.JSON(ctx, nil, bcode.ErrIncorrectRKEConfig)
			return
		}
	}
	if req.Provider == "custom" {
		if req.KubeConfig == "" {
			ginutil.JSON(ctx, nil, bcode.ErrKubeConfigCannotEmpty)
			return
		}
	}
	task, err := e.cluster.CreateKubernetesCluster(req)
	if err != nil {
		ginutil.JSON(ctx, task, err)
		return
	}
	ginutil.JSON(ctx, task, nil)
}

// UpdateKubernetesCluster updates kubernetes cluster.
//
// @Summary updates kubernetes cluster.
// @Tags cluster
// @ID updateKubernetesCluster
// @Accept  json
// @Produce  json
// @Param updateKubernetesReq body v1.UpdateKubernetesReq true "."
// @Success 200 {object} v1.UpdateKubernetesTask
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/update-cluster [post]
func (e *ClusterHandler) UpdateKubernetesCluster(ctx *gin.Context) {
	var req v1.UpdateKubernetesReq
	if err := ginutil.ShouldBindJSON(ctx, &req); err != nil {
		ginutil.Error(ctx, err)
		return
	}
	if req.Provider == "rke" {
		if req.EncodedRKEConfig == "" {
			ginutil.Error(ctx, errors.WithMessage(bcode.ErrIncorrectRKEConfig, "rke config is required"))
			return
		}
	}
	task, err := e.cluster.UpdateKubernetesCluster(req)
	if err != nil {
		ginutil.JSONv2(ctx, task, err)
		return
	}
	ginutil.JSONv2(ctx, task)
}

// GetUpdateKubernetesTask returns the information of the cluster.
//
// @Summary  returns the information of the cluster.
// @Tags cluster
// @ID getUpdateKubernetesTask
// @Accept  json
// @Produce  json
// @Param clusterID path string true "the cluster id"
// @Success 200 {object} v1.UpdateKubernetesTask
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/update-cluster/:clusterID [get]
func (e *ClusterHandler) GetUpdateKubernetesTask(ctx *gin.Context) {
	clusterID := ctx.Param("clusterID")
	providerName := ctx.Query("provider_name")
	re, err := e.cluster.GetUpdateKubernetesTask(clusterID, providerName)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}

	ginutil.JSON(ctx, re, nil)
}

// DeleteKubernetesCluster returns the information of .
//
// swagger:route GET /enterprise-server/api/v1kclusters/{clusterID} cloud kcluster
//
// # DeleteKubernetesClusterReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:Reponse
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) DeleteKubernetesCluster(ctx *gin.Context) {
	var req v1.DeleteKubernetesClusterReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logrus.Errorf("bind query param failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	clusterID := ctx.Param("clusterID")
	err := e.cluster.DeleteKubernetesCluster(clusterID, req.ProviderName)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, nil, nil)
}

// GetLastAddKubernetesClusterTask returns the information of .
//
// swagger:route GET /enterprise-server/api/v1last-ck-task cloud kcluster
//
// # GetLastCreateKubernetesClusterTaskReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:GetCreateKubernetesClusterTaskRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) GetLastAddKubernetesClusterTask(ctx *gin.Context) {
	var req v1.GetLastCreateKubernetesClusterTaskReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logrus.Errorf("bind query param failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	task, err := e.cluster.GetLastCreateKubernetesTask(req.ProviderName)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, task, nil)
}

// GetAddKubernetesClusterTask returns the information of .
//
// swagger:route GET /enterprise-server/api/v1ck-task/{taskID} cloud kcluster
//
// # GetLastCreateKubernetesClusterTaskReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:GetCreateKubernetesClusterTaskRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) GetAddKubernetesClusterTask(ctx *gin.Context) {
	taskID := ctx.Param("taskID")
	task, err := e.cluster.GetCreateKubernetesTask(taskID)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, task, nil)
}

// GetTaskEventList returns the information of .
//
// swagger:route GET /enterprise-server/api/v1ck-task/{taskID}/events cloud kcluster
//
// # GetTaskEventListReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:GetCreateKubernetesClusterTaskRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) GetTaskEventList(ctx *gin.Context) {
	taskID := ctx.Param("taskID")
	events, err := e.cluster.ListTaskEvent(taskID)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, v1.TaskEventListRes{Events: events}, nil)
}

// AddAccessKey add access keys
func (e *ClusterHandler) AddAccessKey(ctx *gin.Context) {
	var req v1.AddAccessKey
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logrus.Errorf("bind add accesskey param failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	clusters, err := e.cluster.AddAccessKey(req)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, clusters, nil)
}

// GetAccessKey add access keys
func (e *ClusterHandler) GetAccessKey(ctx *gin.Context) {
	var req v1.GetAccessKeyReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logrus.Errorf("bind add accesskey param failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	access, err := e.cluster.GetByProvider(req.ProviderName)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	access.SecretKey = md5util.Md5Crypt(access.SecretKey, "")
	ginutil.JSON(ctx, access, nil)
}

// GetInitWutongTask returns the information of .
//
// swagger:route GET /enterprise-server/api/v1init-task/{clusterID} cloud init
//
// # GetInitWutongTaskReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:InitWutongTaskRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) GetInitWutongTask(ctx *gin.Context) {
	clusterID := ctx.Param("clusterID")
	var req v1.GetInitWutongTaskReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logrus.Errorf("bind get init wutong task query failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	task, err := e.cluster.GetInitWutongTaskByClusterID(clusterID, req.ProviderName)
	ginutil.JSON(ctx, task, err)
}

// CreateInitWutongTask returns the information of .
//
// swagger:route POST /enterprise-server/api/v1init-cluster cloud init
//
// # InitWutongRegionReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:InitWutongTaskRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) CreateInitWutongTask(ctx *gin.Context) {
	var req v1.InitWutongRegionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logrus.Errorf("bind init wutong body failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	task, err := e.cluster.InitWutongRegion(ctx.Request.Context(), req)
	if err != nil {
		ginutil.JSON(ctx, task, err)
		return
	}
	ginutil.JSON(ctx, task, nil)
}

// GetRunningInitWutongTask returns the information of .
//
// swagger:route GET /enterprise-server/api/v1init-task/{clusterID} cloud init
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:InitWutongTaskListRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) GetRunningInitWutongTask(ctx *gin.Context) {
	tasks, err := e.cluster.GetTaskRunningLists()
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, v1.InitWutongTaskListRes{Tasks: tasks}, nil)
}

// GetRegionConfig get region config file
//
// swagger:route GET /enterprise-server/api/v1kclusters/{clusterID}/regionconfig cloud kcluster
//
// # GetRegionConfigReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:GetRegionConfigRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) GetRegionConfig(ctx *gin.Context) {
	var req v1.GetRegionConfigReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logrus.Errorf("bind get wutong region config failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	clusterID := ctx.Param("clusterID")
	configs, err := e.cluster.GetRegionConfig(clusterID, req.ProviderName)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	out, _ := yaml.Marshal(configs)
	ginutil.JSON(ctx, v1.GetRegionConfigRes{Configs: configs, ConfigYaml: string(out)}, nil)
}

// UpdateInitWutongTaskStatus get region config file
//
// swagger:route PUT /enterprise-server/api/v1init-tasks/{taskID}/status cloud init
//
// # UpdateInitWutongTaskStatusReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:InitWutongTaskRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) UpdateInitWutongTaskStatus(ctx *gin.Context) {
	var req v1.UpdateInitWutongTaskStatusReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logrus.Errorf("bind update init status failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	taskID := ctx.Param("taskID")
	task, err := e.cluster.UpdateInitWutongTaskStatus(taskID, req.Status)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, task, nil)
}

// GetInitNodeCmd get node init cmd shell
//
// swagger:route GET /enterprise-server/api/v1/init_node_cmd cloud init
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:InitNodeCmdRes
func (e *ClusterHandler) GetInitNodeCmd(c *gin.Context) {
	res, err := e.cluster.GetInitNodeCmd(c.Request.Context())
	ginutil.JSONv2(c, res, err)
}

// GetLogContent get rke create kubernetes log
//
// swagger:route GET /enterprise-server/api/v1kclusters/{clusterID}/create_log cloud init
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:GetLogContentRes
func (e *ClusterHandler) GetLogContent(ctx *gin.Context) {
	cluster, err := e.cluster.GetCluster("rke", ctx.Param("clusterID"))
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	var content []byte
	if cluster.CreateLogPath != "" {
		content, _ = ioutil.ReadFile(cluster.CreateLogPath)
	}
	ginutil.JSON(ctx, v1.GetLogContentRes{Content: string(content)}, nil)
}

// ReInstallKubernetesCluster retry install rke cluster .
//
// swagger:route GET /enterprise-server/api/v1kclusters/{clusterID}/reinstall cloud kcluster
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:CreateKubernetesRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) ReInstallKubernetesCluster(ctx *gin.Context) {
	task, err := e.cluster.InstallCluster(ctx.Param("clusterID"))
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, task, nil)
}

// GetKubeConfig get kubernetes cluster config
//
// swagger:route GET /enterprise-server/api/v1kclusters/{clusterID}/kubeconfig cloud init
//
// # GetRegionConfigReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:GetKubeConfigRes
func (e *ClusterHandler) GetKubeConfig(ctx *gin.Context) {
	var req v1.GetRegionConfigReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logrus.Errorf("bind get wutong region config failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	kubeconfig, err := e.cluster.GetKubeConfig(ctx.Param("clusterID"), req.ProviderName)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			ginutil.JSON(ctx, nil, bcode.NotFound)
		}
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, v1.GetKubeConfigRes{Config: kubeconfig}, nil)
}

// GetWutongClusterConfig -
func (e *ClusterHandler) GetWutongClusterConfig(ctx *gin.Context) {
	clusterID := ctx.Param("clusterID")
	_, config := e.cluster.GetWutongClusterConfig(clusterID)
	if config == "" {
		config = `
# apiVersion: wutong.io/v1alpha1
# kind: WutongCluster
# metadata:
#  name: wutongcluster
#  namespace: wt-system
# spec:
#  ## set source build cache mode, default is hostpath, options: pv, hostpath
#  arch: amd64
#  cacheMode: hostpath
#  configCompleted: true
#  ## Whether to deploy high availability. default is true if the number of nodes is greater than 3.
#  enableHA: false
#  ## etcd config, secret must have ca-file、cert-file and key-file keys.
#  etcdConfig:
#	 endpoints:
#	 - 192.168.10.6:2379
#	 - 192.168.10.8:2379
#	 - 192.168.10.4:2379
#	 secretName: wt-etcd-secret
#  ## Specifies the outer network IP address of the gateway. As the access address.
#  gatewayIngressIPs:
#    - 39.101.149.237
#  ## Specifies image hub info, deployment default hub when not set.
#  imageHub:
#	 domain: wutong.me
#	 password: 526856c5
#	 username: admin
#  installVersion: v1.0.0-stable
#  ## Specifies the node that performs the component CI task.
#  nodesForChaos:
#   - externalIP: 121.89.192.53
#	  internalIP: 192.168.10.3
#	  name: 39.101.149.237
#  ## Specify the gateway node.
#  nodesForGateway:
#   - externalIP: 121.89.192.53
#	  internalIP: 192.168.10.3
#	  name: 39.101.149.237
#  ## Specifies the wutong component image hub address
#  wutongImageRepository: swr.cn-southwest-2.myhuaweicloud.com/wutong
#  ## Specifies shared storage provider.
#  wutongVolumeSpecRWX:
#	 imageRepository: ""
#	 storageClassName: glusterfs-simple
#  ## Specifies the db connection info of region.
#  regionDatabase:
#	 host: 127.0.0.1
#	 name: region
#	 password: wutong123456!
#	 port: 3306
#	 username: root
#  ## Specifies the default component domain name suffix. Not specified will be assigned by default
#  suffixHTTPHost: xxxx.wtapps.cn`
	}
	re := v1.SetWutongClusterConfigReq{
		Config: config,
	}
	ginutil.JSON(ctx, re, nil)
}

// SetWutongClusterConfig -
func (e *ClusterHandler) SetWutongClusterConfig(ctx *gin.Context) {
	clusterID := ctx.Param("clusterID")
	var req v1.SetWutongClusterConfigReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logrus.Errorf("bind update init status failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	err := e.cluster.SetWutongClusterConfig(clusterID, req.Config)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, nil, nil)
}

// UninstallRegion -
func (e *ClusterHandler) UninstallRegion(ctx *gin.Context) {
	clusterID := ctx.Param("clusterID")
	var req v1.UninstallRegionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logrus.Errorf("bind update init status failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	err := e.cluster.UninstallWutongRegion(clusterID, req.ProviderName)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, nil, nil)
}

// @Summary update rke config purely
// @Tags cluster
// @ID pruneUpdateRKEConfig
// @Accept  json
// @Produce  json
// @Param pruneUpdateRKEConfigReq body v1.PruneUpdateRKEConfigReq true "."
// @Success 200 {object} v1.PruneUpdateRKEConfigResp
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/kclusters/prune-update-rkeconfig [POST]
func (e *ClusterHandler) pruneUpdateRKEConfig(c *gin.Context) {
	var req v1.PruneUpdateRKEConfigReq
	if err := ginutil.ShouldBindJSON(c, &req); err != nil {
		ginutil.Error(c, err)
		return
	}

	// clean invalid nodes
	var nodes v1alpha1.NodeList
	for _, node := range req.Nodes {
		if node.IP == "" || node.InternalAddress == "" {
			continue
		}
		nodes = append(nodes, node)
	}
	req.Nodes = nodes

	rkeConfig, err := e.cluster.PruneUpdateRKEConfig(&req)
	ginutil.JSONv2(c, rkeConfig, err)
}

// ListWutongComponents returns a list of wutong components.
// @Summary returns a list of wutong components.
// @Tags cluster
// @ID listWutongComponents
// @Accept  json
// @Produce  json
// @Param clusterID path string true "the identify of cluster"
// @Param providerName query string true "the provider of the cluster"
// @Success 200 {array} v1.WutongComponent
// @Router /api/v1kclusters/{clusterID}/wutong-components [get]
func (e *ClusterHandler) listWutongComponents(c *gin.Context) {
	clusterID := c.Param("clusterID")
	providerName := c.Query("providerName")
	components, err := e.cluster.ListWutongComponents(c.Request.Context(), clusterID, providerName)
	ginutil.JSONv2(c, components, err)
}

// listPodEvents returns a list of wutong component pod events.
// @Summary returns a list of wutong component pod events.
// @Tags cluster
// @ID listPodEvents
// @Accept  json
// @Produce  json
// @Param clusterID path string true "the identify of cluster"
// @Param podName path string true "the name of pod"
// @Param providerName query string true "the provider of the cluster"
// @Success 200 {array} v1.WutongComponentEvent
// @Router /api/v1kclusters/{clusterID}/wutong-components/{podName}/events [get]
func (e *ClusterHandler) listPodEvents(c *gin.Context) {
	clusterID := c.Param("clusterID")
	providerName := c.Query("providerName")
	components, err := e.cluster.ListPodEvents(c.Request.Context(), clusterID, providerName, c.Param("podName"))
	ginutil.JSONv2(c, components, err)
}
