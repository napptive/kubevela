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
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oam-dev/kubevela/apis/core.oam.dev/common"
	"github.com/oam-dev/kubevela/apis/core.oam.dev/v1beta1"
)

var _ = Describe("Component name randomization", func() {

	var app *v1beta1.Application = &v1beta1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "core.oam.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-app",
		},
		Spec: v1beta1.ApplicationSpec{
			Components: []common.ApplicationComponent{
				{
					Name: "component1",
					Type: "test",
				},
				{
					Name: "component2",
					Type: "test",
				},
			},
		},
	}

	Context("randomization functions", func() {
		It("should produce component names with application prefixes", func() {
			componentName := "component1"
			handler := &AppHandler{
				app: app,
			}
			newName := handler.AddAppNameToComponent(componentName)
			Expect(strings.HasPrefix(newName, app.Name)).Should(BeTrue())
			Expect(strings.HasSuffix(newName, componentName)).Should(BeTrue())
		})
		It("should produce component names with random suffixes", func() {
			componentName := "component1"
			handler := &AppHandler{
				app: app,
			}
			newName := handler.AddRandomSuffixToComponent(componentName)
			Expect(strings.HasPrefix(newName, componentName)).Should(BeTrue())
			Expect(len(newName)).Should(Equal(len(componentName) + (randomSuffixLength * 2) + 1))
		})
	})
	FContext("mapping application components", func() {
		It("should be able to transform a simple application with two components", func() {
			handler := &AppHandler{
				app: app,
			}
			expectedMapping := make(map[string]bool)
			for _, c := range app.Spec.Components {
				expectedMapping[handler.AddAppNameToComponent(c.Name)] = true
			}
			modifiedApp, modified, err := handler.RandomizeComponentNames(app)
			Expect(err).To(Succeed())
			Expect(modified).To(BeTrue())
			_, annotationFound := modifiedApp.GetAnnotations()[AnnotationComponentMapping]
			Expect(annotationFound).Should(BeTrue())
			Expect(len(modifiedApp.Spec.Components)).Should(Equal(len(app.Spec.Components)))
			for _, mc := range modifiedApp.Spec.Components {
				_, expectedComponentName := expectedMapping[mc.Name]
				Expect(expectedComponentName).Should(BeTrue())
			}
		})
	})

})
