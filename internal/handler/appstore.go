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

package handler

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/wutong-paas/cloud-adaptor/api/cloud-adaptor/v1"
	"github.com/wutong-paas/cloud-adaptor/internal/domain"
	"github.com/wutong-paas/cloud-adaptor/internal/usecase"
	"github.com/wutong-paas/cloud-adaptor/pkg/util/ginutil"
)

// AppStoreHandler -
type AppStoreHandler struct {
	appStore    *usecase.AppStoreUsecase
	appTemplate *usecase.AppTemplate
}

// NewAppStoreHandler
func NewAppStoreHandler(appStore *usecase.AppStoreUsecase,
	appTemplate *usecase.AppTemplate) *AppStoreHandler {
	return &AppStoreHandler{
		appStore:    appStore,
		appTemplate: appTemplate,
	}
}

// Create creates a new app store.
// @Summary creates a new app store.
// @Tags appstores
// @ID createAppStore
// @Accept  json
// @Produce  json
// @Param createAppStoreReq body v1.CreateAppStoreReq true "."
// @Success 200 {object} v1.AppStore
// @Failure 400 {object} ginutil.Result "8002, app store unavailable"
// @Failure 409 {object} ginutil.Result "8001, app store name conflict"
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/appstores [post]
func (a *AppStoreHandler) Create(c *gin.Context) {
	var req v1.CreateAppStoreReq
	if err := ginutil.ShouldBindJSON(c, &req); err != nil {
		ginutil.Error(c, err)
		return
	}

	// DTO to DO
	// TODO: Code generation or reflection
	appStore := &domain.AppStore{
		Name:     req.Name,
		URL:      req.URL,
		Branch:   req.Branch,
		Username: req.Username,
		Password: req.Password,
	}
	err := a.appStore.Create(c.Request.Context(), appStore)

	ginutil.JSON(c, &v1.AppStore{
		Name:     appStore.Name,
		URL:      appStore.URL,
		Branch:   appStore.Branch,
		Username: appStore.Username,
		Password: appStore.Password,
	}, err)
}

// List returns a list of app stores.
// @Summary returns a list of app stores.
// @Tags appstores
// @ID listAppStores
// @Accept  json
// @Produce  json
// @Success 200 {array} v1.AppStore
// @Router /api/v1/appstores [get]
func (a *AppStoreHandler) List(c *gin.Context) {
	appStores, err := a.appStore.List(c.Request.Context())

	var stores []*v1.AppStore
	for _, as := range appStores {
		stores = append(stores, &v1.AppStore{
			Name:     as.Name,
			URL:      as.URL,
			Branch:   as.Branch,
			Username: as.Username,
			Password: as.Password,
		})
	}

	ginutil.JSON(c, stores, err)
}

// Get returns the app store.
// @Summary returns the app store.
// @Tags appstores
// @ID getAppStore
// @Accept  json
// @Produce  json
// @Param name path string true "the name of the app store"
// @Success 200 {object} v1.AppStore
// @Failure 404 {object} ginutil.Result "8000, app store not found"
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/appstores/:name [get]
func (a *AppStoreHandler) Get(c *gin.Context) {
	appStore := ginutil.MustGet(c, "appStore").(*domain.AppStore)
	ginutil.JSON(c, appStore)
}

// Update updates the app store.
// @Summary updates the app store..
// @Tags appstores
// @ID updateAppStore
// @Accept  json
// @Produce  json
// @Param name path string true "the name of the app store"
// @Param updateAppStoreReq body v1.UpdateAppStoreReq true "."
// @Success 200 {object} v1.AppStore
// @Failure 404 {object} ginutil.Result "8000, app store not found"
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/appstores/:name [put]
func (a *AppStoreHandler) Update(c *gin.Context) {
	var req v1.UpdateAppStoreReq
	if err := ginutil.ShouldBindJSON(c, &req); err != nil {
		ginutil.Error(c, err)
		return
	}

	appStoreI, _ := c.Get("appStore")
	appStore := appStoreI.(*domain.AppStore)

	appStore.URL = req.URL
	appStore.Branch = req.Branch
	appStore.Username = req.Username
	appStore.Password = req.Password

	ginutil.JSON(c, nil, a.appStore.Update(c.Request.Context(), appStore))
}

// Delete deletes the app store.
// @Summary deletes the app store.
// @Tags appstores
// @ID deleteAppStore
// @Accept  json
// @Produce  json
// @Param name path string true "the name of the app store"
// @Success 200
// @Failure 404 {object} ginutil.Result "8000, app store not found"
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/appstores/:name [delete]
func (a *AppStoreHandler) Delete(c *gin.Context) {
	appStore := ginutil.MustGet(c, "appStore").(*domain.AppStore)
	ginutil.JSON(c, nil, a.appStore.Delete(c.Request.Context(), appStore))
}

// Resync resync the app templates.
// @Summary resync the app templates.
// @Tags appstores
// @ID resyncAppStore
// @Accept  json
// @Produce  json
// @Param name path string true "the name of the app store"
// @Success 200
// @Failure 404 {object} ginutil.Result "8000, app store not found"
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/appstores/:name/resync [post]
func (a *AppStoreHandler) Resync(c *gin.Context) {
	appStore := ginutil.MustGet(c, "appStore").(*domain.AppStore)
	a.appStore.Resync(c.Request.Context(), appStore)
}

// ListTemplates returns a list of app templates.
// @Summary returns a list of app templates.
// @Tags appstores
// @ID listTemplates
// @Accept  json
// @Produce  json
// @Param name path string true "the name of the app store"
// @Success 200 {array} v1.AppTemplate
// @Failure 404 {object} ginutil.Result "8000, app store not found"
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/appstores/:name/apps [get]
func (a *AppStoreHandler) ListTemplates(c *gin.Context) {
	appStore := ginutil.MustGet(c, "appStore").(*domain.AppStore)

	appTemplates := appStore.AppTemplates
	var templates []*v1.AppTemplate
	for _, at := range appTemplates {
		templates = append(templates, &v1.AppTemplate{
			Name:     at.Name,
			Versions: at.Versions,
		})
	}

	ginutil.JSON(c, templates)
}

// GetAppTemplate returns the app template.
// @Summary returns the app template.
// @Tags appstores
// @ID getAppTemplate
// @Accept  json
// @Produce  json
// @Param name path string true "the name of the app store"
// @Param templateName path string true "the name of the app template"
// @Success 200 {object} v1.AppTemplate
// @Failure 404 {object} ginutil.Result "8000, app store not found"
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/appstores/:name/apps/:templateName [get]
func (a *AppStoreHandler) GetAppTemplate(c *gin.Context) {
	appStore := ginutil.MustGet(c, "appStore").(*domain.AppStore)
	appTemplate, err := a.appStore.GetAppTemplate(c.Request.Context(), appStore, c.Param("templateName"))
	if err != nil {
		ginutil.Error(c, err)
		return
	}

	ginutil.JSON(c, &v1.AppTemplate{
		Name:     appTemplate.Name,
		Versions: appTemplate.Versions,
	})
}

// GetAppTemplateVersion returns the app template version.
// @Summary returns the app template version.
// @Tags appstores
// @ID getAppTemplateVersion
// @Accept  json
// @Produce  json
// @Param name path string true "the name of the app store"
// @Param templateName path string true "the name of the app template"
// @Param version path string true "the version of the app template"
// @Success 200 {object} v1.TemplateVersion
// @Failure 404 {object} ginutil.Result "8000, app store not found"
// @Failure 404 {object} ginutil.Result "8003, app template not found"
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/appstores/:name/templates/:templateName/versions/:version [get]
func (a *AppStoreHandler) GetAppTemplateVersion(c *gin.Context) {
	appStore := ginutil.MustGet(c, "appStore").(*domain.AppStore)
	version, err := a.appTemplate.GetVersion(c.Request.Context(), appStore, c.Param("templateName"), c.Param("version"))
	if err != nil {
		ginutil.Error(c, err)
		return
	}

	ginutil.JSON(c, &v1.TemplateVersion{
		Readme:    version.Readme,
		Questions: version.Questions,
		Values:    version.Values,
	})
}
