package generate

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/cluster-image-registry-operator/pkg/apis/imageregistry/v1alpha1"
	"github.com/openshift/cluster-image-registry-operator/pkg/parameters"
	"github.com/openshift/cluster-image-registry-operator/pkg/strategy"
)

func ServiceAccount(cr *v1alpha1.ImageRegistry, p *parameters.Globals) (Template, error) {
	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Pod.ServiceAccount,
			Namespace: p.Deployment.Namespace,
		},
	}
	addOwnerRefToObject(sa, asOwner(cr))
	return Template{
		Object:   sa,
		Strategy: strategy.Override{},
	}, nil
}
