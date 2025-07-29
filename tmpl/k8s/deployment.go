package k8s

import (
	"bytes"
	"fmt"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
)

func GenerateDeploymentTemplate(name string, namespace string, image string, pullSecret string) ([]byte, error) {
	if namespace == "" {
		namespace = "default"
	}

	deployment := &apps.Deployment{
		TypeMeta: meta.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: apps.DeploymentSpec{
			Selector: &meta.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Labels: map[string]string{"app": name},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  name,
							Image: image,
						},
					},
				},
			},
		},
	}

	if pullSecret != "" {
		deployment.Spec.Template.Spec.ImagePullSecrets = append(deployment.Spec.Template.Spec.ImagePullSecrets,
			core.LocalObjectReference{Name: pullSecret},
		)
	}

	yamlWriter := &bytes.Buffer{}
	yamlPrinter := printers.YAMLPrinter{}
	err := yamlPrinter.PrintObj(deployment, yamlWriter)
	if err != nil {
		return nil, fmt.Errorf("error printing deployment to yaml: %s", err)
	}

	return yamlWriter.Bytes(), nil
}
