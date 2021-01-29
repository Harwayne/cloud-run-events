/*
Copyright 2021 Google LLC

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

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/google/knative-gcp/pkg/apis/events/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// CloudBuildSourceLister helps list CloudBuildSources.
// All objects returned here must be treated as read-only.
type CloudBuildSourceLister interface {
	// List lists all CloudBuildSources in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.CloudBuildSource, err error)
	// CloudBuildSources returns an object that can list and get CloudBuildSources.
	CloudBuildSources(namespace string) CloudBuildSourceNamespaceLister
	CloudBuildSourceListerExpansion
}

// cloudBuildSourceLister implements the CloudBuildSourceLister interface.
type cloudBuildSourceLister struct {
	indexer cache.Indexer
}

// NewCloudBuildSourceLister returns a new CloudBuildSourceLister.
func NewCloudBuildSourceLister(indexer cache.Indexer) CloudBuildSourceLister {
	return &cloudBuildSourceLister{indexer: indexer}
}

// List lists all CloudBuildSources in the indexer.
func (s *cloudBuildSourceLister) List(selector labels.Selector) (ret []*v1.CloudBuildSource, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.CloudBuildSource))
	})
	return ret, err
}

// CloudBuildSources returns an object that can list and get CloudBuildSources.
func (s *cloudBuildSourceLister) CloudBuildSources(namespace string) CloudBuildSourceNamespaceLister {
	return cloudBuildSourceNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// CloudBuildSourceNamespaceLister helps list and get CloudBuildSources.
// All objects returned here must be treated as read-only.
type CloudBuildSourceNamespaceLister interface {
	// List lists all CloudBuildSources in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.CloudBuildSource, err error)
	// Get retrieves the CloudBuildSource from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.CloudBuildSource, error)
	CloudBuildSourceNamespaceListerExpansion
}

// cloudBuildSourceNamespaceLister implements the CloudBuildSourceNamespaceLister
// interface.
type cloudBuildSourceNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all CloudBuildSources in the indexer for a given namespace.
func (s cloudBuildSourceNamespaceLister) List(selector labels.Selector) (ret []*v1.CloudBuildSource, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.CloudBuildSource))
	})
	return ret, err
}

// Get retrieves the CloudBuildSource from the indexer for a given namespace and name.
func (s cloudBuildSourceNamespaceLister) Get(name string) (*v1.CloudBuildSource, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("cloudbuildsource"), name)
	}
	return obj.(*v1.CloudBuildSource), nil
}
