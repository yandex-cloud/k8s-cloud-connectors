// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package k8sfake

import (
	"context"

	"github.com/jinzhu/copier"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s-connectors/pkg/util"
)

type FakeClient struct {
	scheme  *runtime.Scheme
	objects map[types.NamespacedName]client.Object
}

func NewFakeClient() client.Client {
	return &FakeClient{
		scheme:  nil,
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
		return errors.NewNotFound(
			schema.GroupResource{
				Group:    obj.GetObjectKind().GroupVersionKind().Group,
				Resource: obj.GetObjectKind().GroupVersionKind().Kind,
			}, key.String(),
		)
	}

	// TODO (covariance) check type mismatch behavior
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
	dcObj := obj.DeepCopyObject().(client.Object)

	// This is a workaround for Secret storage system
	// Read: https://pkg.go.dev/k8s.io/api/core/v1@v0.20.2#Secret.StringData
	if sec, ok := dcObj.(*v1.Secret); ok {
		if sec.Data == nil {
			sec.Data = make(map[string][]byte)
		}
		for k, v := range sec.StringData {
			sec.Data[k] = []byte(v)
		}
		sec.StringData = nil
	}
	r.objects[util.NamespacedName(dcObj)] = dcObj
	return nil
}

// Delete deletes the given obj from Kubernetes cluster.
func (r *FakeClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	delete(r.objects, util.NamespacedName(obj))
	return nil
}

// Update updates the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (r *FakeClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if _, ok := r.objects[util.NamespacedName(obj)]; !ok {
		return errors.NewNotFound(
			schema.GroupResource{
				Group:    obj.GetObjectKind().GroupVersionKind().Group,
				Resource: obj.GetObjectKind().GroupVersionKind().Kind,
			}, util.NamespacedName(obj).String(),
		)
	}

	return r.Create(ctx, obj)
}

// Patch patches the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (r *FakeClient) Patch(
	ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption,
) error {
	// TODO (covariance) implement me!
	panic("not implemented")
}

// DeleteAllOf deletes all objects of the given type matching the given options.
func (r *FakeClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	// TODO (covariance) implement me!
	panic("not implemented")
}

// Status client knows how to create a client which can update status subresource
// for kubernetes objects.
func (r *FakeClient) Status() client.StatusWriter {
	return &FakeStatusWriter{}
}

type FakeStatusWriter struct{}

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
func (r *FakeStatusWriter) Patch(
	ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption,
) error {
	// TODO (covariance) implement me!
	panic("not implemented")
}
