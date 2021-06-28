// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func RegisterForManager(
	mgr manager.Manager,
	obj runtime.Object,
	handler admission.Handler,
	webhookType string,
) error {
	log := mgr.GetLogger()

	gvk, err := apiutil.GVKForObject(obj, mgr.GetScheme())
	if err != nil {
		return fmt.Errorf("unable to get GVK of object: %w", err)
	}

	path := generatePath(webhookType, gvk)
	if isPathHandled(path, mgr.GetWebhookServer()) {
		log.Info("webhook path is already handled")
		return nil
	}

	if casted, ok := handler.(ObjectInjector); ok {
		if err := casted.InjectObject(obj); err != nil {
			return fmt.Errorf("unable to inject object into handler: %w", err)
		}
	}

	if casted, ok := handler.(inject.Logger); ok {
		if err := casted.InjectLogger(mgr.GetLogger()); err != nil {
			return fmt.Errorf("unable to inject logger into handler: %w", err)
		}
	}

	wh := admission.Webhook{
		Handler: handler,
	}

	if err := wh.InjectScheme(mgr.GetScheme()); err != nil {
		return fmt.Errorf("unable to inject scheme into wh: %w", err)
	}

	if err := wh.InjectLogger(mgr.GetLogger()); err != nil {
		return fmt.Errorf("unable to inject logger into wh: %w", err)
	}

	mgr.GetWebhookServer().Register(path, &wh)
	log.Info("webhook registered for " + path)
	return nil
}

func isPathHandled(path string, ws *webhook.Server) bool {
	if ws.WebhookMux == nil {
		return false
	}
	h, p := ws.WebhookMux.Handler(&http.Request{URL: &url.URL{Path: path}})
	if p == path && h != nil {
		return true
	}
	return false
}

func generatePath(webhookType string, gvk schema.GroupVersionKind) string {
	return "/" + webhookType + "-" + strings.Replace(gvk.Group, ".", "-", -1) + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind)
}
