// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"context"
	"fmt"

	"k8s-connectors/pkg/errorhandling"

	"github.com/go-logr/logr"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "k8s-connectors/connector/ycr/api/v1"
	"k8s-connectors/pkg/util"
	"k8s-connectors/pkg/webhook"
)

// +kubebuilder:webhook:path=/validate-connectors-cloud-yandex-com-v1-yandexcontainerregistry,mutating=false,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=yandexcontainerregistries,verbs=create;update;delete,versions=v1,name=vyandexcontainerregistry.yandex.com,admissionReviewVersions=v1

type YCRValidator struct {
	sdk *ycsdk.SDK
}

func NewYCRValidator(ctx context.Context) (webhook.Validator, error) {
	sdk, err := ycsdk.Build(
		ctx,
		ycsdk.Config{
			Credentials: ycsdk.InstanceServiceAccount(),
		},
	)
	if err != nil {
		return nil, err
	}
	return &YCRValidator{sdk: sdk}, nil
}

func (r *YCRValidator) ValidateCreation(ctx context.Context, log logr.Logger, obj runtime.Object) error {
	casted := obj.(*v1.YandexContainerRegistry)
	log.Info("validate create", "name", util.NamespacedName(casted))

	if _, err := r.sdk.ResourceManager().Folder().Get(
		ctx, &resourcemanager.GetFolderRequest{
			FolderId: casted.Spec.FolderID,
		},
	); err != nil {
		if errorhandling.CheckRPCErrorNotFound(err) {
			return webhook.NewValidationErrorf("folder %s cannot be found in the cloud", casted.Spec.FolderID)
		}
		return fmt.Errorf("unable to get folder: %w", err)
	}

	return nil
}

func (r *YCRValidator) ValidateUpdate(_ context.Context, log logr.Logger, current, old runtime.Object) error {
	castedOld, castedCurrent := old.(*v1.YandexContainerRegistry), current.(*v1.YandexContainerRegistry)

	log.Info("validate update", "name", util.NamespacedName(castedCurrent))

	if castedCurrent.Spec.FolderID != castedOld.Spec.FolderID {
		return webhook.NewValidationErrorf(
			"folder id must be immutable, was changed from %s to %s",
			castedOld.Spec.FolderID,
			castedCurrent.Spec.FolderID,
		)
	}

	return nil
}

func (r *YCRValidator) ValidateDeletion(ctx context.Context, log logr.Logger, obj runtime.Object) error {
	casted := obj.(*v1.YandexContainerRegistry)
	log.Info("validate delete", "name", util.NamespacedName(casted))

	resp, err := r.sdk.ContainerRegistry().Image().List(
		ctx, &containerregistry.ListImagesRequest{
			RegistryId: casted.Status.ID,
			FolderId:   casted.Spec.FolderID,
			PageSize:   1, // We only need to check if there are ANY number of images
		},
	)
	if err != nil {
		return fmt.Errorf("unable to list images in container registry: %w", err)
	}

	if len(resp.Images) != 0 {
		return webhook.NewValidationErrorf("cannot delete non-empty registry %s", casted.Spec.Name)
	}

	return nil
}
