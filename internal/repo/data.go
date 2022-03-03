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

package repo

import (
	"github.com/google/wire"
	"github.com/wutong-paas/cloud-adaptor/internal/repo/appstore"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewCloudAccessKeyRepo,
	NewCreateKubernetesTaskRepo,
	NewInitWutongRegionTaskRepo,
	NewUpdateKubernetesTaskRepo,
	NewTaskEventRepo,
	NewWutongClusterConfigRepo,
	NewAppStoreRepo,
	NewRKEClusterRepo,
	NewCustomClusterRepository,
	NewTemplateVersionRepo,
	appstore.NewStorer,
	appstore.NewAppTemplater,
	appstore.NewTemplateVersioner,
)
