package configmap

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"github.com/testernetes/bdk/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"
)

type Printer struct {
	lock sync.Mutex
	client.Client
}

func (p *Printer) StartFeature(feature *model.Feature) {
	p.UpdateConfigMap(feature)
}

func (p *Printer) FinishFeature(feature *model.Feature) {
	p.UpdateConfigMap(feature)
}

func (p *Printer) StartScenario(feature *model.Feature, scenario *model.Scenario) {
	p.UpdateConfigMap(feature)
}

func (p *Printer) FinishScenario(feature *model.Feature, scenario *model.Scenario) {
	p.UpdateConfigMap(feature)
}

func (p *Printer) UpdateConfigMap(feature *model.Feature) {
	p.lock.Lock()
	defer p.lock.Unlock()

	out, err := yaml.Marshal(feature)
	if err != nil {
		fmt.Println(err)
		return
	}

	name := viper.GetString("format-configmap-name")
	namespace := viper.GetString("format-configmap-namespace")

	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	_, err = controllerutil.CreateOrUpdate(context.Background(), p.Client, cm, func() error {
		if cm.Data == nil {
			cm.Data = map[string]string{}
		}
		cm.Data[strip(feature.Path)] = string(out)
		return nil
	})
	if err != nil {
		fmt.Println(err)
		return
	}
}

func strip(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') ||
			b == '.' ||
			b == '_' {
			result.WriteByte(b)
		} else {
			result.WriteByte('-')
		}
	}
	return result.String()
}

func (p *Printer) Print(feature *model.Feature) {}
