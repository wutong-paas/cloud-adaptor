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

package domain

import (
	"github.com/pkg/errors"
	"github.com/wutong-paas/cloud-adaptor/pkg/bcode"
)

// AppStore -
type AppStore struct {
	Name         string
	URL          string
	Branch       string
	Username     string
	Password     string
	AppTemplates []*AppTemplate
}

// Key -
func (a *AppStore) Key() string {
	return a.Name
}

// Equals -
func (a *AppStore) Equals(b *AppStore) bool {
	if a.URL != b.URL {
		return false
	}
	if a.Branch != b.Branch {
		return false
	}
	if a.Username != b.Username {
		return false
	}
	if a.Password != b.Password {
		return false
	}
	return true
}

// GetAppTemplate get app template based on the app template name.
func (a *AppStore) GetAppTemplate(templateName string) (*AppTemplate, error) {
	for _, appTemplate := range a.AppTemplates {
		if appTemplate.Name == templateName {
			return appTemplate, nil
		}
	}
	return nil, errors.Wrap(bcode.ErrAppTemplateNotFound, "get app template")
}
