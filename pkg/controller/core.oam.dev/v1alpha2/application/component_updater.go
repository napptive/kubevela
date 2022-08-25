/*
Copyright 2022 The KubeVela Authors.

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

package application

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/oam-dev/kubevela/apis/core.oam.dev/v1beta1"
)

const (
	AnnotationComponentMapping    = "app.oam.dev/component-mapping"
	errAnnotatingComponentMapping = "unable to annotate new component name mapping"
	randomSuffixLength            = 6
)

var (
	// EnableComponentNameRandomization is the main flag that determines if component renaming is set.
	EnableComponentNameRandomization = true
	// ComponentNameRandomizationAddingApplication determines if the method to generate new component
	// names is by prefixing the application name.
	ComponentNameRandomizationAddingApplicationName = true
	// ComponentNameRandomizationAddingSuffix determines if the method to generate new component names
	// is by adding a random suffix similar to the k8s pod naming.
	ComponentNameRandomizationAddingSuffix = false
)

// RandomizeComponentNames checks if the application contains the component mapping annotation, and if not
// set iterates over the components randomizing their names.
func (h *AppHandler) RandomizeComponentNames(app *v1beta1.Application) (*v1beta1.Application, bool, error) {
	_, exists := app.GetAnnotations()[AnnotationComponentMapping]
	if exists {
		return nil, false, nil
	}
	klog.InfoS("randomizing component names", "application", h.app.Name)
	componentMapping := make(map[string]string, 0)
	for componentIndex, v := range app.Spec.Components {
		newName := h.GenerateRandomizedComponentName(v.Name)
		klog.InfoS("renaming component", "previous", v.Name, "new_name", newName)
		componentMapping[v.Name] = newName
		app.Spec.Components[componentIndex].Name = newName
	}
	if err := h.SetComponentMappingAnnotation(app, componentMapping); err != nil {
		return nil, false, err
	}
	return app, true, nil
}

// SetComponentMappingAnnotation takes a map of component name mappings, generates the JSON
// equivalent, and sets the corresponding annotation on the application.
func (h *AppHandler) SetComponentMappingAnnotation(app *v1beta1.Application, mapping map[string]string) error {
	klog.InfoS("application component mapping", "app_name", app.Name, "names", mapping)
	encoded, err := json.Marshal(mapping)
	if err != nil {
		return err
	}
	metav1.SetMetaDataAnnotation(&h.app.ObjectMeta, AnnotationComponentMapping, string(encoded))
	return nil
}

// GenerateRandomizedComponentName creates a new name for a component based on the current configuration options.
func (h *AppHandler) GenerateRandomizedComponentName(componentName string) string {
	if ComponentNameRandomizationAddingApplicationName {
		return h.AddAppNameToComponent(componentName)
	} else if ComponentNameRandomizationAddingSuffix {
		return h.AddRandomSuffixToComponent(componentName)
	}
	// return component name as fallback option in case of misconfiguration
	return componentName
}

// AddAppNameToComponent sets the application name as a prefix of the component name.
func (h *AppHandler) AddAppNameToComponent(componentName string) string {
	return fmt.Sprintf("%s-%s", h.app.Name, componentName)
}

// AddRandomSuffixToComponent adds a random suffix to the component name following a similar approach to Kubernetes pod naming.
func (h *AppHandler) AddRandomSuffixToComponent(componentName string) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, randomSuffixLength)
	rand.Read(b)
	return fmt.Sprintf("%s-%x", componentName, b[:randomSuffixLength])
}
