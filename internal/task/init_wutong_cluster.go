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

package task

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/rancher/rke/k8s"
	"github.com/sirupsen/logrus"
	apiv1 "github.com/wutong-paas/cloud-adaptor/api/cloud-adaptor/v1"
	"github.com/wutong-paas/cloud-adaptor/internal/adaptor/factory"
	"github.com/wutong-paas/cloud-adaptor/internal/datastore"
	"github.com/wutong-paas/cloud-adaptor/internal/operator"
	"github.com/wutong-paas/cloud-adaptor/internal/repo"
	"github.com/wutong-paas/cloud-adaptor/internal/types"
	"github.com/wutong-paas/cloud-adaptor/internal/usecase"
	"github.com/wutong-paas/cloud-adaptor/pkg/util/constants"
	"github.com/wutong-paas/cloud-adaptor/pkg/util/versionutil"
	"github.com/wutong-paas/cloud-adaptor/version"
	wutongv1alpha1 "github.com/wutong-paas/wutong-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//InitWutongCluster init wutong cluster
type InitWutongCluster struct {
	config *types.InitWutongConfig
	result chan apiv1.Message
}

func (c *InitWutongCluster) rollback(step, message, status string) {
	if status == "failure" {
		logrus.Errorf("%s failure, Message: %s", step, message)
	}
	c.result <- apiv1.Message{StepType: step, Message: message, Status: status}
}

//Run run take time 214.10s
func (c *InitWutongCluster) Run(ctx context.Context) {
	defer c.rollback("Close", "", "")
	c.rollback("Init", "", "start")
	// create adaptor
	adaptor, err := factory.GetCloudFactory().GetWutongClusterAdaptor(c.config.Provider, c.config.AccessKey, c.config.SecretKey)
	if err != nil {
		c.rollback("Init", fmt.Sprintf("create cloud adaptor failure %s", err.Error()), "failure")
		return
	}

	c.rollback("Init", "cloud adaptor create success", "success")
	c.rollback("CheckCluster", "", "start")
	// get kubernetes cluster info
	cluster, err := adaptor.DescribeCluster(c.config.EnterpriseID, c.config.ClusterID)
	if err != nil {
		cluster, err = adaptor.DescribeCluster(c.config.EnterpriseID, c.config.ClusterID)
		if err != nil {
			c.rollback("CheckCluster", err.Error(), "failure")
			return
		}
	}
	// check cluster status
	if cluster.State != "running" {
		c.rollback("CheckCluster", fmt.Sprintf("cluster status is %s,not support init wutong", cluster.State), "failure")
		return
	}
	// check cluster version
	if !versionutil.CheckVersion(cluster.KubernetesVersion) {
		c.rollback("CheckCluster", fmt.Sprintf("current cluster version is %s, init wutong support kubernetes version is 1.16.x-1.22.x", cluster.KubernetesVersion), "failure")
		return
	}
	// check cluster connection status
	logrus.Infof("init kubernetes url %s", cluster.MasterURL)
	if cluster.MasterURL.APIServerEndpoint == "" {
		c.rollback("CheckCluster", "cluster api not open eip,not support init wutong", "failure")
		return
	}

	kubeConfig, err := adaptor.GetKubeConfig(c.config.EnterpriseID, c.config.ClusterID)
	if err != nil {
		kubeConfig, err = adaptor.GetKubeConfig(c.config.EnterpriseID, c.config.ClusterID)
		if err != nil {
			c.rollback("CheckCluster", fmt.Sprintf("get kube config failure %s", err.Error()), "failure")
			return
		}
	}

	// check cluster not init wutong
	coreClient, _, err := kubeConfig.GetKubeClient()
	if err != nil {
		c.rollback("CheckCluster", fmt.Sprintf("get kube config failure %s", err.Error()), "failure")
		return
	}

	// get cluster node lists
	getctx, cancel := context.WithTimeout(ctx, time.Second*10)
	nodes, err := coreClient.CoreV1().Nodes().List(getctx, metav1.ListOptions{})
	if err != nil {
		nodes, err = coreClient.CoreV1().Nodes().List(getctx, metav1.ListOptions{})
		cancel()
		if err != nil {
			logrus.Errorf("get kubernetes cluster node failure %s", err.Error())
			c.rollback("CheckCluster", "cluster node list can not found, please check cluster public access and account authorization", "failure")
			return
		}
	} else {
		cancel()
	}
	if len(nodes.Items) == 0 {
		c.rollback("CheckCluster", "node num is 0, can not init wutong", "failure")
		return
	}
	c.rollback("CheckCluster", c.config.ClusterID, "success")

	// select gateway and chaos node
	gatewayNodes, chaosNodes := c.GetWutongGatewayNodeAndChaosNodes(nodes.Items)
	initConfig := adaptor.GetWutongInitConfig(c.config.EnterpriseID, cluster, gatewayNodes, chaosNodes, c.rollback)
	initConfig.WutongVersion = version.WutongRegionVersion
	// init wutong
	c.rollback("InitWutongRegionOperator", "", "start")
	if len(initConfig.EIPs) == 0 {
		c.rollback("InitWutongRegionOperator", "can not select eip", "failure")
		return
	}

	rri := operator.NewWutongRegionInit(*kubeConfig, repo.NewWutongClusterConfigRepo(datastore.GetGDB()))
	if err := rri.InitWutongRegion(initConfig); err != nil {
		c.rollback("InitWutongRegionOperator", err.Error(), "failure")
		return
	}
	ticker := time.NewTicker(time.Second * 5)
	timer := time.NewTimer(time.Minute * 60)
	defer timer.Stop()
	defer ticker.Stop()
	var operatorMessage, imageHubMessage, packageMessage, apiReadyMessage bool
	for {
		select {
		case <-ctx.Done():
			c.rollback("InitWutongRegion", "context cancel", "failure")
			return
		case <-ticker.C:
		case <-timer.C:
			c.rollback("InitWutongRegion", "waiting wutong region ready timeout", "failure")
			return
		}
		status, err := rri.GetWutongRegionStatus(initConfig.ClusterID)
		if err != nil {
			if k8sErrors.IsNotFound(err) {
				c.rollback("InitWutongRegion", err.Error(), "failure")
				return
			}
			logrus.Errorf("get wutong region status failure %s", err.Error())
		}
		if status == nil {
			continue
		}
		if status.OperatorReady && !operatorMessage {
			c.rollback("InitWutongRegionOperator", "", "success")
			c.rollback("InitWutongRegionImageHub", "", "start")
			operatorMessage = true
			continue
		}

		if idx, condition := status.WutongCluster.Status.GetCondition(wutongv1alpha1.WutongClusterConditionTypeImageRepository); !imageHubMessage && idx != -1 && condition.Status == v1.ConditionTrue {
			c.rollback("InitWutongRegionImageHub", "", "success")
			c.rollback("InitWutongRegionPackage", "", "start")
			imageHubMessage = true
			continue
		}
		statusStr := fmt.Sprintf("Push Images:%d/%d\t", len(status.WutongPackage.Status.ImagesPushed), status.WutongPackage.Status.ImagesNumber)
		for _, con := range status.WutongCluster.Status.Conditions {
			if con.Status == v1.ConditionTrue {
				statusStr += fmt.Sprintf("%s=>%s;\t", con.Type, con.Status)
			} else {
				statusStr += fmt.Sprintf("%s=>%s=>%s=>%s;\t", con.Type, con.Status, con.Reason, con.Message)
			}
		}
		logrus.Infof("cluster %s states: %s", cluster.Name, statusStr)

		for _, con := range status.WutongPackage.Status.Conditions {
			if con.Type == wutongv1alpha1.Ready && con.Status == wutongv1alpha1.Completed && !packageMessage {
				c.rollback("InitWutongRegionPackage", "", "success")
				c.rollback("InitWutongRegionRegionConfig", "", "start")
				packageMessage = true
			}
			continue
		}

		idx, condition := status.WutongCluster.Status.GetCondition(wutongv1alpha1.WutongClusterConditionTypeRunning)
		if idx != -1 && condition.Status == v1.ConditionTrue && packageMessage && !apiReadyMessage {
			apiReadyMessage = true
			break
		}
	}
	c.rollback("InitWutongRegion", cluster.ClusterID, "success")
}

//GetWutongGatewayNodeAndChaosNodes get gateway nodes
func (c *InitWutongCluster) GetWutongGatewayNodeAndChaosNodes(nodes []v1.Node) (gatewayNodes, chaosNodes []*wutongv1alpha1.K8sNode) {
	for _, node := range nodes {
		if node.Annotations["wutong.io/gateway-node"] == "true" {
			gatewayNodes = append(gatewayNodes, getK8sNode(node))
		}
		if node.Annotations["wutong.io/chaos-node"] == "true" {
			chaosNodes = append(chaosNodes, getK8sNode(node))
		}
	}
	if len(gatewayNodes) == 0 {
		if len(nodes) < 2 {
			gatewayNodes = []*wutongv1alpha1.K8sNode{
				getK8sNode(nodes[0]),
			}
		} else {
			gatewayNodes = []*wutongv1alpha1.K8sNode{
				getK8sNode(nodes[0]),
				getK8sNode(nodes[1]),
			}
		}
	}
	if len(chaosNodes) == 0 {
		if len(nodes) < 2 {
			chaosNodes = []*wutongv1alpha1.K8sNode{
				getK8sNode(nodes[0]),
			}
		} else {
			chaosNodes = []*wutongv1alpha1.K8sNode{
				getK8sNode(nodes[0]),
				getK8sNode(nodes[1]),
			}
		}
	}
	return
}

// Stop init
func (c *InitWutongCluster) Stop() error {
	return nil
}

//GetChan get message chan
func (c *InitWutongCluster) GetChan() chan apiv1.Message {
	return c.result
}

func getK8sNode(node v1.Node) *wutongv1alpha1.K8sNode {
	var Knode wutongv1alpha1.K8sNode
	for _, address := range node.Status.Addresses {
		if address.Type == v1.NodeInternalIP {
			Knode.InternalIP = address.Address
		}
		if address.Type == v1.NodeExternalIP {
			Knode.ExternalIP = address.Address
		}
		if address.Type == v1.NodeHostName {
			Knode.Name = address.Address
		}
	}
	if externamAddress, exist := node.Annotations[k8s.ExternalAddressAnnotation]; exist && externamAddress != "" {
		logrus.Infof("set node %s externalIP %s by %s", node.Name, externamAddress, k8s.ExternalAddressAnnotation)
		Knode.ExternalIP = externamAddress
	}
	return &Knode
}

//cloudInitTaskHandler cloud init task handler
type cloudInitTaskHandler struct {
	eventHandler *CallBackEvent
	handledTask  map[string]string
}

// NewCloudInitTaskHandler -
func NewCloudInitTaskHandler(clusterUsecase *usecase.ClusterUsecase) CloudInitTaskHandler {
	return &cloudInitTaskHandler{
		eventHandler: &CallBackEvent{TopicName: constants.CloudInit, ClusterUsecase: clusterUsecase},
		handledTask:  make(map[string]string),
	}
}

// HandleMsg -
func (h *cloudInitTaskHandler) HandleMsg(ctx context.Context, initConfig types.InitWutongConfigMessage) error {
	if _, exist := h.handledTask[initConfig.TaskID]; exist {
		logrus.Infof("task %s is running or complete,ignore", initConfig.TaskID)
		return nil
	}
	initTask, err := CreateTask(InitWutongClusterTask, initConfig.InitWutongConfig)
	if err != nil {
		logrus.Errorf("create task failure %s", err.Error())
		h.eventHandler.HandleEvent(initConfig.GetEvent(&apiv1.Message{
			StepType: "CreateTask",
			Message:  err.Error(),
			Status:   "failure",
		}))
		return nil
	}
	// Asynchronous execution to prevent message consumption from taking too long.
	// Idempotent consumption of messages is not currently supported
	go h.run(ctx, initTask, initConfig)
	h.handledTask[initConfig.TaskID] = "running"
	return nil
}

// HandleMessage implements the Handler interface.
// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
func (h *cloudInitTaskHandler) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		return nil
	}
	var initConfig types.InitWutongConfigMessage
	if err := json.Unmarshal(m.Body, &initConfig); err != nil {
		logrus.Errorf("unmarshal init wutong config message failure %s", err.Error())
		return nil
	}
	if err := h.HandleMsg(context.Background(), initConfig); err != nil {
		logrus.Errorf("handle init wutong config message failure %s", err.Error())
		return nil
	}
	return nil
}

func (h *cloudInitTaskHandler) run(ctx context.Context, initTask Task, initConfig types.InitWutongConfigMessage) {
	defer func() {
		h.handledTask[initConfig.TaskID] = "complete"
	}()
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()
	closeChan := make(chan struct{})
	go func() {
		defer close(closeChan)
		for message := range initTask.GetChan() {
			if message.StepType == "Close" {
				return
			}
			h.eventHandler.HandleEvent(initConfig.GetEvent(&message))
		}
	}()
	initTask.Run(ctx)
	//waiting message handle complete
	<-closeChan
	logrus.Infof("init wutong region task %s handle success", initConfig.TaskID)
}
