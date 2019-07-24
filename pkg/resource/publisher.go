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

package resource

import (
	"context"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplaneio/crossplane/pkg/util"
)

// A PublisherChain chains multiple ManagedPublishers.
type PublisherChain []ManagedConnectionPublisher

// Calls each ManagedConnectionPublisher serially. It returns the first error it
// encounters, if any.
func (pc PublisherChain) ApplyConnection(ctx context.Context, mg Managed, c ConnectionDetails) error {
	for _, p := range pc {
		if err := p.ApplyConnection(ctx, mg, c); err != nil {
			return err
		}
	}
	return nil
}

// Calls each ManagedConnectionPublisher serially. It returns the first error it
// encounters, if any.
func (pc PublisherChain) DeleteConnection(ctx context.Context, mg Managed, c ConnectionDetails) error {
	for _, p := range pc {
		if err := p.DeleteConnection(ctx, mg, c); err != nil {
			return err
		}
	}
	return nil
}

// An APISecretPublisher publishes ConnectionDetails by submitting a Secret to a
// Kubernetes API server.
type APISecretPublisher struct {
	client client.Client
	typer  runtime.ObjectTyper
}

// NewAPISecretPublisher returns a new APISecretPublisher.
func NewAPISecretPublisher(c client.Client, ot runtime.ObjectTyper) *APISecretPublisher {
	return &APISecretPublisher{client: c, typer: ot}
}

// ApplyConnection the supplied ConnectionDetails to a Secret in the same namespace as
// the supplied Managed resource. Applying is a no-op if the secret already
// exists with the supplied ConnectionDetails.
func (a *APISecretPublisher) ApplyConnection(ctx context.Context, mg Managed, c ConnectionDetails) error {
	s := ConnectionSecretFor(mg, MustGetKind(mg, a.typer))

	err := util.CreateOrUpdate(ctx, a.client, s, func() error {
		// Inside this anonymous function s could either be unchanged (if it
		// does not exist in the API server) or updated to reflect its current
		// state according to the API server.
		if c := metav1.GetControllerOf(s); c == nil || c.UID != mg.GetUID() {
			return errors.New(errSecretConflict)
		}

		// NOTE(negz): We want to support additive publishing, i.e. support
		// setting one subset of secret values then later setting another subset
		// without effecting the original subset.
		if s.Data == nil {
			s.Data = make(map[string][]byte, len(c))
		}

		for k, v := range c {
			s.Data[k] = v
		}

		return nil
	})

	return errors.Wrap(err, errCreateOrUpdateSecret)
}

// Delete the connection Secret belonging to Managed Resource.
func (a *APISecretPublisher) DeleteConnection(ctx context.Context, mg Managed, c ConnectionDetails) error {
	s := ConnectionSecretFor(mg, MustGetKind(mg, a.typer))
	return errors.Wrap(IgnoreNotFound(a.client.Delete(ctx, s)), errCreateOrUpdateSecret)
}
