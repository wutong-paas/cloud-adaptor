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

package producer

import (
	"github.com/wutong-paas/cloud-adaptor/internal/types"
	"github.com/wutong-paas/cloud-adaptor/pkg/util/constants"
)

//TaskProducer task producer
type taskChannelProducer struct {
	createQueue chan types.KubernetesConfigMessage
	initQueue   chan types.InitWutongConfigMessage
	updateQueue chan types.UpdateKubernetesConfigMessage
}

//NewTaskChannelProducer new task channel producer
func NewTaskChannelProducer(createQueue chan types.KubernetesConfigMessage,
	initQueue chan types.InitWutongConfigMessage,
	updateQueue chan types.UpdateKubernetesConfigMessage) TaskProducer {
	return &taskChannelProducer{
		createQueue: createQueue,
		initQueue:   initQueue,
		updateQueue: updateQueue,
	}
}

//Start start
func (c *taskChannelProducer) Start() error {
	return nil
}

//SendTask send task
func (c *taskChannelProducer) sendTask(topicName string, taskConfig interface{}) error {
	if topicName == constants.CloudCreate {
		c.createQueue <- taskConfig.(types.KubernetesConfigMessage)
	}
	if topicName == constants.CloudInit {
		c.initQueue <- taskConfig.(types.InitWutongConfigMessage)
	}
	if topicName == constants.CloudUpdate {
		c.updateQueue <- taskConfig.(types.UpdateKubernetesConfigMessage)
	}
	return nil
}

//SendCreateKuerbetesTask send create kubernetes task
func (c *taskChannelProducer) SendCreateKuerbetesTask(config types.KubernetesConfigMessage) error {
	return c.sendTask(constants.CloudCreate, config)
}

//SendInitWutongRegionTask send init wutong region task
func (c *taskChannelProducer) SendInitWutongRegionTask(config types.InitWutongConfigMessage) error {
	return c.sendTask(constants.CloudInit, config)
}
func (c *taskChannelProducer) SendUpdateKuerbetesTask(config types.UpdateKubernetesConfigMessage) error {
	return c.sendTask(constants.CloudUpdate, config)
}

//Stop stop
func (c *taskChannelProducer) Stop() {

}
