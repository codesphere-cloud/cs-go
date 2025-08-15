// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"bytes"
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
	networking "k8s.io/api/networking/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cli-runtime/pkg/printers"
)

// GenerateIngressTemplate creates a single Ingress resource from a CiYml struct.
// It generates a separate rule for each public service, routing traffic based on host and path.
func GenerateIngressTemplate(ciYml *ci.CiYml, namespace string, host string, ingressClass string) ([]byte, error) {
	if namespace == "" {
		namespace = "default"
	}

	pathType := networking.PathTypePrefix

	var ingressPaths []networking.HTTPIngressPath
	for serviceName, service := range ciYml.Run {
		if service.IsPublic {
			if service.Network.Path != "" {
				ingressPaths = append(ingressPaths, networking.HTTPIngressPath{
					Path:     service.Network.Path,
					PathType: &pathType,
					Backend: networking.IngressBackend{
						Service: &networking.IngressServiceBackend{
							Name: serviceName,
							Port: networking.ServiceBackendPort{
								Number: intstr.FromInt(3000).IntVal,
							},
						},
					},
				})
			}
			for _, path := range service.Network.Paths {
				ingressPaths = append(ingressPaths, networking.HTTPIngressPath{
					Path:     path.Path,
					PathType: &pathType,
					Backend: networking.IngressBackend{
						Service: &networking.IngressServiceBackend{
							Name: serviceName,
							Port: networking.ServiceBackendPort{
								Number: intstr.FromInt(path.Port).IntVal,
							},
						},
					},
				})
			}
		}
	}

	if len(ingressPaths) == 0 {
		return nil, fmt.Errorf("no public paths found in the provided ci file")
	}

	var ingressRules []networking.IngressRule
	ingressRules = append(ingressRules, networking.IngressRule{
		Host: host,
		IngressRuleValue: networking.IngressRuleValue{
			HTTP: &networking.HTTPIngressRuleValue{
				Paths: ingressPaths,
			},
		},
	})

	ingress := &networking.Ingress{
		TypeMeta: meta.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      fmt.Sprintf("%s-ingress", namespace),
			Namespace: namespace,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target": "/",
			},
		},
		Spec: networking.IngressSpec{
			IngressClassName: &ingressClass,
			Rules:            ingressRules,
		},
	}

	yamlWriter := &bytes.Buffer{}
	yamlPrinter := printers.YAMLPrinter{}
	err := yamlPrinter.PrintObj(ingress, yamlWriter)
	if err != nil {
		return nil, fmt.Errorf("error printing ingress to yaml: %s", err)
	}

	return yamlWriter.Bytes(), nil
}
