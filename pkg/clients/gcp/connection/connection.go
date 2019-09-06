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

package connection

import (
	"sort"

	servicenetworking "google.golang.org/api/servicenetworking/v1"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"

	"github.com/crossplaneio/crossplane/gcp/apis/servicenetworking/v1alpha1"
)

// FromParameters converts the supplied ConnectionParameters into an
// Address suitable for use with the Google Compute API.
func FromParameters(p v1alpha1.ConnectionParameters) *servicenetworking.Connection {
	// Kubernetes API conventions dictate that optional, unspecified fields must
	// be nil. GCP API clients omit any field set to its zero value, using
	// NullFields and ForceSendFields to handle edge cases around unsetting
	// previously set values, or forcing zero values to be set. The Address API
	// does not support updates, so we can safely convert any nil pointer to
	// string or int64 to their zero values.
	return &servicenetworking.Connection{
		Network:               p.Network,
		ReservedPeeringRanges: p.ReservedPeeringRanges,
		ForceSendFields:       []string{"ReservedPeeringRanges"},
	}
}

// UpToDate returns true if the observed Connection is up to date with the
// supplied ConnectionParameters.
func UpToDate(p v1alpha1.ConnectionParameters, observed *servicenetworking.Connection) bool {
	if len(p.ReservedPeeringRanges) != len(observed.ReservedPeeringRanges) {
		return false
	}

	sort.Strings(p.ReservedPeeringRanges)
	sort.Strings(observed.ReservedPeeringRanges)

	for i := range p.ReservedPeeringRanges {
		if p.ReservedPeeringRanges[i] != observed.ReservedPeeringRanges[i] {
			return false
		}
	}

	return true
}

// UpdateStatus updates any fields of the supplied ConnectionStatus to
// reflect the state of the supplied Address.
func UpdateStatus(s *v1alpha1.ConnectionStatus, observed *servicenetworking.Connection) {
	// If it exists, it's available
	s.SetConditions(runtimev1alpha1.Available())

	s.Peering = observed.Peering
	s.Service = observed.Service
}
