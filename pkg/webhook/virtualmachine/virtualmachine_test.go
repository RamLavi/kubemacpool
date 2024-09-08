/*
Copyright 2024 The KubeMacPool Authors.

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
	"net"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	virtv1 "kubevirt.io/api/core/v1"

	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	pool_manager "github.com/k8snetworkplumbingwg/kubemacpool/pkg/pool-manager"
)

var _ = Describe("Handle", func() {
	var testEnv *envtest.Environment

	BeforeEach(func() {
		testEnv = &envtest.Environment{}
		_, err := testEnv.Start()
		Expect(err).NotTo(HaveOccurred())

		Expect(virtv1.AddToScheme(scheme.Scheme)).To(Succeed())
		// +kubebuilder:scaffold:scheme
	})

	AfterEach(func() {
		Expect(testEnv.Stop()).To(Succeed())
	})

	DescribeTable("admits / rejects pod creation requests as expected", func(inputVM *virtv1.VirtualMachine, expectedAdmissionResponse admission.Response) {
		vmWebhookManager, err := createWebhookManager()
		Expect(err).NotTo(HaveOccurred())

		Fail("RAM")
		Expect(
			vmWebhookManager.Handle(context.Background(), vmAdmissionRequest(inputVM)),
		).To(
			Equal(expectedAdmissionResponse),
		)
	},
		Entry("VM without template",
			createVMWithTemplate(nil),
			admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					Allowed: true,
					Result: &metav1.Status{
						Message: "no secondary networks requested",
						Code:    http.StatusOK,
					},
				},
			},
		),
		Entry("VM with template",
			createVMWithTemplate(VMTemplate()),
			admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					Allowed: true,
					Result: &metav1.Status{
						Message: "no secondary networks requested",
						Code:    http.StatusOK,
					},
				},
			},
		),
	)
})

func vmAdmissionRequest(vm *virtv1.VirtualMachine) admission.Request {
	rawPod, err := json.Marshal(vm)
	if err != nil {
		return admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{Object: runtime.RawExtension{}}}
	}
	return admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Object: runtime.RawExtension{
				Raw: rawPod,
			},
		},
	}
}

func createPoolManager(startMacAddr, endMacAddr string, fakeClient client.Client) (*pool_manager.PoolManager, error) {
	startPoolRangeEnv, err := net.ParseMAC(startMacAddr)
	if err != nil {
		return nil, err
	}
	endPoolRangeEnv, err := net.ParseMAC(endMacAddr)
	if err != nil {
		return nil, err
	}
	const (
		testManagerNamespace = "kubemacpool-system"
		waitTimeSeconds      = 10
	)
	poolManager, err := pool_manager.NewPoolManager(fakeClient, fakeClient, startPoolRangeEnv, endPoolRangeEnv, testManagerNamespace, false, waitTimeSeconds)
	if err != nil {
		return nil, err
	}
	err = poolManager.Start()
	return poolManager, err
}

func createWebhookManager() (*virtualMachineAnnotator, error) {
	var initialObjects []client.Object
	ctrlOptions := controllerruntime.Options{
		Scheme: scheme.Scheme,
		NewClient: func(cache cache.Cache, config *rest.Config, options client.Options, uncachedObjects ...client.Object) (client.Client, error) {
			return fake.NewClientBuilder().
				WithScheme(scheme.Scheme).
				WithObjects(initialObjects...).
				Build(), nil
		},
	}

	mgr, err := controllerruntime.NewManager(config.GetConfigOrDie(), ctrlOptions)
	if err != nil {
		return nil, err
	}

	Fail("RAM2")
	decoder, err := admission.NewDecoder(mgr.GetScheme())
	if err != nil {
		return nil, err
	}

	const (
		startPoolRangeEnv = "02:00:00:00:00:00"
		endPoolRangeEnv   = "02:FF:FF:FF:FF:FF"
	)
	poolManager, err := createPoolManager(startPoolRangeEnv, endPoolRangeEnv, mgr.GetClient())
	if err != nil {
		return nil, err
	}

	vmWebhookManager := &virtualMachineAnnotator{
		client:      mgr.GetClient(),
		decoder:     decoder,
		poolManager: poolManager,
	}
	return vmWebhookManager, nil
}

func createVMWithTemplate(template *virtv1.VirtualMachineInstanceTemplateSpec) *virtv1.VirtualMachine {
	const (
		vmName      = "vm1"
		vmNamespace = "default"
	)
	return &virtv1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{Name: vmName, Namespace: vmNamespace},
		Spec: virtv1.VirtualMachineSpec{
			Template: template,
		},
	}
}

func VMTemplate() *virtv1.VirtualMachineInstanceTemplateSpec {
	masqueradeInterface := virtv1.Interface{
		Name: "pod",
		InterfaceBindingMethod: virtv1.InterfaceBindingMethod{
			Masquerade: &virtv1.InterfaceMasquerade{}}}
	podNetwork := virtv1.Network{Name: "pod", NetworkSource: virtv1.NetworkSource{Pod: &virtv1.PodNetwork{}}}

	return &virtv1.VirtualMachineInstanceTemplateSpec{
		Spec: virtv1.VirtualMachineInstanceSpec{
			Domain: virtv1.DomainSpec{
				Devices: virtv1.Devices{
					Interfaces: []virtv1.Interface{masqueradeInterface}}},
			Networks: []virtv1.Network{podNetwork}}}
}
