/*
  Copyright 2022 The Fluid Authors.

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package thin

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/thin/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	securityutil "github.com/fluid-cloudnative/fluid/pkg/utils/security"
)

func (t *ThinEngine) totalStorageBytesInternal() (total int64, err error) {
	stsName := t.getFuseDaemonsetName()
	pods, err := t.GetRunningPodsOfDaemonset(stsName, t.namespace)
	if err != nil || len(pods) == 0 {
		return
	}
	fileUtils := operations.NewThinFileUtils(pods[0].Name, common.ThinFuseContainer, t.namespace, t.Log)
	total, err = fileUtils.GetUsedSpace(t.getTargetPath())
	if err != nil {
		return
	}

	return
}

func (t *ThinEngine) totalFileNumsInternal() (fileCount int64, err error) {
	stsName := t.getFuseDaemonsetName()
	pods, err := t.GetRunningPodsOfDaemonset(stsName, t.namespace)
	if err != nil || len(pods) == 0 {
		return
	}
	fileUtils := operations.NewThinFileUtils(pods[0].Name, common.ThinFuseContainer, t.namespace, t.Log)
	fileCount, err = fileUtils.GetFileCount(t.getTargetPath())
	if err != nil {
		return
	}

	return
}

func (t *ThinEngine) usedSpaceInternal() (usedSpace int64, err error) {
	stsName := t.getFuseDaemonsetName()
	pods, err := t.GetRunningPodsOfDaemonset(stsName, t.namespace)
	if err != nil || len(pods) == 0 {
		return
	}
	fileUtils := operations.NewThinFileUtils(pods[0].Name, common.ThinFuseContainer, t.namespace, t.Log)
	usedSpace, err = fileUtils.GetUsedSpace(t.getTargetPath())
	if err != nil {
		return
	}

	return
}

func (t *ThinEngine) genUFSMountOptions(m datav1alpha1.Mount, publicOptions map[string]string, publicEncryptOptions []datav1alpha1.EncryptOption) (map[string]string, error) {

	// initialize mount options
	mOptions := map[string]string{}
	if len(m.Options) > 0 {
		mOptions = m.Options
	}
	for key, value := range publicOptions {
		mOptions[key] = value
	}

	// if encryptOptions have the same key with options
	// it will overwrite the corresponding value
	var err error
	mOptions, err = t.genEncryptOptions(m.EncryptOptions, mOptions, m.Name)
	if err != nil {
		return mOptions, err
	}

	//gen public encryptOptions
	mOptions, err = t.genEncryptOptions(publicEncryptOptions, mOptions, m.Name)
	if err != nil {
		return mOptions, err
	}

	return mOptions, nil
}

// thin encrypt mount options
func (t *ThinEngine) genEncryptOptions(EncryptOptions []datav1alpha1.EncryptOption, mOptions map[string]string, name string) (map[string]string, error) {
	for _, item := range EncryptOptions {

		securityutil.UpdateSensitiveKey(item.Name)
		sRef := item.ValueFrom.SecretKeyRef
		secret, err := kubeclient.GetSecret(t.Client, sRef.Name, t.namespace)
		if err != nil {
			t.Log.Error(err, "get secret by mount encrypt options failed", "name", item.Name)
			return mOptions, err
		}

		t.Log.Info("get value from secret", "mount name", name, "secret key", sRef.Key)

		v := secret.Data[sRef.Key]
		mOptions[item.Name] = string(v)
	}

	return mOptions, nil
}