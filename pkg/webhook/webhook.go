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
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func RegisterForManager(
	manager manager.Manager,
	obj runtime.Object,
	handler admission.Handler,
	webhookType string,
) error {
	log := manager.GetLogger()

	gvk, err := apiutil.GVKForObject(obj, manager.GetScheme())
	if err != nil {
		return fmt.Errorf("unable to get GVK of object: %w", err)
	}

	path := generatePath(webhookType, gvk)
	if isPathHandled(path, manager.GetWebhookServer()) {
		log.Info("webhook path is already handled")
		return nil
	}

	if casted, ok := handler.(ObjectInjector); ok {
		if err := casted.InjectObject(obj); err != nil {
			return fmt.Errorf("unable to inject object into handler: %w", err)
		}
	}

	wh := admission.Webhook{
		Handler: handler,
	}

	if err := wh.InjectScheme(manager.GetScheme()); err != nil {
		return fmt.Errorf("unable to inject scheme into wh: %w", err)
	}

	if err := wh.InjectLogger(manager.GetLogger()); err != nil {
		return fmt.Errorf("unable to inject logger into wh: %w", err)
	}

	manager.GetWebhookServer().Register(path, &wh)
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
