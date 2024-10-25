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
	v1 "k8s.io/api/core/v1"
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
		IncludedResources: []string{"statefulsets", "deployments"},
	}, nil
}

const envVarName = "RESTORED_FROM_BACKUP"

// Execute allows the RestorePlugin to perform arbitrary logic with the item being restored,
// in this case, setting a custom annotation on the item being restored.
func (p *RestorePlugin) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
	var resource interface{}
	var sts apps.StatefulSet
	var deploy apps.Deployment
	kind := input.Item.GetObjectKind().GroupVersionKind().Kind
	if kind == "StatefulSet" {
		resource = &sts
	} else if kind == "Deployment" {
		resource = &deploy
	} else {
		return nil, errors.Errorf("unsupported kind %s", kind)
	}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(input.Item.UnstructuredContent(), resource); err != nil {
		return nil, errors.WithStack(err)
	}
	var name string
	var podTemplateSpec *v1.PodTemplateSpec
	if kind == "StatefulSet" {
		name = sts.Name
		podTemplateSpec = &sts.Spec.Template
	} else if kind == "Deployment" {
		name = deploy.Name
		podTemplateSpec = &deploy.Spec.Template
	} else {
		return nil, errors.Errorf("unsupported kind %s", kind)
	}
	log := p.log.WithField("kind", kind).WithField("name", name)
	log.Infof("Looking for containers")
	for i := range podTemplateSpec.Spec.Containers {
		var env []core.EnvVar
		for _, v := range podTemplateSpec.Spec.Containers[i].Env {
			if v.Name != envVarName {
				env = append(env, v)
			}
		}
		env = append(env, core.EnvVar{Name: envVarName, Value: input.Restore.Name})
		podTemplateSpec.Spec.Containers[i].Env = env
		log.Infof("Added %s to environment of container %s", envVarName, podTemplateSpec.Spec.Containers[i].Name)
	}

	inputMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return velero.NewRestoreItemActionExecuteOutput(&unstructured.Unstructured{Object: inputMap}), nil
}
