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

package operator

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/wutong-paas/wutong-operator/api/v1alpha1"
	wutongv1alpha1 "github.com/wutong-paas/wutong-operator/api/v1alpha1"
	"github.com/wutong-paas/wutong-operator/util/commonutil"
	"github.com/wutong-paas/wutong-operator/util/constants"
	"github.com/wutong-paas/wutong-operator/util/retryutil"
	"github.com/wutong-paas/wutong-operator/util/wtutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//Operator operator
type Operator struct {
	Config
}

//Config operator config
type Config struct {
	WutongVersion         string
	Namespace             string
	ArchiveFilePath       string
	RuntimeClient         client.Client
	Wutongpackage         string
	WutongImageRepository string
	ImageHubUser          string
	ImageHubPass          string
	OnlyInstallRegion     bool
}

//NewOperator new operator
func NewOperator(c Config) (*Operator, error) {
	if c.RuntimeClient == nil {
		return nil, fmt.Errorf("config runtime client can not be nil")
	}
	return &Operator{
		Config: c,
	}, nil
}

type componentClaim struct {
	namespace       string
	name            string
	version         string
	imageRepository string
	imageName       string
	Configs         map[string]string
	isInit          bool
	replicas        *int32
}

func (c *componentClaim) image() string {
	return path.Join(c.imageRepository, c.imageName) + ":" + c.version
}

func parseComponentClaim(claim *componentClaim) *v1alpha1.WutongComponent {
	component := &v1alpha1.WutongComponent{}
	component.Namespace = claim.namespace
	component.Name = claim.name
	component.Spec.Image = claim.image()
	component.Spec.ImagePullPolicy = corev1.PullIfNotPresent
	component.Spec.Replicas = claim.replicas
	labels := wtutil.LabelsForWutong(map[string]string{"name": claim.name})
	if claim.isInit {
		component.Spec.PriorityComponent = true
		labels["priorityComponent"] = "true"
	}
	component.Labels = labels
	return component
}

//Install install
func (o *Operator) Install(cluster *wutongv1alpha1.WutongCluster) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	if err := o.RuntimeClient.Create(ctx, cluster); err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("create wutong cluster failure %s", err.Error())
		}
		var old wutongv1alpha1.WutongCluster
		if err := o.RuntimeClient.Get(ctx, types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, &old); err != nil {
			return fmt.Errorf("get wutong cluster failure %s", err.Error())
		}

		// Keep the image configuration
		if cluster.Spec.ImageHub == nil && old.Spec.ImageHub != nil {
			cluster.Spec.ImageHub = old.Spec.ImageHub
		}

		// Keep the database configuration
		if cluster.Spec.RegionDatabase == nil && old.Spec.RegionDatabase != nil {
			cluster.Spec.RegionDatabase = old.Spec.RegionDatabase
		}
		if cluster.Spec.UIDatabase == nil && old.Spec.UIDatabase != nil {
			cluster.Spec.UIDatabase = old.Spec.UIDatabase
		}
		old.Spec = cluster.Spec
		if err := o.RuntimeClient.Update(ctx, &old); err != nil {
			return fmt.Errorf("update wutong cluster failure %s", err.Error())
		}
		*cluster = old
	}
	if err := o.createWutongVolumes(cluster); err != nil {
		return fmt.Errorf("create wutong volume failure %s", err.Error())
	}
	if err := o.createWutongPackage(); err != nil {
		return fmt.Errorf("create wutong package failure %s", err.Error())
	}
	if err := o.createComponents(cluster); err != nil {
		return err
	}
	return nil
}

func (o *Operator) createComponents(cluster *v1alpha1.WutongCluster) error {
	claims := o.genComponentClaims(cluster)
	for _, claim := range claims {
		// update image repository for priority components
		claim.imageRepository = cluster.Spec.WutongImageRepository
		data := parseComponentClaim(claim)
		// init component
		data.Namespace = o.Namespace

		err := retryutil.Retry(time.Second*2, 3, func() (bool, error) {
			if err := o.createResourceIfNotExists(data); err != nil {
				return false, err
			}
			return true, nil
		})
		if err != nil {
			return fmt.Errorf("create wutong component %s failure %s", data.GetName(), err.Error())
		}
	}
	return nil
}

func (o *Operator) genComponentClaims(cluster *v1alpha1.WutongCluster) map[string]*componentClaim {
	var defReplicas = commonutil.Int32(1)
	if cluster.Spec.EnableHA {
		defReplicas = commonutil.Int32(2)
	}

	var isInit bool
	imageRepository := constants.DefImageRepository
	if cluster.Spec.ImageHub == nil || cluster.Spec.ImageHub.Domain == constants.DefImageRepository {
		isInit = true
	} else {
		imageRepository = path.Join(cluster.Spec.ImageHub.Domain, cluster.Spec.ImageHub.Namespace)
	}

	newClaim := func(name string) *componentClaim {
		defClaim := componentClaim{name: name, imageRepository: imageRepository, version: o.WutongVersion, replicas: defReplicas}
		defClaim.imageName = name
		return &defClaim
	}
	name2Claim := map[string]*componentClaim{
		"wt-api":            newClaim("wt-api"),
		"wt-chaos":          newClaim("wt-chaos"),
		"wt-eventlog":       newClaim("wt-eventlog"),
		"wt-monitor":        newClaim("wt-monitor"),
		"wt-mq":             newClaim("wt-mq"),
		"wt-worker":         newClaim("wt-worker"),
		"wt-webcli":         newClaim("wt-webcli"),
		"wt-resource-proxy": newClaim("wt-resource-proxy"),
	}
	if !o.OnlyInstallRegion {
		name2Claim["wt-app-ui"] = newClaim("wt-app-ui")
	}
	name2Claim["metrics-server"] = newClaim("metrics-server")
	name2Claim["metrics-server"].version = "v0.3.6"

	if cluster.Spec.RegionDatabase == nil || (cluster.Spec.UIDatabase == nil && !o.OnlyInstallRegion) {
		claim := newClaim("wt-db")
		claim.version = "8.0.19"
		claim.replicas = commonutil.Int32(1)
		name2Claim["wt-db"] = claim
	}

	if cluster.Spec.ImageHub == nil || cluster.Spec.ImageHub.Domain == constants.DefImageRepository {
		claim := newClaim("wt-hub")
		claim.imageName = "registry"
		claim.version = "2.6.2"
		claim.isInit = isInit
		name2Claim["wt-hub"] = claim
	}

	name2Claim["wt-gateway"] = newClaim("wt-gateway")
	name2Claim["wt-gateway"].isInit = isInit
	name2Claim["wt-node"] = newClaim("wt-node")
	name2Claim["wt-node"].isInit = isInit

	if cluster.Spec.EtcdConfig == nil || len(cluster.Spec.EtcdConfig.Endpoints) == 0 {
		claim := newClaim("wt-etcd")
		claim.imageName = "etcd"
		claim.version = "v3.3.18"
		claim.isInit = isInit
		if cluster.Spec.EnableHA {
			claim.replicas = commonutil.Int32(3)
		}
		name2Claim["wt-etcd"] = claim
	}

	// kubernetes dashboard
	k8sdashboard := newClaim("kubernetes-dashboard")
	k8sdashboard.version = "v2.0.1-3"
	name2Claim["kubernetes-dashboard"] = k8sdashboard
	dashboardscraper := newClaim("dashboard-metrics-scraper")
	dashboardscraper.imageName = "metrics-scraper"
	dashboardscraper.version = "v1.0.4"
	name2Claim["dashboard-metrics-scraper"] = dashboardscraper

	if rwx := cluster.Spec.WutongVolumeSpecRWX; rwx != nil && rwx.CSIPlugin != nil {
		if rwx.CSIPlugin.NFS != nil {
			name2Claim["nfs-provisioner"] = newClaim("nfs-provisioner")
			name2Claim["nfs-provisioner"].replicas = commonutil.Int32(1)
			name2Claim["nfs-provisioner"].isInit = isInit
		}
		if rwx.CSIPlugin.AliyunNas != nil {
			name2Claim[constants.AliyunCSINasPlugin] = newClaim(constants.AliyunCSINasPlugin)
			name2Claim[constants.AliyunCSINasPlugin].isInit = isInit
			name2Claim[constants.AliyunCSINasProvisioner] = newClaim(constants.AliyunCSINasProvisioner)
			name2Claim[constants.AliyunCSINasProvisioner].isInit = isInit
			name2Claim[constants.AliyunCSINasProvisioner].replicas = commonutil.Int32(1)
		}
	}
	if rwo := cluster.Spec.WutongVolumeSpecRWO; rwo != nil && rwo.CSIPlugin != nil {
		if rwo.CSIPlugin.AliyunCloudDisk != nil {
			name2Claim[constants.AliyunCSIDiskPlugin] = newClaim(constants.AliyunCSIDiskPlugin)
			name2Claim[constants.AliyunCSIDiskPlugin].isInit = isInit
			name2Claim[constants.AliyunCSIDiskProvisioner] = newClaim(constants.AliyunCSIDiskProvisioner)
			name2Claim[constants.AliyunCSIDiskProvisioner].isInit = isInit
			name2Claim[constants.AliyunCSIDiskProvisioner].replicas = commonutil.Int32(1)
		}
	}

	return name2Claim
}

func (o *Operator) createWutongPackage() error {
	pkg := &v1alpha1.WutongPackage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.Wutongpackage,
			Namespace: o.Namespace,
		},
		Spec: v1alpha1.WutongPackageSpec{
			PkgPath:      o.ArchiveFilePath,
			ImageHubUser: o.ImageHubUser,
			ImageHubPass: o.ImageHubPass,
		},
	}
	return o.createResourceIfNotExists(pkg)
}

func (o *Operator) createWutongVolumes(cluster *v1alpha1.WutongCluster) error {
	if cluster.Spec.WutongVolumeSpecRWX != nil {
		rwx := setWutongVolume("wutongvolumerwx", o.Namespace, wtutil.LabelsForAccessModeRWX(), cluster.Spec.WutongVolumeSpecRWX)
		rwx.Spec.ImageRepository = o.WutongImageRepository
		if err := o.createResourceIfNotExists(rwx); err != nil {
			return err
		}
	}
	if cluster.Spec.WutongVolumeSpecRWO != nil {
		rwo := setWutongVolume("wutongvolumerwo", o.Namespace, wtutil.LabelsForAccessModeRWO(), cluster.Spec.WutongVolumeSpecRWO)
		rwo.Spec.ImageRepository = o.WutongImageRepository
		if err := o.createResourceIfNotExists(rwo); err != nil {
			return err
		}
	}
	return nil
}

func (o *Operator) createResourceIfNotExists(resource client.Object) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := o.RuntimeClient.Create(ctx, resource)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		return fmt.Errorf("create resource %s/%s failure %s", resource.GetObjectKind(), resource.GetName(), err.Error())
	}
	return nil
}

func setWutongVolume(name, namespace string, labels map[string]string, spec *v1alpha1.WutongVolumeSpec) *v1alpha1.WutongVolume {
	volume := &v1alpha1.WutongVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    wtutil.LabelsForWutong(labels),
		},
		Spec: *spec,
	}
	return volume
}
