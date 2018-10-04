package app

import (
	"context"

	"github.com/mumoshu/falco-operator/pkg/apis/mumoshu/v1alpha1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"github.com/gin-gonic/gin/json"
	"fmt"
)

func NewHandler(opts OperateOpts) sdk.Handler {
	return &Handler{OperateOpts: opts}
}

type Handler struct {
	OperateOpts

	nsToRules map[string]map[string]v1alpha1.FalcoRuleSpec
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	fmt.Printf("handling: %+v\n", event)
	switch o := event.Object.(type) {
	case *v1alpha1.FalcoRule:
		logrus.Infof("handling falcorule: %v", o)

		h.Store(o, event.Deleted)

		logrus.Infof("updating configmap...", o)

		cm := h.newConfigmap()

		err := sdk.Update(cm)
		if err != nil {
			if errors.IsNotFound(err) {
				logrus.Infof("creating configmap...", o)

				err := sdk.Create(cm)
				if err != nil {
					if errors.IsAlreadyExists(err) {
						logrus.Info("expected error. retrying: %v", err)
					} else {
						logrus.Errorf("failed to create busybox pod : %v", err)
						return err
					}
				}
			} else {
				return fmt.Errorf("unexpected err: %v", err)
			}
		}
	default:
		fmt.Printf("unexpected event: %+v\n", event)
	}
	return nil
}

func (c *Handler) Store(cr *v1alpha1.FalcoRule, deleted bool) {
	ns := cr.Namespace

	if c.nsToRules == nil {
		c.nsToRules = map[string]map[string]v1alpha1.FalcoRuleSpec{}
	}

	_, ok := c.nsToRules[ns]
	if !ok {
		c.nsToRules[ns] = map[string]v1alpha1.FalcoRuleSpec{}
	}

	if deleted {
		delete(c.nsToRules[ns], cr.Name)
	} else {
		c.nsToRules[ns][cr.Name] = cr.Spec
	}
}

func (c *Handler) newConfigmap() *corev1.ConfigMap {
	files := map[string]string{}
	labels := map[string]string{}

	for ns, nameToRule := range c.nsToRules {
		rules := []v1alpha1.FalcoRuleSpec{}
		for _, rule := range nameToRule {
			rules = append(rules, rule)
		}
		bytes, err := json.Marshal(rules)
		if err != nil {
			panic(err)
		}
		files[ns] = string(bytes)
	}

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.ConfigMapName,
			Namespace: c.ConfigMapNamespace,
			OwnerReferences: []metav1.OwnerReference{},
			Labels: labels,
		},
		Data: files,
	}
}

// newbusyBoxPod demonstrates how to create a busybox pod
func newbusyBoxPod(cr *v1alpha1.FalcoRule) *corev1.Pod {
	labels := map[string]string{
		"app": "busy-box",
	}
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "busy-box",
			Namespace: cr.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cr, schema.GroupVersionKind{
					Group:   v1alpha1.SchemeGroupVersion.Group,
					Version: v1alpha1.SchemeGroupVersion.Version,
					Kind:    "FalcoRule",
				}),
			},
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "docker.io/busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
