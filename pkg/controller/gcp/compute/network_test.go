/*
Copyright 2019 The Crossplane Authors.

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

package compute

import (
	"testing"

	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane/gcp/apis/compute/v1alpha1"
	googlecompute "google.golang.org/api/compute/v1"
)

type mockNetworksService struct {
	MockGetFn    func(project string, network string) *googlecompute.NetworksGetCall
	MockInsertFn func(project string, network *googlecompute.Network) *googlecompute.NetworksInsertCall
	MockPatchFn  func(project string, network string, network2 *googlecompute.Network) *googlecompute.NetworksPatchCall
	MockDeleteFn func(project string, network string) *googlecompute.NetworksDeleteCall
}

func (n *mockNetworksService) Get(project string, network string) *googlecompute.NetworksGetCall {
	return n.MockGetFn(project, network)
}
func (n *mockNetworksService) Insert(project string, network *googlecompute.Network) *googlecompute.NetworksInsertCall {
	return n.MockInsertFn(project, network)
}
func (n *mockNetworksService) Patch(project string, network string, network2 *googlecompute.Network) *googlecompute.NetworksPatchCall {
	return n.MockPatchFn(project, network, network2)
}
func (n *mockNetworksService) Delete(project string, network string) *googlecompute.NetworksDeleteCall {
	return n.MockDeleteFn(project, network)
}

//func TestConnector_Connect(t *testing.T) {
//	secretName := "test-secret"
//	keyName := "test-creds-key"
//	testSecret := corev1.Secret{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      secretName,
//			Namespace: testNamespace,
//		},
//	}
//	testProvider := gcpv1alpha1.Provider{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: "test",
//			Namespace: testNamespace,
//		},
//		Spec: gcpv1alpha1.ProviderSpec{
//			ProjectID: "test-project",
//			Secret: corev1.SecretKeySelector{
//				LocalObjectReference: corev1.LocalObjectReference{
//					Name: secretName,
//				},
//				Key: keyName,
//			},
//		},
//	}
//	baseCr := &v1alpha1.Network{
//		Spec: v1alpha1.NetworkSpec{
//			ResourceSpec: cpv1alpha1.ResourceSpec{
//				ProviderReference: &corev1.ObjectReference{
//
//				},
//			},
//			GCPNetworkSpec: v1alpha1.GCPNetworkSpec{
//				Name: "test-cr",
//			},
//		},
//	}
//	testClient := test.NewMockClient()
//
//}

func TestExternal(t *testing.T) {
	type want struct {
		resource.ExternalUpdate
		resource.ExternalObservation
		resource.ExternalCreation
		error
	}
	type args struct {
		cr *v1alpha1.Network
		s  networksService
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulCreate": {
			args: args{
				cr: &v1alpha1.Network{
					Spec: v1alpha1.NetworkSpec{
						GCPNetworkSpec: v1alpha1.GCPNetworkSpec{
							Name: "test-network",
						},
					},
				},
				s: &mockNetworksService{
					MockInsertFn: func(project string, network *googlecompute.Network) *googlecompute.NetworksInsertCall {
						return nil
					},
				},
			},
			want: want{
				ExternalCreation: resource.ExternalCreation{
					ConnectionDetails: nil,
				},
			},
		},
	}
}
