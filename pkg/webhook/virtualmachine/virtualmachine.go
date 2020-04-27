/*
Copyright 2019 The KubeMacPool Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package virtualmachine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	webhookserver "github.com/qinqon/kube-admission-webhook/pkg/webhook/server"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	kubevirt "kubevirt.io/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/k8snetworkplumbingwg/kubemacpool/pkg/pool-manager"
)

var log = logf.Log.WithName("Webhook mutatevirtualmachines")

type virtualMachineAnnotator struct {
	client      client.Client
	decoder     *admission.Decoder
	poolManager *pool_manager.PoolManager
}

// Add adds server modifiers to the server, like registering the hook to the webhook server.
func Add(s *webhookserver.Server, poolManager *pool_manager.PoolManager) error {
	virtualMachineAnnotator := &virtualMachineAnnotator{poolManager: poolManager}
	s.UpdateOpts(webhookserver.WithHook("/mutate-virtualmachines", &webhook.Admission{Handler: virtualMachineAnnotator}))
	return nil
}

// podAnnotator adds an annotation to every incoming pods.
func (a *virtualMachineAnnotator) Handle(ctx context.Context, req admission.Request) admission.Response {
	virtualMachine := &kubevirt.VirtualMachine{}

	err := a.decoder.Decode(req, virtualMachine)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if virtualMachine.Annotations == nil {
		virtualMachine.Annotations = map[string]string{}
	}
	if virtualMachine.Namespace == "" {
		virtualMachine.Namespace = req.AdmissionRequest.Namespace
	}

	if req.AdmissionRequest.Operation == admissionv1beta1.Create {
		err = a.mutateCreateVirtualMachinesFn(ctx, virtualMachine)
		if err != nil {
			return admission.Errored(http.StatusInternalServerError,
				fmt.Errorf("Failed to create virtual machine allocation error: %v", err))
		}
	} else if req.AdmissionRequest.Operation == admissionv1beta1.Update {
		err = a.mutateUpdateVirtualMachinesFn(ctx, virtualMachine)
		if err != nil {
			return admission.Errored(http.StatusInternalServerError,
				fmt.Errorf("Failed to update virtual machine allocation error: %v", err))
		}
	} else {
		log.V(1).Info("received an unsupported request", "req.AdmissionRequest.Operation ", req.AdmissionRequest.Operation, "virtualMachineName", virtualMachine.Name, "virtualMachineNamespace", virtualMachine.Namespace)
	}

	marshaledVirtualMachine, _ := json.Marshal(virtualMachine)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// admission.PatchResponse generates a Response containing patches.
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledVirtualMachine)
}

// mutateCreateVirtualMachinesFn calls the create allocation function
func (a *virtualMachineAnnotator) mutateCreateVirtualMachinesFn(ctx context.Context, virtualMachine *kubevirt.VirtualMachine) error {
	log.Info("got a create mutate virtual machine event",
		"virtualMachineName", virtualMachine.Name,
		"virtualMachineNamespace", virtualMachine.Namespace)

	existingVirtualMachine := &kubevirt.VirtualMachine{}
	err := a.client.Get(context.TODO(), client.ObjectKey{Namespace: virtualMachine.Namespace, Name: virtualMachine.Name}, existingVirtualMachine)
	if err != nil {
		// If the VM does not exist yet, allocate a new MAC address
		if errors.IsNotFound(err) {
			return a.poolManager.AllocateVirtualMachineMac(virtualMachine)
		}

		// Unexpected error
		return err
	}

	// If the object exist this mean the user run kubectl/oc create
	// This request will failed by the api server so we can just leave it without any allocation
	return nil
}

// mutateUpdateVirtualMachinesFn calls the update allocation function
func (a *virtualMachineAnnotator) mutateUpdateVirtualMachinesFn(ctx context.Context, virtualMachine *kubevirt.VirtualMachine) error {
	log.Info("got a update mutate virtual machine event",
		"virtualMachineName", virtualMachine.Name,
		"virtualMachineNamespace", virtualMachine.Namespace)
	previousVirtualMachine := &kubevirt.VirtualMachine{}
	err := a.client.Get(context.TODO(), client.ObjectKey{Namespace: virtualMachine.Namespace, Name: virtualMachine.Name}, previousVirtualMachine)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if !reflect.DeepEqual(virtualMachine.Spec.Template.Spec.Domain.Devices.Interfaces, previousVirtualMachine.Spec.Template.Spec.Domain.Devices.Interfaces) {
		return a.poolManager.UpdateMacAddressesForVirtualMachine(previousVirtualMachine, virtualMachine)
	}
	return nil
}

// InjectClient injects the client into the podAnnotator
func (a *virtualMachineAnnotator) InjectClient(c client.Client) error {
	a.client = c
	return nil
}

// InjectDecoder injects the decoder.
func (a *virtualMachineAnnotator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}
