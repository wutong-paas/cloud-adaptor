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
	"encoding/base64"
	"strings"

	"github.com/pkg/errors"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/sirupsen/logrus"
	"github.com/wutong-paas/cloud-adaptor/internal/domain"
	"github.com/wutong-paas/cloud-adaptor/internal/repo/appstore"
	"github.com/wutong-paas/cloud-adaptor/pkg/bcode"
	"gopkg.in/yaml.v2"
)

type questions struct {
	Questions []v3.Question `json:"questions,omitempty"`
}

// TemplateVersionRepo -
type TemplateVersionRepo interface {
	// returns the specified version fo the app template.
	GetTemplateVersion(appStore *domain.AppStore, templateName, version string) (*domain.AppTemplateVersion, error)
}

// NewTemplateVersionRepo creates a new template version.
func NewTemplateVersionRepo(templateVersioner *appstore.TemplateVersioner) TemplateVersionRepo {
	return &templateVersionRepo{
		templateVersioner: templateVersioner,
	}
}

type templateVersionRepo struct {
	templateVersioner *appstore.TemplateVersioner
}

func (t *templateVersionRepo) GetTemplateVersion(appStore *domain.AppStore, templateName, version string) (*domain.AppTemplateVersion, error) {
	chart, err := t.templateVersioner.LoadChart(appStore.Name, appStore.URL, appStore.Username, appStore.Password, templateName, version)
	if err != nil {
		if strings.Contains(errors.Cause(err).Error(), "improper constraint: ") ||
			strings.Contains(errors.Cause(err).Error(), "no chart version found for") {
			return nil, errors.Wrap(bcode.ErrTemplateVersionNotFound, err.Error())
		}
		return nil, err
	}

	templateVersion := &domain.AppTemplateVersion{
		Values: make(map[string]string),
	}

	for _, file := range chart.Files {
		if file.Name == "README.md" {
			templateVersion.Readme = base64.StdEncoding.EncodeToString(file.Data)
		}
		if file.Name == "questions.yaml" || file.Name == "questions.yml" {
			var questions questions
			if err := yaml.Unmarshal(file.Data, &questions); err != nil {
				logrus.Warningf("unmarshal questions data: %v", err)
				continue
			}

			templateVersion.Questions = questions.Questions
		}
	}

	for i := len(chart.Raw) - 1; i >= 0; i-- {
		file := chart.Raw[i]
		if !strings.HasSuffix(file.Name, "values.yaml") && !strings.HasSuffix(file.Name, "values.yml") {
			continue
		}
		templateVersion.Values[file.Name] = base64.StdEncoding.EncodeToString(file.Data)
	}

	return templateVersion, nil
}
