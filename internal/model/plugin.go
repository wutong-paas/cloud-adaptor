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
package model

// Plugin -
type Plugin struct {
	Model
	PluginAlias string `gorm:"column:plugin_alias size:255"`
	Desc        string `gorm:"column:desc type:longtext"`
	Category    string `gorm:"column:category size:50"`
	Image       string `gorm:"column:image size:512"`
	Command     string `gorm:"column:command type:longtext"`
}

// PluginConfigGroup -
type PluginConfigGroup struct {
	Model
	PluginID        uint            `gorm:"column:plugin_id"`
	ConfigName      string          `gorm:"column:config_name size:255"`
	ServiceMetaType ServiceMetaType `gorm:"column:service_meta_type size:255"`
	Injection       InjectionType   `gorm:"column:injection size:255"`
}

// PluginConfigItem -
type PluginConfigItem struct {
	Model
	PluginConfigGroupID uint     `gorm:"column:plugin_config_group_id"`
	AttrName            string   `gorm:"column:attr_name size:255"`
	AttrType            AttrType `gorm:"column:attr_type size:255"`
	AttrAltValue        string   `gorm:"column:attr_alt_value size:255"`
	AttrDefaultValue    string   `gorm:"column:attr_default_value size:255"`
	IsChange            bool     `gorm:"column:is_change"`
	Protocol            string   `gorm:"column:protocol size:255"`
	AttrInfo            string   `gorm:"column:attr_info type:longtext"`
}

type (
	ServiceMetaType string
	InjectionType   string
	AttrType        string
)

const (
	SeviceMetaTypeUnDefine       ServiceMetaType = "un_define"
	SeviceMetaTypeUpStreamPort   ServiceMetaType = "upstream_port"
	SeviceMetaTypeDownStreamPort ServiceMetaType = "down_stream_port"

	InjectionTypeEnv  InjectionType = "env"
	InjectionTypeAuto InjectionType = "auto"

	AttrTypeString   AttrType = "string"
	AttrTypeRadio    AttrType = "radio"
	AttrTypeCheckBox AttrType = "checkbox"
)
