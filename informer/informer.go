package informer

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	labels "k8s.io/apimachinery/pkg/labels"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
)

type ResourceHandler func(interface{})

type RancherInformer struct {
	informer cache.SharedIndexInformer
	stopCh   chan struct{}
}

type InformerManager struct {
	client       dynamic.Interface
	informers    map[string]*RancherInformer
	mu           sync.RWMutex
	stopCh       chan struct{}
	resyncPeriod time.Duration
}

func NewInformerManager(client dynamic.Interface, resyncPeriod int) *InformerManager {
	return &InformerManager{
		client:       client,
		informers:    make(map[string]*RancherInformer),
		stopCh:       make(chan struct{}),
		resyncPeriod: time.Duration(resyncPeriod) * time.Second,
	}
}

func (m *InformerManager) AddEventHandler(resourceName string, handler ResourceHandler) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	informer, ok := m.informers[resourceName]
	if !ok {
		return
	}

	informer.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: handler,
		UpdateFunc: func(oldObj, newObj interface{}) {
			handler(newObj)
		},
		DeleteFunc: handler,
	})
}

func (m *InformerManager) Start() error {
	m.mu.Lock()

	for _, informer := range m.informers {
		go informer.informer.Run(informer.stopCh)
	}

	m.mu.Unlock()

	if !cache.WaitForCacheSync(m.stopCh, m.hasSyncedFuncs()...) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	return nil
}

func (m *InformerManager) hasSyncedFuncs() []cache.InformerSynced {
	m.mu.RLock()
	defer m.mu.RUnlock()

	syncedFuncs := make([]cache.InformerSynced, 0, len(m.informers))
	for _, informer := range m.informers {
		syncedFuncs = append(syncedFuncs, informer.informer.HasSynced)
	}
	return syncedFuncs
}

func (m *InformerManager) CreateInformer(resourceName string, group string, version string, resource string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.informers[resourceName]; exists {
		return nil
	}

	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	_, err := m.client.Resource(gvr).List(context.Background(), metav1.ListOptions{Limit: 1})
	if err != nil {
		return fmt.Errorf("resource %s (%s/%s/%s) not available: %w", resourceName, group, version, resource, err)
	}

	listWatch := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return m.client.Resource(gvr).List(context.Background(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			w, err := m.client.Resource(gvr).Watch(context.Background(), options)
			if err != nil {
				log.Debugf("Resource %s does not support watch, falling back to list-only mode", resourceName)
				return watch.NewFake(), nil
			}
			return w, nil
		},
	}

	informer := cache.NewSharedIndexInformer(
		listWatch,
		&unstructured.Unstructured{},
		m.resyncPeriod,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	informer.SetWatchErrorHandler(func(r *cache.Reflector, err error) {
		log.Errorf("Watch error for resource %s (%s/%s): %v", resourceName, gvr.Group, gvr.Resource, err)
	})

	m.informers[resourceName] = &RancherInformer{
		informer: informer,
		stopCh:   make(chan struct{}),
	}
	return nil
}

func (m *InformerManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, informer := range m.informers {
		close(informer.stopCh)
	}
	close(m.stopCh)
}

func (m *InformerManager) GetLister(resourceName string, namespace string) cache.GenericLister {
	m.mu.RLock()
	defer m.mu.RUnlock()

	informer, ok := m.informers[resourceName]
	if !ok {
		return nil
	}
	return &genericLister{
		indexer:   informer.informer.GetIndexer(),
		namespace: namespace,
	}
}

type genericLister struct {
	indexer   cache.Indexer
	namespace string
}

type genericNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

func (l *genericLister) List(selector labels.Selector) ([]runtime.Object, error) {
	var result []runtime.Object

	if l.namespace != "" {
		objs, err := l.indexer.ByIndex(cache.NamespaceIndex, l.namespace)
		if err != nil {
			return nil, err
		}
		for _, obj := range objs {
			result = append(result, obj.(runtime.Object))
		}
	} else {
		objs := l.indexer.List()
		for _, obj := range objs {
			result = append(result, obj.(runtime.Object))
		}
	}
	return result, nil
}

func (l *genericLister) Get(name string) (runtime.Object, error) {
	key := name
	if l.namespace != "" {
		key = l.namespace + "/" + name
	}
	obj, exists, err := l.indexer.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("resource %s not found", name)
	}
	return obj.(runtime.Object), nil
}

func (l *genericLister) ByNamespace(ns string) cache.GenericNamespaceLister {
	return &genericNamespaceLister{
		indexer:   l.indexer,
		namespace: ns,
	}
}

func (l *genericNamespaceLister) List(selector labels.Selector) ([]runtime.Object, error) {
	objs, err := l.indexer.ByIndex(cache.NamespaceIndex, l.namespace)
	if err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, obj := range objs {
		result = append(result, obj.(runtime.Object))
	}
	return result, nil
}

func (l *genericNamespaceLister) Get(name string) (runtime.Object, error) {
	key := l.namespace + "/" + name
	obj, exists, err := l.indexer.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("resource %s not found", name)
	}
	return obj.(runtime.Object), nil
}

func (l *genericLister) ByIndex(indexName, keyValue string) ([]runtime.Object, error) {
	objs, err := l.indexer.ByIndex(indexName, keyValue)
	if err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, obj := range objs {
		result = append(result, obj.(runtime.Object))
	}
	return result, nil
}
