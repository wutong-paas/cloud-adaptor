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

//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/wutong-paas/cloud-adaptor/cmd/cloud-adaptor/config"
	"github.com/wutong-paas/cloud-adaptor/internal/handler"
	"github.com/wutong-paas/cloud-adaptor/internal/middleware"
	"github.com/wutong-paas/cloud-adaptor/internal/nsqc"
	"github.com/wutong-paas/cloud-adaptor/internal/repo"
	"github.com/wutong-paas/cloud-adaptor/internal/repo/dao"
	"github.com/wutong-paas/cloud-adaptor/internal/task"
	"github.com/wutong-paas/cloud-adaptor/internal/types"
	"github.com/wutong-paas/cloud-adaptor/internal/usecase"
	"gorm.io/gorm"
)

// initApp init the application.
func initApp(context.Context,
	*gorm.DB,
	*config.Config,
	chan types.KubernetesConfigMessage,
	chan types.InitWutongConfigMessage,
	chan types.UpdateKubernetesConfigMessage) (*gin.Engine, error) {
	panic(wire.Build(handler.ProviderSet, usecase.ProviderSet, repo.ProviderSet, task.ProviderSet,
		nsqc.ProviderSet, dao.ProviderSet, middleware.ProviderSet, newApp))
}
