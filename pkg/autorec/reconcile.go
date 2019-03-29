// Copyright (c) Walmart Inc.
//
// This source code is licensed under the Apache 2.0 license found in the
// LICENSE file in the root directory of this source tree.
package autorec

import (
	"context"
	"encoding/json"
	"github.com/walmartlabs/object-diff/pkg/obj_diff"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"log"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ClientManager interface {
	client.Client
	GetScheme() *runtime.Scheme
	GetConfig() *rest.Config
	GetRESTMapper() meta.RESTMapper
	GetContext() context.Context
}

type checkpointType string

const ClientSide checkpointType = "client_checkpoint"
const ServerSide checkpointType = "server_checkpoint"

func AutoReconcile(
	cm ClientManager,
	owner ApiObject,
	clientExpect Replacer,
	serverActual Replacer) error {
	if err := setOwner(owner, clientExpect, cm.GetScheme()); err != nil {
		return err
	}

	namespacedName := types.NamespacedName{Name: clientExpect.GetName(), Namespace: clientExpect.GetNamespace()}
	err := cm.Get(cm.GetContext(), namespacedName, serverActual.GetApiObject())
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if err != nil /* && errors.IsNotFound(err) */ {
		// Object missing, create
		clientExpect, err = setClientCheckpoint(clientExpect, clientExpect)
		if err != nil {
			return err
		}

		log.Printf("Creating %T %s/%s\n", clientExpect.GetApiObject(), clientExpect.GetNamespace(), clientExpect.GetName())
		err = cm.Create(cm.GetContext(), clientExpect.GetApiObject())
		if err != nil {
			return err
		}

		serverActual = clientExpect
	} else {
		// Object exists, modify
		clientActual, changesMade, err := reconcileObject(clientExpect, serverActual)
		if err != nil {
			return err
		} else if !changesMade {
			// Nothing has changed, no updates necessary.
			return nil
		}

		clientActual, err = setClientCheckpoint(clientActual, clientExpect)
		if err != nil {
			return err
		}

		log.Printf("Updating %T %s/%s\n", clientActual.GetApiObject(), clientActual.GetNamespace(), clientActual.GetName())
		err = cm.Update(cm.GetContext(), clientActual.GetApiObject())
		if err != nil {
			return err
		}

		serverActual = clientActual
	}

	err = setServerCheckpoint(cm, serverActual)
	if err != nil {
		return err
	}

	return nil
}

func setOwner(owner ApiObject, object ApiObject, scheme *runtime.Scheme) error {
	return controllerutil.SetControllerReference(owner, object, scheme)
}

func setClientCheckpoint(checkpointDst Replacer, checkpointSrc Replacer) (Replacer, error) {
	srcAnnotations := checkpointSrc.GetAnnotations()
	if srcAnnotations == nil {
		srcAnnotations = map[string]string{}
	}

	serverCheckpointBytes, serverCheckpointExists := srcAnnotations[string(ServerSide)]

	delete(srcAnnotations, string(ClientSide))
	delete(srcAnnotations, string(ServerSide))
	checkpointSrc.SetAnnotations(srcAnnotations)

	clientCheckpointBytes, err := json.Marshal(checkpointSrc.GetApiObject())
	if err != nil {
		return nil, err
	}

	dstAnnotations := checkpointDst.GetAnnotations()
	if dstAnnotations == nil {
		dstAnnotations = map[string]string{}
	}

	dstAnnotations[string(ClientSide)] = string(clientCheckpointBytes)
	if serverCheckpointExists {
		dstAnnotations[string(ServerSide)] = serverCheckpointBytes
	}
	checkpointDst.SetAnnotations(dstAnnotations)

	return checkpointDst, nil
}

func reconcileObject(clientExpect Replacer, serverActual Replacer) (Replacer, bool, error) {
	serverActualCopy := serverActual.GetApiObject().DeepCopyObject()
	serverActualPart, err := serverActual.Extract()
	if err != nil {
		return nil, false, err
	}

	clientExpectPart, err := clientExpect.Extract()
	if err != nil {
		return nil, false, err
	}

	clientCheckpoint, serverCheckpoint, err := getCheckpoints(serverActual)
	if err != nil {
		return nil, false, err
	}

	clientCheckpointPart, err := clientCheckpoint.Extract()
	if err != nil {
		return nil, false, err
	}

	changesMade := false

	// In rare cases it is possible that a serverCheckpoint is not saved. We
	// skip this part of the process in that case. A missing server checkpoint
	// does have the capability to enact lasting change on the resource as we
	// have no way of knowing if something has changed.
	if serverCheckpoint != nil {
		serverCheckpointPart, err := serverCheckpoint.Extract()
		if err != nil {
			return nil, false, err
		}

		actualServerChanges, err := obj_diff.Diff(serverActualPart, serverCheckpointPart)
		if err != nil {
			return nil, false, err
		}
		if len(actualServerChanges.Changes) > 0 {
			err := actualServerChanges.Patch(serverActualPart)
			if err != nil {
				return nil, false, err
			}

			changesMade = true
		}
	}

	clientExpectChanges, err := obj_diff.Diff(clientCheckpointPart, clientExpectPart)
	if err != nil {
		return nil, false, err
	}
	if len(clientExpectChanges.Changes) > 0 {
		err := clientExpectChanges.Patch(serverActualPart)
		if err != nil {
			return nil, false, err
		}

		changesMade = true
	}

	if changesMade {
		err := serverActual.Implant(serverActualPart)
		if err != nil {
			return nil, false, err
		}
	}

	// Here we use reflect.DeepEqual instead of changesMade as our changes may
	// cancel each other out, or have not resulted in any actual change.
	return serverActual, !reflect.DeepEqual(serverActualCopy, serverActual.GetApiObject()), nil
}

func getCheckpoints(serverActual Replacer) (clientCheckpoint Replacer, serverCheckpoint Replacer, err error) {
	replacerType := reflect.TypeOf(serverActual).Elem()
	objectType := reflect.TypeOf(serverActual.GetApiObject()).Elem()

	clientCheckpointObject := reflect.New(objectType).Interface().(ApiObject)
	clientCheckpoint = reflect.New(replacerType).Interface().(Replacer)
	if err := clientCheckpoint.SetReplacerConfig(serverActual.GetReplacerConfig()); err != nil {
		return nil, nil, err
	}

	serverCheckpointObject := reflect.New(objectType).Interface().(ApiObject)
	serverCheckpoint = reflect.New(replacerType).Interface().(Replacer)
	if err := serverCheckpoint.SetReplacerConfig(serverActual.GetReplacerConfig()); err != nil {
		return nil, nil, err
	}

	annotations := serverActual.GetAnnotations()
	clientCheckpointBytes := annotations[string(ClientSide)]
	err = json.Unmarshal([]byte(clientCheckpointBytes), clientCheckpointObject)
	if err != nil {
		return nil, nil, err
	}
	clientCheckpoint.SetApiObject(clientCheckpointObject)

	serverCheckpointBytes, serverCheckpointExists := annotations[string(ServerSide)]
	if serverCheckpointExists {
		err = json.Unmarshal([]byte(serverCheckpointBytes), serverCheckpointObject)
		if err != nil {
			return nil, nil, err
		}
		serverCheckpoint.SetApiObject(serverCheckpointObject)
	} else {
		serverCheckpoint = nil
	}

	return clientCheckpoint, serverCheckpoint, nil
}

func setServerCheckpoint(cm ClientManager, serverActual Replacer) error {
	// An error in this function can put us in an inconsistent state by failing
	// to set the server-side checkpoint. This should be low risk as we're
	// patching the original object instead of doing a full add or update.
	serverPatch, err := buildServerCheckpoint(serverActual)
	if err != nil {
		log.Printf("Failed marshalling server side patch!")
		return err
	}

	err = patchServerCheckpoint(cm, serverActual, types.JSONPatchType, serverPatch)
	if err != nil {
		log.Printf("Failed applying server side patch!")
		return err
	}

	return nil
}

func buildServerCheckpoint(serverActual Replacer) ([]byte, error) {
	type JsonPatchAdd struct {
		Op    string `json:"op"`
		Path  string `json:"path"`
		Value string `json:"value"`
	}

	annotations := serverActual.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	lastClientDef, lastClientDefExists := annotations[string(ClientSide)]
	lastServerDef, lastServerDefExists := annotations[string(ServerSide)]

	delete(annotations, string(ClientSide))
	delete(annotations, string(ServerSide))

	serverActual.SetAnnotations(annotations)

	serverCheckpointBytes, err := json.Marshal(serverActual.GetApiObject())
	if err != nil {
		return nil, err
	}

	path := "/metadata/annotations/" + string(ServerSide)
	patchObj := []JsonPatchAdd{{Op: "add", Path: path, Value: string(serverCheckpointBytes)}}
	patchBytes, err := json.Marshal(patchObj)
	if err != nil {
		return nil, err
	}

	if lastClientDefExists {
		annotations[string(ClientSide)] = lastClientDef
	}

	if lastServerDefExists {
		annotations[string(ServerSide)] = lastServerDef
	}
	serverActual.SetAnnotations(annotations)

	return patchBytes, nil
}

func patchServerCheckpoint(cm ClientManager, serverActual Replacer, patchType types.PatchType, patch []byte) error {
	scheme := cm.GetScheme()
	config := cm.GetConfig()
	restMapper := cm.GetRESTMapper()

	gvk, err := apiutil.GVKForObject(serverActual.GetApiObject(), scheme)
	if err != nil {
		return err
	}

	restClient, err := apiutil.RESTClientForGVK(gvk, config, serializer.NewCodecFactory(scheme))
	if err != nil {
		return err
	}

	mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	err = restClient.
		Patch(patchType).
		Namespace(serverActual.GetNamespace()).
		Resource(mapping.Resource.Resource).
		Name(serverActual.GetName()).
		Body(patch).
		Do().
		Into(serverActual.GetApiObject())
	if err != nil {
		return err
	}

	return nil
}
