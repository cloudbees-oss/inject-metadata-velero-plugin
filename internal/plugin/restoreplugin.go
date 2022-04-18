/*
Copyright 2018, 2019 the Velero contributors and CloudBees.

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

package plugin

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// RestorePlugin is a restore item action plugin for Velero
type RestorePlugin struct {
	log logrus.FieldLogger
}

// NewRestorePlugin instantiates a RestorePlugin.
func NewRestorePlugin(log logrus.FieldLogger) *RestorePlugin {
	return &RestorePlugin{log: log}
}

// AppliesTo returns information about which resources this action should be invoked for.
// The IncludedResources and ExcludedResources slices can include both resources
// and resources with group names. These work: "ingresses", "ingresses.extensions".
// A RestoreItemAction's Execute function will only be invoked on items that match the returned
// selector. A zero-valued ResourceSelector matches all resources.
func (p *RestorePlugin) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		IncludedResources: []string{"statefulsets"},
	}, nil
}

const envVarName = "RESTORED_FROM_BACKUP"

// Execute allows the RestorePlugin to perform arbitrary logic with the item being restored,
// in this case, setting a custom annotation on the item being restored.
func (p *RestorePlugin) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
	var sts apps.StatefulSet
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(input.Item.UnstructuredContent(), &sts); err != nil {
		return nil, errors.WithStack(err)
	}
	log := p.log.WithField("sts", sts.Name)

	for i := range sts.Spec.Template.Spec.Containers {
		var env []core.EnvVar
		for _, v := range sts.Spec.Template.Spec.Containers[i].Env {
			if v.Name != envVarName {
				env = append(env, v)
			}
		}
		env = append(env, core.EnvVar{Name: envVarName, Value: input.Restore.Name})
		sts.Spec.Template.Spec.Containers[i].Env = env
		log.Infof("Added %s to environment of container %s", envVarName, sts.Spec.Template.Spec.Containers[i].Name)
	}

	inputMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&sts)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return velero.NewRestoreItemActionExecuteOutput(&unstructured.Unstructured{Object: inputMap}), nil
}
