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

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	"github.com/wutong-paas/cloud-adaptor/cmd/cloud-adaptor/config"
	"github.com/wutong-paas/cloud-adaptor/internal/datastore"
	"github.com/wutong-paas/cloud-adaptor/internal/handler"
	"github.com/wutong-paas/cloud-adaptor/internal/nsqc"
	"github.com/wutong-paas/cloud-adaptor/internal/task"
	"github.com/wutong-paas/cloud-adaptor/internal/types"

	// Import all dependent packages in main.go for swag to generate doc.
	// More detail: https://github.com/swaggo/swag/issues/817#issuecomment-730895033
	_ "github.com/helm/helm/pkg/repo"
	_ "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	_ "github.com/wutong-paas/cloud-adaptor/api/cloud-adaptor/v1"
	_ "k8s.io/api/core/v1"
	_ "k8s.io/apimachinery/pkg/api/resource"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/apimachinery/pkg/types"
	_ "k8s.io/apimachinery/pkg/util/intstr"
	_ "k8s.io/helm/pkg/proto/hapi/chart"
)

// @contact.name Wutong
// @contact.url https://wutong.com

// @title Cloud Adaptor API
// @description Cloud Adaptor
// @version 1.0
// @BasePath /api/v1
// @schemes http, https
func main() {
	app := &cli.App{
		Name:  "cloud adapter",
		Usage: "run cloud adaptor server",
		Flags: append([]cli.Flag{
			&cli.BoolFlag{
				Name:  "testMode",
				Value: false,
				Usage: "A trigger to enable test mode.",
			},
			&cli.StringFlag{
				Name:    "logLevel",
				Value:   "debug",
				Usage:   "The level of logger.",
				EnvVars: []string{"LOG_LEVEL"},
			},
			&cli.StringFlag{
				Name:    "helm-repo-file",
				Value:   "/app/data/helm/repositories.yaml",
				Usage:   "path to the file containing repository names and URLs",
				EnvVars: []string{"HELM_REPO_FILE"},
			},
			&cli.StringFlag{
				Name:    "helm-cache",
				Value:   "/app/data/helm/cache",
				Usage:   "path to the file containing cached repository indexes",
				EnvVars: []string{"HELM_CACHE"},
			},
			&cli.StringFlag{
				Name:    "nsqd-server",
				Aliases: []string{"nsqd"},
				Value:   "127.0.0.1:4150",
				Usage:   "nsqd server address",
			},
			&cli.StringFlag{
				Name:    "nsq-lookupd-server",
				Aliases: []string{"lookupd"},
				Value:   "127.0.0.1:4161",
				Usage:   "nsq lookupd server address",
			},
			&cli.StringFlag{
				Name:    "listen",
				Aliases: []string{"l"},
				Value:   "127.0.0.1:8080",
				Usage:   "daemon server listen address",
				EnvVars: []string{"LISTEN"},
			},
		}, dbInfoFlag...),
		Action: run,
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Errorf("run cloud-adapter: %+v", err)
		os.Exit(1)
	}
}

func run(c *cli.Context) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	config.Parse(c)
	config.SetLogLevel()

	db := datastore.NewDB()
	if err := datastore.AutoMigrate(db); err != nil {
		return err
	}

	createChan := make(chan types.KubernetesConfigMessage, 10)
	initChan := make(chan types.InitWutongConfigMessage, 10)
	updateChan := make(chan types.UpdateKubernetesConfigMessage, 10)

	engine, err := initApp(ctx, db, config.C, createChan, initChan, updateChan)
	if err != nil {
		return err
	}

	logrus.Infof("start listen %s", c.String("listen"))
	go engine.Run(c.String("listen"))

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	select {
	case <-term:
		logrus.Warn("Received SIGTERM, exiting gracefully...")
	}
	logrus.Info("See you next time!")
	return nil
}

func newApp(ctx context.Context,
	router *handler.Router,
	createQueue chan types.KubernetesConfigMessage,
	initQueue chan types.InitWutongConfigMessage,
	updateQueue chan types.UpdateKubernetesConfigMessage,
	createHandler task.CreateKubernetesTaskHandler,
	initHandler task.CloudInitTaskHandler,
	cloudUpdateTaskHandler task.UpdateKubernetesTaskHandler) *gin.Engine {
	engine := router.NewRouter()
	engine.Use(gin.Recovery())

	msgConsumer := nsqc.NewTaskChannelConsumer(ctx, createQueue, initQueue, updateQueue, createHandler, initHandler, cloudUpdateTaskHandler)
	go msgConsumer.Start()

	return engine
}
