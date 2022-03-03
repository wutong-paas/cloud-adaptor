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

package producer

import (
	"encoding/json"
	"time"

	nsq "github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
	"github.com/wutong-paas/cloud-adaptor/cmd/cloud-adaptor/config"
	"github.com/wutong-paas/cloud-adaptor/internal/types"
	"github.com/wutong-paas/cloud-adaptor/pkg/util/constants"
)

//TaskProducer task producer
type TaskProducer interface {
	Start() error
	SendCreateKuerbetesTask(config types.KubernetesConfigMessage) error
	SendUpdateKuerbetesTask(config types.UpdateKubernetesConfigMessage) error
	SendInitWutongRegionTask(config types.InitWutongConfigMessage) error
	Stop()
}

//TaskProducer task producer
type taskProducer struct {
	taskProducer *nsq.Producer
}

//NewTaskProducer new task producer
func NewTaskProducer() TaskProducer {
	return &taskProducer{}
}

//Start start
func (m *taskProducer) Start() error {
	c := nsq.NewConfig()
	taskProducer, err := nsq.NewProducer(config.C.NSQConfig.NsqdAddress, c)
	if err != nil {
		return err
	}
	for {
		if err := taskProducer.Ping(); err != nil {
			logrus.Errorf("ping nsqd server failure %s", err.Error())
			time.Sleep(time.Second * 3)
			continue
		}
		logrus.Infof("ping nsqd server success")
		break
	}
	m.taskProducer = taskProducer
	logrus.Infof("task producer start success")
	return nil
}

//SendTask send task
func (m *taskProducer) sendTask(topicName string, taskConfig interface{}) error {
	body, err := json.Marshal(taskConfig)
	if err != nil {
		return err
	}
	return m.taskProducer.Publish(topicName, body)
}

//SendCreateKuerbetesTask send create kubernetes task
func (m *taskProducer) SendCreateKuerbetesTask(config types.KubernetesConfigMessage) error {
	return m.sendTask(constants.CloudCreate, config)
}

//SendInitWutongRegionTask send init wutong region task
func (m *taskProducer) SendInitWutongRegionTask(config types.InitWutongConfigMessage) error {
	return m.sendTask(constants.CloudInit, config)
}

func (m *taskProducer) SendUpdateKuerbetesTask(config types.UpdateKubernetesConfigMessage) error {
	return m.sendTask(constants.CloudUpdate, config)
}

//Stop stop
func (m *taskProducer) Stop() {
	m.taskProducer.Stop()
}
