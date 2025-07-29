package k8s

import (
	"bytes"
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cli-runtime/pkg/printers"
)

func GenerateServiceTemplate(name string, namespace string, ports []ci.Port) ([]byte, error) {
	if namespace == "" {
		namespace = "default"
	}

	service := &core.Service{
		TypeMeta: meta.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: core.ServiceSpec{
			Selector: map[string]string{"app": name},
			Ports:    make([]core.ServicePort, len(ports)),
			Type:     core.ServiceTypeClusterIP,
		},
	}

	for i, port := range ports {
		service.Spec.Ports[i] = core.ServicePort{
			Port:       int32(port.Port),
			TargetPort: intstr.FromInt(port.Port),
			Name:       fmt.Sprintf("%s-%d", name, port.Port),
		}
	}

	yamlWriter := &bytes.Buffer{}
	yamlPrinter := printers.YAMLPrinter{}
	err := yamlPrinter.PrintObj(service, yamlWriter)
	if err != nil {
		return nil, fmt.Errorf("error printing service to yaml: %s", err)
	}

	return yamlWriter.Bytes(), nil
}
