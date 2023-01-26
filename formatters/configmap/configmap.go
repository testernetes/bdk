package configmap

import (
	"context"
	"fmt"

	"github.com/testernetes/bdk/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"
)

type Printer struct{}

func (p Printer) Print(feature *model.Feature) {
	c, err := client.New(config.GetConfigOrDie(), client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	out, err := yaml.Marshal(feature)
	if err != nil {
		fmt.Println(err)
		return
	}
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "results", Namespace: "default"}}
	controllerutil.CreateOrUpdate(context.Background(), c, cm, func() error {
		if cm.Data == nil {
			cm.Data = map[string]string{}
		}
		cm.Data["results"] = string(out)
		return nil
	})
}
