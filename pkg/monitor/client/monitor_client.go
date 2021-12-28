/*
Copyright 2021 The KubeVela Authors.

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

package client

import (
	"context"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"

	"github.com/oam-dev/kubevela/pkg/monitor/metrics"
	"github.com/oam-dev/kubevela/pkg/multicluster"
)

type MonitorClient interface {
	client.Client
}

type monitorClient struct {
	client.Client
}

func DefaultNewMonitorClient(cache cache.Cache, config *rest.Config, options client.Options, uncachedObjects ...client.Object) (client.Client, error) {
	c, err := cluster.DefaultNewClient(cache, config, options, uncachedObjects...)
	if err != nil {
		return nil, err
	}
	return &monitorClient{c}, nil
}

func monitor(ctx context.Context, verb string, obj runtime.Object) func() {
	o := obj.GetObjectKind().GroupVersionKind()
	_, isUnstructured := obj.(*unstructured.Unstructured)
	_, isUnstructuredList := obj.(*unstructured.UnstructuredList)
	un := "false"
	if isUnstructured || isUnstructuredList {
		un = "true"
	}
	clusterName := multicluster.ClusterNameInContext(ctx)
	if clusterName == "" {
		clusterName = multicluster.ClusterLocalName
	}
	kind := strings.TrimSuffix(o.Kind, "List")
	begin := time.Now()
	return func() {
		v := time.Now().Sub(begin).Seconds()
		metrics.ClientRequestHistogram.WithLabelValues(verb, kind, o.GroupVersion().String(), un, clusterName).Observe(v)
	}
}

func (c *monitorClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	cb := monitor(ctx, "Get", obj)
	defer cb()
	return c.Client.Get(ctx, key, obj)
}

func (c *monitorClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	cb := monitor(ctx, "List", list)
	defer cb()
	return c.Client.List(ctx, list, opts...)
}

func (c *monitorClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	cb := monitor(ctx, "Create", obj)
	defer cb()
	return c.Client.Create(ctx, obj, opts...)
}

func (c *monitorClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	cb := monitor(ctx, "Delete", obj)
	defer cb()
	return c.Client.Delete(ctx, obj, opts...)
}

func (c *monitorClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	cb := monitor(ctx, "Update", obj)
	defer cb()
	return c.Client.Update(ctx, obj, opts...)
}

func (c *monitorClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	cb := monitor(ctx, "Patch", obj)
	defer cb()
	return c.Client.Patch(ctx, obj, patch, opts...)
}

func (c *monitorClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	cb := monitor(ctx, "DeleteAllOf", obj)
	defer cb()
	return c.Client.DeleteAllOf(ctx, obj, opts...)
}

func (c *monitorClient) Status() client.StatusWriter {
	return &monitorStatusWriter{c.Status()}
}

type monitorStatusWriter struct {
	client.StatusWriter
}

func (w *monitorStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	cb := monitor(ctx, "StatusUpdate", obj)
	defer cb()
	return w.StatusWriter.Update(ctx, obj, opts...)
}

func (w *monitorStatusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	cb := monitor(ctx, "StatusPatch", obj)
	defer cb()
	return w.StatusWriter.Patch(ctx, obj, patch, opts...)
}