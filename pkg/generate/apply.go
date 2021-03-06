package generate

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/sdk"

	"k8s.io/apimachinery/pkg/api/errors"
	kmeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/util/retry"

	"github.com/openshift/cluster-image-registry-operator/pkg/parameters"
)

func checksum(o interface{}) (string, error) {
	data, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(data)), nil
}

func ApplyTemplate(tmpl Template, force bool, modified *bool) error {
	expected := tmpl.Expected()

	dgst, err := checksum(expected)
	if err != nil {
		return fmt.Errorf("unable to generate checksum for %s: %s", tmpl.Name(), err)
	}

	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		current := expected

		err := sdk.Get(current)
		if err != nil {
			if !errors.IsNotFound(err) {
				return fmt.Errorf("failed to get object %s: %s", tmpl.Name(), err)
			}
			err = sdk.Create(expected)
			if err == nil {
				*modified = true
				return nil
			}
			return fmt.Errorf("failed to create object %s: %s", tmpl.Name(), err)
		}

		if tmpl.Validator != nil {
			err = tmpl.Validator(current)
			if err != nil {
				return err
			}
		}

		currentMeta, err := kmeta.Accessor(current)
		if err != nil {
			return fmt.Errorf("unable to get meta accessor for current object %s: %s", tmpl.Name(), err)
		}

		curdgst, ok := currentMeta.GetAnnotations()[parameters.ChecksumOperatorAnnotation]
		if !force && ok && dgst == curdgst {
			return nil
		}

		updated, err := tmpl.Apply(current)
		if err != nil {
			return fmt.Errorf("unable to apply template %s: %s", tmpl.Name(), err)
		}

		updatedMeta, err := kmeta.Accessor(updated)
		if err != nil {
			return fmt.Errorf("unable to get meta accessor for updated object %s: %s", tmpl.Name(), err)
		}

		if updatedMeta.GetAnnotations() == nil {
			updatedMeta.SetAnnotations(map[string]string{})
		}
		updatedMeta.GetAnnotations()[parameters.ChecksumOperatorAnnotation] = dgst

		if force {
			updatedMeta.SetGeneration(currentMeta.GetGeneration() + 1)
		}

		err = sdk.Update(updated)
		if err == nil {
			*modified = true
			return nil
		}
		return fmt.Errorf("failed to update object %s: %s", tmpl.Name(), err)
	})
}

func RemoveByTemplate(tmpl Template, modified *bool) error {
	err := sdk.Delete(tmpl.Expected())
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete %s: %s", tmpl.Name(), err)
		}
		return nil
	}
	*modified = true
	return nil
}
