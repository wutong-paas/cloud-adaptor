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
	"io/ioutil"
	"os/user"
	"path"
	"reflect"
	"testing"

	"github.com/wutong-paas/cloud-adaptor/internal/adaptor/v1alpha1"
	wutongv1alpha1 "github.com/wutong-paas/wutong-operator/api/v1alpha1"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestInstall(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	configBytes, err := ioutil.ReadFile(path.Join(u.HomeDir, "/.kube/config"))
	if err != nil {
		t.Fatal(err)
	}
	rri := WutongRegionInit{
		kubeconfig: v1alpha1.KubeConfig{Config: string(configBytes)},
	}
	if err := rri.InitWutongRegion(&v1alpha1.WutongInitConfig{
		EnableHA:        false,
		ClusterID:       "texxxxy",
		WutongVersion:   "v5.3.0-cloud",
		WutongCIVersion: "v5.3.0",
		SuffixHTTPHost:  "",
		GatewayNodes: []*wutongv1alpha1.K8sNode{
			{Name: "192.168.56.104", InternalIP: "192.168.56.104"},
		},
		ChaosNodes: []*wutongv1alpha1.K8sNode{
			{Name: "192.168.56.104", InternalIP: "192.168.56.104"},
		},
	}); err != nil {
		t.Fatal(err)
	}
}
func TestGetWutongRegionStatus(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	configBytes, err := ioutil.ReadFile(path.Join(u.HomeDir, "/.kube/config"))
	if err != nil {
		t.Fatal(err)
	}
	rri := WutongRegionInit{
		kubeconfig: v1alpha1.KubeConfig{Config: string(configBytes)},
	}
	status, err := rri.GetWutongRegionStatus("")
	if err != nil {
		t.Fatal(err)
	}
	configMap := status.RegionConfig
	if configMap != nil {
		regionConfig := map[string]string{
			"client.pem":          string(configMap.BinaryData["client.pem"]),
			"client.key.pem":      string(configMap.BinaryData["client.key.pem"]),
			"ca.pem":              string(configMap.BinaryData["ca.pem"]),
			"apiAddress":          configMap.Data["apiAddress"],
			"websocketAddress":    configMap.Data["websocketAddress"],
			"defaultDomainSuffix": configMap.Data["defaultDomainSuffix"],
			"defaultTCPHost":      configMap.Data["defaultTCPHost"],
		}
		body, err := yaml.Marshal(regionConfig)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(string(body))
	}
	t.Logf("%+v", status)
}

func TestUninstallRegion(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	configBytes, err := ioutil.ReadFile(path.Join(u.HomeDir, "/.kube/config"))
	if err != nil {
		t.Fatal(err)
	}
	rri := WutongRegionInit{
		kubeconfig: v1alpha1.KubeConfig{Config: string(configBytes)},
		namespace:  "wt-system",
	}
	err = rri.UninstallRegion("")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	configBytes, err := ioutil.ReadFile(path.Join(u.HomeDir, "/.kube/config"))
	if err != nil {
		t.Fatal(err)
	}
	config := v1alpha1.KubeConfig{Config: string(configBytes)}
	_, baseclient, _ := config.GetKubeClient()
	var obj client.Object = &corev1.Pod{}
	var oldOjb = reflect.New(reflect.ValueOf(obj).Elem().Type()).Interface().(client.Object)
	if err := baseclient.Get(context.TODO(), types.NamespacedName{Name: "wutong-operator-76b867cd66-5b7k4", Namespace: "wt-system"}, oldOjb); err != nil {
		t.Fatal(err)
	}
	t.Log(oldOjb.GetName())
}
