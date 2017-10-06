/*
Copyright 2017 The Kubernetes Authors.

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

package v1alpha1

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
)

// ClusterServiceBrokerExpansion interface allows relisting a ClusterServiceBroker
type ClusterServiceBrokerExpansion interface {
	Relist(servicebroker *v1alpha1.ClusterServiceBroker) (*v1alpha1.ClusterServiceBroker, error)
}

func (c *clusterServiceBrokers) Relist(serviceBroker *v1alpha1.ClusterServiceBroker) (result *v1alpha1.ClusterServiceBroker, err error) {
	result = &v1alpha1.ClusterServiceBroker{}
	err = c.client.Put().
		Namespace(serviceBroker.Namespace).
		Resource("clusterservicebrokers").
		Name(serviceBroker.Name).
		SubResource("relist").
		Body(serviceBroker).
		Do().
		Into(result)
	return
}
