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

package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"k8s.io/client-go/util/homedir"
)

//GenerateKey -
func GenerateKey(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	private, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	return private, &private.PublicKey, nil

}

//EncodePrivateKey -
func EncodePrivateKey(private *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Bytes: x509.MarshalPKCS1PrivateKey(private),
		Type:  "RSA PRIVATE KEY",
	})
}

//EncodePublicKey -
func EncodePublicKey(public *rsa.PublicKey) ([]byte, error) {
	publicBytes, err := x509.MarshalPKIXPublicKey(public)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Bytes: publicBytes,
		Type:  "PUBLIC KEY",
	}), nil
}

//EncodeSSHKey -
func EncodeSSHKey(public *rsa.PublicKey) ([]byte, error) {
	publicKey, err := ssh.NewPublicKey(public)
	if err != nil {
		return nil, err
	}
	return ssh.MarshalAuthorizedKey(publicKey), nil
}

//MakeSSHKeyPair -
func MakeSSHKeyPair() (string, string, error) {

	pkey, pubkey, err := GenerateKey(2048)
	if err != nil {
		return "", "", err
	}

	pub, err := EncodeSSHKey(pubkey)
	if err != nil {
		return "", "", err
	}

	return string(EncodePrivateKey(pkey)), string(pub), nil
}

//GetOrMakeSSHRSA get or make ssh rsa
func GetOrMakeSSHRSA() (string, error) {
	home := homedir.HomeDir()
	if _, err := os.Stat(path.Join(home, ".ssh")); err != nil && os.IsNotExist(err) {
		os.MkdirAll(path.Join(home, ".ssh"), 0700)
	}
	idRsaPath := path.Join(home, ".ssh", "id_rsa")
	idRsaPubPath := path.Join(home, ".ssh", "id_rsa.pub")
	stat, err := os.Stat(idRsaPubPath)
	if os.IsNotExist(err) || stat.IsDir() {
		os.Remove(idRsaPath)
		os.Remove(idRsaPubPath)
		private, pub, err := MakeSSHKeyPair()
		if err != nil {
			return "", fmt.Errorf("create ssh rsa failure %s", err.Error())
		}
		logrus.Infof("init ssh rsa file %s %s ", idRsaPath, idRsaPubPath)
		if err := ioutil.WriteFile(idRsaPath, []byte(private), 0600); err != nil {
			return "", fmt.Errorf("write ssh rsa file failure %s", err.Error())
		}
		if err := ioutil.WriteFile(idRsaPubPath, []byte(pub), 0644); err != nil {
			return "", fmt.Errorf("write ssh rsa pub file failure %s", err.Error())
		}
		return pub, nil
	}
	if err != nil {
		return "", err
	}
	pub, err := ioutil.ReadFile(idRsaPubPath)
	if err != nil {
		return "", fmt.Errorf("read rsa pub file failure %s", err.Error())
	}
	return string(pub), nil
}
