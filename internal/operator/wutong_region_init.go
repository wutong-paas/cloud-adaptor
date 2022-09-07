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
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"github.com/wutong-paas/cloud-adaptor/internal/adaptor/v1alpha1"
	"github.com/wutong-paas/cloud-adaptor/internal/repo"
	"github.com/wutong-paas/cloud-adaptor/version"
	wutongv1alpha1 "github.com/wutong-paas/wutong-operator/api/v1alpha1"
	"github.com/wutong-paas/wutong-operator/util/commonutil"
	"github.com/wutong-paas/wutong-operator/util/constants"
	"github.com/wutong-paas/wutong-operator/util/retryutil"
	"github.com/wutong-paas/wutong-operator/util/suffixdomain"
	"github.com/wutong-paas/wutong-operator/util/wtutil"
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var helmPath = "/usr/local/bin/helm"

func init() {
	if os.Getenv("HELM_PATH") != "" {
		helmPath = os.Getenv("HELM_PATH")
	}
}

// WutongRegionInit wutong region init by operator
type WutongRegionInit struct {
	kubeconfig              v1alpha1.KubeConfig
	namespace               string
	wutongClusterConfigRepo repo.WutongClusterConfigRepository
	wutongCluster           *wutongv1alpha1.WutongCluster
}

// NewWutongRegionInit new
func NewWutongRegionInit(kubeconfig v1alpha1.KubeConfig, wutongClusterConfigRepo repo.WutongClusterConfigRepository, initConfig *v1alpha1.WutongInitConfig) *WutongRegionInit {
	res := &WutongRegionInit{
		kubeconfig:              kubeconfig,
		namespace:               constants.Namespace,
		wutongClusterConfigRepo: wutongClusterConfigRepo,
	}

	if initConfig != nil {
		rcc, err := wutongClusterConfigRepo.Get(initConfig.ClusterID)
		if err != nil && err != gorm.ErrRecordNotFound {
			logrus.Errorf("get wutong cluster config failure %s", err.Error())
		}
		cluster := &wutongv1alpha1.WutongCluster{
			Spec: wutongv1alpha1.WutongClusterSpec{
				InstallVersion:        initConfig.WutongVersion,
				CIVersion:             initConfig.WutongCIVersion,
				EnableHA:              initConfig.EnableHA,
				WutongImageRepository: version.InstallImageRepo,
				SuffixHTTPHost:        initConfig.SuffixHTTPHost,
				NodesForChaos:         initConfig.ChaosNodes,
				NodesForGateway:       initConfig.GatewayNodes,
				GatewayIngressIPs:     initConfig.EIPs,
			},
		}
		if rcc != nil {
			logrus.Info("use custom wutongcluster config")
			if err := yaml.Unmarshal([]byte(rcc.Config), cluster); err != nil {
				logrus.Errorf("Unmarshal wutong config failure %s", err.Error())
			}
		}
		res.wutongCluster = cluster
	}

	return res
}

// InitWutongRegion init wutong region
func (r *WutongRegionInit) InitWutongRegion(initConfig *v1alpha1.WutongInitConfig) error {
	clusterID := initConfig.ClusterID
	kubeconfigFileName := "/tmp/" + clusterID + ".kubeconfig"
	if err := r.kubeconfig.Save(kubeconfigFileName); err != nil {
		return fmt.Errorf("warite kubeconfig file failure %s", err.Error())
	}
	defer func() {
		os.Remove(kubeconfigFileName)
	}()
	// create namespace
	client, runtimeClient, err := r.kubeconfig.GetKubeClient()
	if err != nil {
		return fmt.Errorf("create kube client failure %s", err.Error())
	}
	cn := &v1.Namespace{}
	cn.Name = r.namespace
	if err := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		_, err = client.CoreV1().Namespaces().Create(ctx, cn, metav1.CreateOptions{})
		if err != nil && !k8sErrors.IsAlreadyExists(err) {
			return fmt.Errorf("create namespace failure %s", err.Error())
		}
		return nil
	}(); err != nil {
		return err
	}

	// helm create wutong operator chart
	defaultArgs := []string{
		helmPath, "install", "wutong-operator", "wutong/wutong-operator", "-n", r.namespace,
		"--kubeconfig", kubeconfigFileName,
		"--set", "operator.image.name=" + fmt.Sprintf("%s/wutong-operator", version.InstallImageRepo),
		"--set", "operator.image.tag=" + imageFitArch(version.OperatorVersion, r.wutongCluster.Spec.Arch)}
	logrus.Infof(strings.Join(defaultArgs, " "))
	for {
		var stdout = bytes.NewBuffer(nil)
		cmd := &exec.Cmd{
			Path:   helmPath,
			Args:   defaultArgs,
			Stdout: stdout,
			Stdin:  os.Stdin,
			Stderr: stdout,
		}
		if err := cmd.Run(); err != nil {
			fmt.Println("helm run err:", err.Error())
			errout := stdout.String()
			if !strings.Contains(errout, "cannot re-use a name that is still in use") {
				if strings.Contains(errout, `ClusterRoleBinding "wutong-operator" in namespace`) {
					func() {
						ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
						defer cancel()
						client.RbacV1().ClusterRoleBindings().Delete(ctx, "wutong-operator", metav1.DeleteOptions{})
					}()
					continue
				}
				return fmt.Errorf("install chart failure %s, %s", err.Error(), errout)
			}
			logrus.Warning("wutong operator chart release is exist")
		} else {
			fmt.Println("helm run well.")
		}
		break
	}
	// waiting operator is ready
	ticker := time.NewTicker(time.Second * 5)
	timer := time.NewTimer(time.Minute * 10)
	defer timer.Stop()
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-timer.C:
			return fmt.Errorf("waiting wutong operator ready timeout")
		}
		var rb *rbacv1.ClusterRoleBinding
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			rb, err = client.RbacV1().ClusterRoleBindings().Get(ctx, "wutong-operator", metav1.GetOptions{})
			if err != nil {
				logrus.Errorf("get role binding wutong-operator status failure %s", err.Error())
			}
		}()
		if rb != nil && rb.Name == "wutong-operator" {
			break
		}
	}
	// create custom resource
	if err := r.createWutongCR(client, runtimeClient, initConfig); err != nil {
		return fmt.Errorf("create wutong CR failure %s", err.Error())
	}
	return nil
}

func (r *WutongRegionInit) createWutongCR(kubeClient *kubernetes.Clientset, client client.Client, initConfig *v1alpha1.WutongInitConfig) error {
	// create wutong cluster resource
	//TODO: define etcd config by WutongInitConfig
	cluster := r.wutongCluster
	if cluster == nil {
		return fmt.Errorf("wutong cluster not initialized")
	}

	if len(cluster.Spec.GatewayIngressIPs) == 0 {
		return fmt.Errorf("can not select eip, please specify `gatewayIngressIPs` in the custom cluster init configuration")
	}
	if cluster.Spec.EtcdConfig != nil && len(cluster.Spec.EtcdConfig.Endpoints) == 0 {
		cluster.Spec.EtcdConfig = nil
	}
	cluster.Spec.InstallMode = "WithoutPackage"
	// default build cache mode set is `hostpath`
	if cluster.Spec.CacheMode == "" {
		cluster.Spec.CacheMode = "hostpath"
	}

	cluster.Spec.ConfigCompleted = true
	// image hub must be nil, where not define
	if cluster.Spec.ImageHub != nil && cluster.Spec.ImageHub.Domain == "" {
		cluster.Spec.ImageHub = nil
	}
	if cluster.Spec.InstallVersion == "" {
		cluster.Spec.InstallVersion = initConfig.WutongVersion
	}
	if cluster.Spec.CIVersion == "" {
		cluster.Spec.CIVersion = initConfig.WutongCIVersion
	}
	if cluster.Spec.WutongImageRepository == "" {
		cluster.Spec.WutongImageRepository = version.InstallImageRepo
	}
	if initConfig.ETCDConfig != nil && len(initConfig.ETCDConfig.Endpoints) > 0 {
		cluster.Spec.EtcdConfig = initConfig.ETCDConfig
	}
	if initConfig.RegionDatabase != nil && initConfig.RegionDatabase.Host != "" {
		cluster.Spec.RegionDatabase = &wutongv1alpha1.Database{
			Host:     initConfig.RegionDatabase.Host,
			Port:     initConfig.RegionDatabase.Port,
			Username: initConfig.RegionDatabase.UserName,
			Password: initConfig.RegionDatabase.Password,
		}
	}
	if initConfig.NasServer != "" {
		cluster.Spec.WutongVolumeSpecRWX = &wutongv1alpha1.WutongVolumeSpec{
			CSIPlugin: &wutongv1alpha1.CSIPluginSource{
				AliyunNas: &wutongv1alpha1.AliyunNasCSIPluginSource{
					AccessKeyID:     "",
					AccessKeySecret: "",
				},
			},
			StorageClassParameters: &wutongv1alpha1.StorageClassParameters{
				Parameters: map[string]string{
					"volumeAs":        "subpath",
					"server":          initConfig.NasServer,
					"archiveOnDelete": "true",
				},
			},
		}
	}
	// handle volume spec
	if cluster.Spec.WutongVolumeSpecRWX != nil {
		if cluster.Spec.WutongVolumeSpecRWX.CSIPlugin != nil {
			if cluster.Spec.WutongVolumeSpecRWX.CSIPlugin.AliyunCloudDisk == nil &&
				cluster.Spec.WutongVolumeSpecRWX.CSIPlugin.AliyunNas == nil &&
				cluster.Spec.WutongVolumeSpecRWX.CSIPlugin.NFS == nil {
				cluster.Spec.WutongVolumeSpecRWX.CSIPlugin = nil
			}
		}
	}
	if cluster.Spec.WutongVolumeSpecRWO != nil {
		if cluster.Spec.WutongVolumeSpecRWO.CSIPlugin != nil {
			if cluster.Spec.WutongVolumeSpecRWO.CSIPlugin.AliyunCloudDisk == nil &&
				cluster.Spec.WutongVolumeSpecRWO.CSIPlugin.AliyunNas == nil &&
				cluster.Spec.WutongVolumeSpecRWO.CSIPlugin.NFS == nil {
				cluster.Spec.WutongVolumeSpecRWO.CSIPlugin = nil
			}
		}
		if cluster.Spec.WutongVolumeSpecRWO.CSIPlugin == nil && cluster.Spec.WutongVolumeSpecRWO.StorageClassName == "" {
			cluster.Spec.WutongVolumeSpecRWO = nil
		}
	}
	if cluster.Spec.WutongVolumeSpecRWX == nil ||
		(cluster.Spec.WutongVolumeSpecRWX.CSIPlugin == nil &&
			cluster.Spec.WutongVolumeSpecRWX.StorageClassName == "") {
		cluster.Spec.WutongVolumeSpecRWX = &wutongv1alpha1.WutongVolumeSpec{
			CSIPlugin: &wutongv1alpha1.CSIPluginSource{
				NFS: &wutongv1alpha1.NFSCSIPluginSource{},
			},
		}
	}
	if cluster.Spec.SuffixHTTPHost == "" {
		var ip string
		if len(initConfig.GatewayNodes) > 0 {
			ip = initConfig.GatewayNodes[0].InternalIP
		}
		if len(initConfig.EIPs) > 0 && initConfig.EIPs[0] != "" {
			ip = initConfig.EIPs[0]
		}
		if ip != "" {
			err := retryutil.Retry(1*time.Second, 3, func() (bool, error) {
				domain, err := r.genSuffixHTTPHost(kubeClient, ip)
				if err != nil {
					return false, err
				}
				cluster.Spec.SuffixHTTPHost = domain
				return true, nil
			})
			if err != nil {
				logrus.Warningf("generate suffix http host: %v", err)
				cluster.Spec.SuffixHTTPHost = constants.DefHTTPDomainSuffix
			}
		} else {
			cluster.Spec.SuffixHTTPHost = constants.DefHTTPDomainSuffix
		}
	}
	cluster.Name = "wutongcluster"
	cluster.Namespace = r.namespace
	operator, err := NewOperator(Config{
		WutongVersion:         initConfig.WutongVersion,
		Namespace:             r.namespace,
		ArchiveFilePath:       "/opt/wutong/pkg/tgz/wutong.tgz",
		RuntimeClient:         client,
		Wutongpackage:         "wutongpackage",
		WutongImageRepository: version.InstallImageRepo,
		OnlyInstallRegion:     true,
	})
	if err != nil {
		return fmt.Errorf("create operator instance failure %s", err.Error())
	}
	return operator.Install(cluster)
}

func (r *WutongRegionInit) genSuffixHTTPHost(kubeClient *kubernetes.Clientset, ip string) (domain string, err error) {
	id, auth, err := r.getOrCreateUUIDAndAuth(kubeClient)
	if err != nil {
		return "", err
	}
	domain, err = suffixdomain.GenerateDomain(ip, id, auth)
	if err != nil {
		return "", err
	}
	return domain, nil
}

func (r *WutongRegionInit) getOrCreateUUIDAndAuth(kubeClient *kubernetes.Clientset) (id, auth string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	cm, err := kubeClient.CoreV1().ConfigMaps(r.namespace).Get(ctx, "wt-suffix-host", metav1.GetOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return "", "", err
	}
	if k8sErrors.IsNotFound(err) {
		logrus.Info("not found configmap wt-suffix-host, create it")
		cm = generateSuffixConfigMap("wt-suffix-host", r.namespace)
		if _, err = kubeClient.CoreV1().ConfigMaps(r.namespace).Create(ctx, cm, metav1.CreateOptions{}); err != nil {
			return "", "", err
		}

	}
	return cm.Data["uuid"], cm.Data["auth"], nil
}

func generateSuffixConfigMap(name, namespace string) *v1.ConfigMap {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string]string{
			"uuid": string(uuid.NewUUID()),
			"auth": string(uuid.NewUUID()),
		},
	}
	return cm
}

// GetWutongRegionStatus get wutong region status
func (r *WutongRegionInit) GetWutongRegionStatus(clusterID string) (*v1alpha1.WutongRegionStatus, error) {
	coreClient, wutongClient, err := r.kubeconfig.GetKubeClient()
	if err != nil {
		return nil, err
	}
	status := &v1alpha1.WutongRegionStatus{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	deployment, err := coreClient.AppsV1().Deployments("wt-system").Get(ctx, "wutong-operator", metav1.GetOptions{})
	if err != nil {
		logrus.Warningf("get operator failure %s", err.Error())
	}
	if deployment != nil && deployment.Status.ReadyReplicas >= 1 {
		status.OperatorReady = true
	}
	if deployment != nil {
		status.OperatorInstalled = true
	}
	var cluster wutongv1alpha1.WutongCluster
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel2()
	err = wutongClient.Get(ctx2, types.NamespacedName{Name: "wutongcluster", Namespace: "wt-system"}, &cluster)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, err
		}
		logrus.Warningf("get cluster failure %s", err.Error())
	}
	status.WutongCluster = &cluster
	var pkgStatus wutongv1alpha1.WutongPackage
	ctx3, cancel3 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel3()
	err = wutongClient.Get(ctx3, types.NamespacedName{Name: "wutongpackage", Namespace: "wt-system"}, &pkgStatus)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, err
		}
		logrus.Warningf("get pkgStatus failure %s", err.Error())
	}
	status.WutongPackage = &pkgStatus
	var volume wutongv1alpha1.WutongVolume
	ctx4, cancel4 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel4()
	err = wutongClient.Get(ctx4, types.NamespacedName{Name: "wutongvolumerwx", Namespace: "wt-system"}, &volume)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, err
		}
		logrus.Warningf("get wutong volume failure %s", err.Error())
	}
	status.WutongVolume = &volume
	ctx5, cancel5 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel5()
	config, err := coreClient.CoreV1().ConfigMaps("wt-system").Get(ctx5, "region-config", metav1.GetOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		logrus.Warningf("get region config failure %s", err.Error())
	}
	status.RegionConfig = config
	return status, nil
}

// UninstallRegion uninstall
func (r *WutongRegionInit) UninstallRegion(clusterID string) error {
	deleteOpts := metav1.DeleteOptions{
		GracePeriodSeconds: commonutil.Int64(0),
	}
	coreClient, runtimeClient, err := r.kubeconfig.GetKubeClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	// delte wutong components
	if err := runtimeClient.DeleteAllOf(ctx, &wutongv1alpha1.WutongComponent{}, client.InNamespace(r.namespace)); err != nil {
		return fmt.Errorf("delete component failure: %v", err)
	}
	// delete wutong packages
	if err := runtimeClient.DeleteAllOf(ctx, &wutongv1alpha1.WutongPackage{}, client.InNamespace(r.namespace)); err != nil {
		return fmt.Errorf("delete wutong package failure: %v", err)
	}
	// delete wutongvolume
	if err := runtimeClient.DeleteAllOf(ctx, &wutongv1alpha1.WutongVolume{}, client.InNamespace(r.namespace)); err != nil {
		return fmt.Errorf("delete wutong volume failure: %v", err)
	}

	// delete pv based on pvc
	claims, err := coreClient.CoreV1().PersistentVolumeClaims(r.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list pv: %v", err)
	}
	for _, claim := range claims.Items {
		if claim.Spec.VolumeName == "" {
			// unbound pvc
			continue
		}
		if err := coreClient.CoreV1().PersistentVolumes().Delete(ctx, claim.Spec.VolumeName, metav1.DeleteOptions{}); err != nil {
			if k8sErrors.IsNotFound(err) {
				continue
			}
			return fmt.Errorf("delete persistent volume: %v", err)
		}
	}
	// delete pvc
	if err := coreClient.CoreV1().PersistentVolumeClaims(r.namespace).DeleteCollection(ctx, deleteOpts, metav1.ListOptions{}); err != nil {
		return fmt.Errorf("delete persistent volume claims: %v", err)
	}

	// delete pv
	if err := coreClient.CoreV1().PersistentVolumes().DeleteCollection(ctx, deleteOpts, metav1.ListOptions{
		LabelSelector: "belongTo=wutong-operator",
	}); err != nil {
		return fmt.Errorf("delete persistent volume claims: %v", err)
	}

	// delete storage class and csidriver
	wutongLabelSelector := fields.SelectorFromSet(wtutil.LabelsForWutong(nil)).String()
	if err := coreClient.StorageV1().StorageClasses().DeleteCollection(ctx, deleteOpts, metav1.ListOptions{LabelSelector: wutongLabelSelector}); err != nil {
		return fmt.Errorf("delete storageclass: %v", err)
	}
	if err := coreClient.StorageV1().StorageClasses().Delete(ctx, "wutongslsc", metav1.DeleteOptions{}); err != nil {
		if !k8sErrors.IsNotFound(err) {
			return fmt.Errorf("delete storageclass wutongslsc: %v", err)
		}
	}
	if err := coreClient.StorageV1().StorageClasses().Delete(ctx, "wutongsssc", metav1.DeleteOptions{}); err != nil {
		if !k8sErrors.IsNotFound(err) {
			return fmt.Errorf("delete storageclass wutongsssc: %v", err)
		}
	}
	if err := coreClient.StorageV1beta1().CSIDrivers().DeleteCollection(ctx, deleteOpts, metav1.ListOptions{LabelSelector: wutongLabelSelector}); err != nil {
		if !k8sErrors.IsNotFound(err) {
			return fmt.Errorf("delete csidriver: %v", err)
		}
	}

	// delete wutong-operator ClusterRoleBinding
	if err := coreClient.RbacV1().ClusterRoleBindings().Delete(ctx, "wutong-operator", metav1.DeleteOptions{}); err != nil {
		if !k8sErrors.IsNotFound(err) {
			return fmt.Errorf("delete cluster role bindings: %v", err)
		}
	}

	// delete wutong cluster
	var wtcluster wutongv1alpha1.WutongCluster
	if err := runtimeClient.DeleteAllOf(ctx, &wtcluster, client.InNamespace(r.namespace)); err != nil {
		if !k8sErrors.IsNotFound(err) {
			return fmt.Errorf("delete wutong volume failure: %v", err)
		}
	}

	if err := coreClient.CoreV1().Namespaces().Delete(ctx, r.namespace, metav1.DeleteOptions{}); err != nil {
		if !k8sErrors.IsNotFound(err) {
			return fmt.Errorf("delete namespace %s failure: %v", r.namespace, err)
		}
	}
	ticker := time.NewTicker(time.Second * 5)
	timer := time.NewTimer(time.Minute * 10)
	defer timer.Stop()
	defer ticker.Stop()
	for {
		if _, err := coreClient.CoreV1().Namespaces().Get(ctx, r.namespace, metav1.GetOptions{}); err != nil {
			if k8sErrors.IsNotFound(err) {
				return nil
			}
		}
		select {
		case <-timer.C:
			return fmt.Errorf("waiting namespace deleted timeout")
		case <-ticker.C:
			logrus.Debugf("waiting namespace wt-system deleted")
		}
	}
}
