package k8s

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	k8sScheme        = runtime.NewScheme()
	metadataAccessor = meta.NewAccessor()
)

const (
	AnnotationLastAppliedConfig = "last-applied-configuration"
)

func patchApply(c ResourceInterface, body, rcv interface{}, dryrun ...string) error {
	var desired client.Object
	if obj, ok := body.(client.Object); ok {
		desired = obj
	} else {
		panic("body does not implement the Object interfaces")
	}

	existing, created, err := createOrGetExisting(c, desired, dryrun...)
	if err != nil {
		return err
	}

	if created {
		return nil
	}

	patch, err := threeWayMergePatch(existing, desired)
	if err != nil {
		return err
	}
	bytes, _ := patch.Data(nil)
	return c.Patch(desired.GetName(), patch.Type(), bytes, rcv, metav1.PatchOptions{
		DryRun: dryrun,
	})
}

func threeWayMergePatch(currentObj, modifiedObj client.Object) (client.Patch, error) {
	current, err := json.Marshal(currentObj)
	if err != nil {
		return nil, err
	}
	original, err := getOriginalConfiguration(currentObj)
	if err != nil {
		return nil, err
	}
	modified, err := getModifiedConfiguration(modifiedObj, true)
	if err != nil {
		return nil, err
	}

	var patchType types.PatchType
	var patchData []byte
	var lookupPatchMeta strategicpatch.LookupPatchMeta

	versionedObject, err := k8sScheme.New(currentObj.GetObjectKind().GroupVersionKind())
	switch {
	case runtime.IsNotRegisteredError(err):
		// use JSONMergePatch for custom resources
		// because StrategicMergePatch doesn't support custom resources
		patchType = types.MergePatchType
		preconditions := []mergepatch.PreconditionFunc{
			mergepatch.RequireKeyUnchanged("apiVersion"),
			mergepatch.RequireKeyUnchanged("kind"),
			mergepatch.RequireMetadataKeyUnchanged("name")}
		patchData, err = jsonmergepatch.CreateThreeWayJSONMergePatch(original, modified, current, preconditions...)
		if err != nil {
			return nil, err
		}
	case err != nil:
		return nil, err
	default:
		// use StrategicMergePatch for K8s built-in resources
		patchType = types.StrategicMergePatchType
		lookupPatchMeta, err = strategicpatch.NewPatchMetaFromStruct(versionedObject)
		if err != nil {
			return nil, err
		}
		patchData, err = strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, true)
		if err != nil {
			return nil, err
		}
	}
	return client.RawPatch(patchType, patchData), nil
}

// getOriginalConfiguration gets original configuration of the object
// form the annotation, or nil if no annotation found.
func getOriginalConfiguration(obj runtime.Object) ([]byte, error) {
	annots, err := metadataAccessor.Annotations(obj)
	if err != nil {
		return nil, err
	}
	if annots == nil {
		return nil, nil
	}
	original, ok := annots[AnnotationLastAppliedConfig]
	if !ok {
		return nil, nil
	}
	return []byte(original), nil
}

// getModifiedConfiguration serializes the object into byte stream.
// If `updateAnnotation` is true, it embeds the result as an annotation in the
// modified configuration.
func getModifiedConfiguration(obj runtime.Object, updateAnnotation bool) ([]byte, error) {
	annots, err := metadataAccessor.Annotations(obj)
	if err != nil {
		return nil, err
	}
	if annots == nil {
		annots = make(map[string]string)
	}

	original := annots[AnnotationLastAppliedConfig]
	// remove the annotation to avoid recursion
	delete(annots, AnnotationLastAppliedConfig)
	_ = metadataAccessor.SetAnnotations(obj, annots)
	// do not include an empty map
	if len(annots) == 0 {
		_ = metadataAccessor.SetAnnotations(obj, nil)
	}

	var modified []byte
	modified, err = json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	if updateAnnotation {
		annots[AnnotationLastAppliedConfig] = string(modified)
		err = metadataAccessor.SetAnnotations(obj, annots)
		if err != nil {
			return nil, err
		}
		modified, err = json.Marshal(obj)
		if err != nil {
			return nil, err
		}
	}

	// restore original annotations back to the object
	annots[AnnotationLastAppliedConfig] = original
	_ = metadataAccessor.SetAnnotations(obj, annots)
	return modified, nil
}

// addLastAppliedConfigAnnotation creates annotation recording current configuration as
// original configuration for latter use in computing a three way diff
func addLastAppliedConfigAnnotation(obj runtime.Object) error {
	config, err := getModifiedConfiguration(obj, false)
	if err != nil {
		return err
	}
	annots, _ := metadataAccessor.Annotations(obj)
	if annots == nil {
		annots = make(map[string]string)
	}
	annots[AnnotationLastAppliedConfig] = string(config)
	return metadataAccessor.SetAnnotations(obj, annots)
}

// createOrGetExisting will create the object if it does not exist
// or get and return the existing object
func createOrGetExisting(c ResourceInterface, desired client.Object, dryrun ...string) (client.Object, bool, error) {
	existing := &unstructured.Unstructured{}
	existing.GetObjectKind().SetGroupVersionKind(desired.GetObjectKind().GroupVersionKind())
	var create = func() (client.Object, bool, error) {
		if err := addLastAppliedConfigAnnotation(desired); err != nil {
			return nil, false, err
		}

		err := c.Create(desired, existing, metav1.CreateOptions{
			DryRun: dryrun,
		})
		if err != nil {
			return nil, false, err
		}
		return existing, true, nil
	}

	// allow to create object with only generateName
	if desired.GetName() == "" && desired.GetGenerateName() != "" {
		return create()
	}

	err := c.Get(desired.GetName(), existing, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return create()
	}
	if err != nil {
		return nil, false, err
	}
	return existing, false, nil
}
