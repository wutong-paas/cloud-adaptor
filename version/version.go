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

package version

import (
	"os"
	"strings"
)

// WutongRegionVersion wutong region install version
var WutongRegionVersion = "v1.9.0"

// OperatorVersion operator image tag
var OperatorVersion = "v1.9.0"

// InstallImageRepo install image repo
var InstallImageRepo = "swr.cn-southwest-2.myhuaweicloud.com/wutong"

func init() {
	if os.Getenv("INSTALL_IMAGE_REPO") != "" {
		InstallImageRepo = os.Getenv("INSTALL_IMAGE_REPO")
	}
	if os.Getenv("WUTONG_VERSION") != "" {
		WutongRegionVersion = os.Getenv("WUTONG_VERSION")
	}
	if os.Getenv("OPERATOR_VERSION") != "" {
		OperatorVersion = os.Getenv("OPERATOR_VERSION")
	}
	InstallImageRepo = strings.TrimSuffix(InstallImageRepo, "/")
}
