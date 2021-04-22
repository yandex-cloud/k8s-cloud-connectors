// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package k8s_fake

import (
	"context"
	"github.com/jinzhu/copier"
	"k8s-connectors/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FakeClient struct {
	scheme *runtime.Scheme
	objects map[types.NamespacedName]client.Object
}

func NewFakeClient() client.Client {
	return &FakeClient{
		scheme: nil,
		objects: map[types.NamespacedName]client.Object{},
	}
}

func NewFakeClientWithScheme(scheme *runtime.Scheme) client.Client {
	return &FakeClient{
		scheme:  scheme,
		objects: map[types.NamespacedName]client.Object{},
	}
}

func (r *FakeClient) Scheme() *runtime.Scheme {
	return r.scheme
}

func (r *FakeClient) RESTMapper() meta.RESTMapper {
	// TODO (covariance) implement me!
	panic("not implemented")
}

// Get retrieves an obj for the given object key from the Kubernetes Cluster.
// obj must be a struct pointer so that obj can be updated with the response
// returned by the Server.
func (r *FakeClient) Get(_ context.Context, key client.ObjectKey, obj client.Object) error {
	if _, ok := r.objects[key]; !ok {
		return errors.NewNotFound(schema.GroupResource{
			Group:    obj.GetObjectKind().GroupVersionKind().Group,
			Resource: obj.GetObjectKind().GroupVersionKind().Kind,
		}, key.String())
	}

	// TODO (covariance) check type mismatch behaviour
	return copier.Copy(obj, r.objects[key])
}

// List retrieves list of objects for a given namespace and list options. On a
// successful call, Items field in the list will be populated with the
// result returned from the server.
func (r *FakeClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	// TODO (covariance) implement me!
	panic("not implemented")
}

// Create saves the object obj in the Kubernetes cluster.
func (r *FakeClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	r.objects[utils.NamespacedName(obj)] = obj
	return nil
}

// Delete deletes the given obj from Kubernetes cluster.
func (r *FakeClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	delete(r.objects, utils.NamespacedName(obj))
	return nil
}

// Update updates the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (r *FakeClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if _, ok := r.objects[utils.NamespacedName(obj)]; !ok {
		return errors.NewNotFound(schema.GroupResource{
			Group:    obj.GetObjectKind().GroupVersionKind().Group,
			Resource: obj.GetObjectKind().GroupVersionKind().Kind,
		}, utils.NamespacedName(obj).String())
	}

	return r.Create(ctx, obj)
}

// Patch patches the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (r *FakeClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	// TODO (covariance) implement me!
	panic("not implemented")
}

// DeleteAllOf deletes all objects of the given type matching the given options.
func (r *FakeClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	// TODO (covariance) implement me!
	panic("not implemented")
}

// StatusClient knows how to create a client which can update status subresource
// for kubernetes objects.
func (r *FakeClient) Status() client.StatusWriter {
	return &FakeStatusWriter{}
}

type FakeStatusWriter struct {}

// Update updates the fields corresponding to the status subresource for the
// given obj. obj must be a struct pointer so that obj can be updated
// with the content returned by the Server.
func (r *FakeStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	// TODO (covariance) implement me!
	panic("not implemented")
}

// Patch patches the given object's subresource. obj must be a struct
// pointer so that obj can be updated with the content returned by the
// Server.
func (r *FakeStatusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	// TODO (covariance) implement me!
	panic("not implemented")
}
