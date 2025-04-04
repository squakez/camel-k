/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

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
	camelv1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	listers "k8s.io/client-go/listers"
	cache "k8s.io/client-go/tools/cache"
)

// CamelCatalogLister helps list CamelCatalogs.
// All objects returned here must be treated as read-only.
type CamelCatalogLister interface {
	// List lists all CamelCatalogs in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*camelv1.CamelCatalog, err error)
	// CamelCatalogs returns an object that can list and get CamelCatalogs.
	CamelCatalogs(namespace string) CamelCatalogNamespaceLister
	CamelCatalogListerExpansion
}

// camelCatalogLister implements the CamelCatalogLister interface.
type camelCatalogLister struct {
	listers.ResourceIndexer[*camelv1.CamelCatalog]
}

// NewCamelCatalogLister returns a new CamelCatalogLister.
func NewCamelCatalogLister(indexer cache.Indexer) CamelCatalogLister {
	return &camelCatalogLister{listers.New[*camelv1.CamelCatalog](indexer, camelv1.Resource("camelcatalog"))}
}

// CamelCatalogs returns an object that can list and get CamelCatalogs.
func (s *camelCatalogLister) CamelCatalogs(namespace string) CamelCatalogNamespaceLister {
	return camelCatalogNamespaceLister{listers.NewNamespaced[*camelv1.CamelCatalog](s.ResourceIndexer, namespace)}
}

// CamelCatalogNamespaceLister helps list and get CamelCatalogs.
// All objects returned here must be treated as read-only.
type CamelCatalogNamespaceLister interface {
	// List lists all CamelCatalogs in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*camelv1.CamelCatalog, err error)
	// Get retrieves the CamelCatalog from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*camelv1.CamelCatalog, error)
	CamelCatalogNamespaceListerExpansion
}

// camelCatalogNamespaceLister implements the CamelCatalogNamespaceLister
// interface.
type camelCatalogNamespaceLister struct {
	listers.ResourceIndexer[*camelv1.CamelCatalog]
}
