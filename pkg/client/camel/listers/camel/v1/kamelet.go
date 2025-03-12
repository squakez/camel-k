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

// KameletLister helps list Kamelets.
// All objects returned here must be treated as read-only.
type KameletLister interface {
	// List lists all Kamelets in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*camelv1.Kamelet, err error)
	// Kamelets returns an object that can list and get Kamelets.
	Kamelets(namespace string) KameletNamespaceLister
	KameletListerExpansion
}

// kameletLister implements the KameletLister interface.
type kameletLister struct {
	listers.ResourceIndexer[*camelv1.Kamelet]
}

// NewKameletLister returns a new KameletLister.
func NewKameletLister(indexer cache.Indexer) KameletLister {
	return &kameletLister{listers.New[*camelv1.Kamelet](indexer, camelv1.Resource("kamelet"))}
}

// Kamelets returns an object that can list and get Kamelets.
func (s *kameletLister) Kamelets(namespace string) KameletNamespaceLister {
	return kameletNamespaceLister{listers.NewNamespaced[*camelv1.Kamelet](s.ResourceIndexer, namespace)}
}

// KameletNamespaceLister helps list and get Kamelets.
// All objects returned here must be treated as read-only.
type KameletNamespaceLister interface {
	// List lists all Kamelets in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*camelv1.Kamelet, err error)
	// Get retrieves the Kamelet from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*camelv1.Kamelet, error)
	KameletNamespaceListerExpansion
}

// kameletNamespaceLister implements the KameletNamespaceLister
// interface.
type kameletNamespaceLister struct {
	listers.ResourceIndexer[*camelv1.Kamelet]
}
