package operator

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	routeapi "github.com/openshift/api/route/v1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"

	regopapi "github.com/openshift/cluster-image-registry-operator/pkg/apis/imageregistry/v1alpha1"
	"github.com/openshift/cluster-image-registry-operator/pkg/generate"
	"github.com/openshift/cluster-image-registry-operator/pkg/parameters"
)

func syncRoutes(o *regopapi.ImageRegistry, p *parameters.Globals, modified *bool) error {
	routeList := &routeapi.RouteList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: routeapi.SchemeGroupVersion.String(),
			Kind:       "Route",
		},
	}

	err := sdk.List(p.Deployment.Namespace, routeList)
	if err != nil {
		return fmt.Errorf("failed to list routes: %s", err)
	}

	routes := generate.GetRouteGenerators(o, p)

	for _, route := range routeList.Items {
		if !metav1.IsControlledBy(&route, o) {
			continue
		}
		if _, found := routes[route.ObjectMeta.Name]; found {
			continue
		}
		err = sdk.Delete(&route)
		if err != nil {
			return err
		}
		*modified = true
	}

	return nil
}
