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
	"context"
	"fmt"
	"strings"

	"github.com/crossplaneio/crossplane/gcp/apis/compute/v1alpha1"

	"github.com/pkg/errors"
	googlecompute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	computev1alpha1 "github.com/crossplaneio/crossplane/gcp/apis/compute/v1alpha1"
	gcpapis "github.com/crossplaneio/crossplane/gcp/apis/v1alpha1"
	gcpclients "github.com/crossplaneio/crossplane/pkg/clients/gcp"
)

const (
	// Error strings.
	errNewClient = "cannot create new Compute Service"
	errNotVPC    = "managed resource is not a Network resource"

	namePrefix = "vpc"
)

// NetworkController is the controller for Network CRD.
type NetworkController struct{}

// SetupWithManager creates a new Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func (c *NetworkController) SetupWithManager(mgr ctrl.Manager) error {
	r := resource.NewManagedReconciler(mgr,
		resource.ManagedKind(computev1alpha1.NetworkGroupVersionKind),
		resource.WithExternalConnecter(&connector{client: mgr.GetClient()}))

	name := strings.ToLower(fmt.Sprintf("%s.%s", computev1alpha1.NetworkKindAPIVersion, computev1alpha1.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Network{}).
		Complete(r)
}

type connector struct {
	client      client.Client
	newClientFn func(ctx context.Context, opts ...option.ClientOption) (*googlecompute.Service, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (resource.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Network)
	if !ok {
		return nil, errors.New(errNotVPC)
	}

	provider := &gcpapis.Provider{}
	n := meta.NamespacedNameOf(cr.Spec.ProviderReference)
	if err := c.client.Get(ctx, n, provider); err != nil {
		return nil, errors.Wrapf(err, "cannot get provider %s", n)
	}

	gcpCreds, err := gcpclients.ProviderCredentials(c.client, provider)
	if err != nil {
		return nil, err
	}
	if c.newClientFn == nil {
		c.newClientFn = googlecompute.NewService
	}
	// NOTE(muvaf): when using option.WithCredentials(), scopes are not set at all, even if passed explicitly.
	s, err := c.newClientFn(ctx, option.WithCredentialsJSON(gcpCreds.JSON), option.WithScopes(googlecompute.ComputeScope))
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &external{networksService: s.Networks, projectID: provider.Spec.ProjectID}, nil
}

type external struct {
	networksService *googlecompute.NetworksService
	projectID       string
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (resource.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Network)
	if !ok {
		return resource.ExternalObservation{}, errors.New(errNotVPC)
	}
	if cr.Spec.SpecForProvider == nil || cr.Spec.SpecForProvider.Name == "" {
		return resource.ExternalObservation{
			ResourceExists: false,
		}, nil
	}
	call := c.networksService.Get(c.projectID, cr.Spec.SpecForProvider.Name)
	observed, err := call.Do()
	if gcpclients.IsErrorNotFound(err) {
		return resource.ExternalObservation{
			ResourceExists: false,
		}, nil
	}
	if err != nil {
		return resource.ExternalObservation{}, err
	}
	cr.Status.StatusAtProvider = computev1alpha1.GenerateGCPNetworkStatus(observed)
	return resource.ExternalObservation{
		ResourceExists: true,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (resource.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Network)
	if !ok {
		return resource.ExternalCreation{}, errors.New(errNotVPC)
	}
	if cr.Spec.SpecForProvider == nil {
		cr.Spec.SpecForProvider = &computev1alpha1.GCPNetworkSpec{}
	}
	if cr.Spec.SpecForProvider.Name == "" {
		cr.Spec.SpecForProvider.Name = fmt.Sprintf("%s-%s", namePrefix, string(cr.ObjectMeta.UID))
	}
	call := c.networksService.Insert(
		c.projectID,
		computev1alpha1.GenerateGCPNetworkSpec(cr.Spec.SpecForProvider))
	if _, err := call.Do(); err != nil {
		return resource.ExternalCreation{}, err
	}
	return resource.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (resource.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Network)
	if !ok {
		return resource.ExternalUpdate{}, errors.New(errNotVPC)
	}
	call := c.networksService.Patch(
		c.projectID,
		cr.Spec.SpecForProvider.Name,
		computev1alpha1.GenerateGCPNetworkSpec(cr.Spec.SpecForProvider))
	if _, err := call.Do(); err != nil {
		return resource.ExternalUpdate{}, err
	}
	return resource.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Network)
	if !ok {
		return errors.New(errNotVPC)
	}
	call := c.networksService.Delete(c.projectID, cr.Spec.SpecForProvider.Name)
	if _, err := call.Do(); !gcpclients.IsErrorNotFound(err) && err != nil {
		return err
	}
	return nil
}
