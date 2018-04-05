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

package controller

import (
	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

	"github.com/kubernetes-incubator/service-catalog/pkg/pretty"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

// Service plan handlers and control-loop

func (c *controller) servicePlanAdd(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		pcb := pretty.NewContextBuilder(pretty.ServicePlan, "", "")
		glog.Errorf(pcb.Messagef("Couldn't get key for object %+v: %v", obj, err))
		return
	}
	c.servicePlanQueue.Add(key)
}

func (c *controller) servicePlanUpdate(oldObj, newObj interface{}) {
	c.servicePlanAdd(newObj)
}

func (c *controller) servicePlanDelete(obj interface{}) {
	servicePlan, ok := obj.(*v1beta1.ServicePlan)
	if servicePlan == nil || !ok {
		return
	}

	pcb := pretty.NewContextBuilder(pretty.ServicePlan, servicePlan.ObjectMeta.Namespace, servicePlan.ObjectMeta.Name)
	glog.V(4).Infof(pcb.Message("Received delete event, no further processing will occur"))
}

// reconcileServicePlanKey reconciles a ServicePlan due to resync
// or an event on the ServicePlan.  Note that this is NOT the main
// reconciliation loop for ServicePlans. ServicePlans are
// primarily reconciled in a separate flow when a ServiceBroker is
// reconciled.
func (c *controller) reconcileServicePlanKey(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	pcb := pretty.NewContextBuilder(pretty.ServicePlan, namespace, name)
	plan, err := c.servicePlanLister.ServicePlans(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.Info(pcb.Message("Not doing work because the ServicePlan has been deleted"))
		return nil
	}
	if err != nil {
		glog.Info(pcb.Messagef("Unable to retrieve ServicePlan: %v", err))
		return err
	}

	return c.reconcileServicePlan(plan)
}

func (c *controller) reconcileServicePlan(servicePlan *v1beta1.ServicePlan) error {
	pcb := pretty.NewContextBuilder(pretty.ServicePlan, servicePlan.ObjectMeta.Namespace, servicePlan.ObjectMeta.Name)
	glog.Info(pcb.Message("processing"))

	if !servicePlan.Status.RemovedFromBrokerCatalog {
		return nil
	}

	glog.Infof(pcb.Message("ServicePlan has been removed from the broker catalog; determining whether there are instances remaining"))

	serviceInstances, err := c.findServiceInstancesOnServicePlan(servicePlan)
	if err != nil {
		return err
	}

	if len(serviceInstances.Items) != 0 {
		return nil
	}

	glog.Info(pcb.Message("ServicePlan has been removed from the broker catalog and has zero instances remaining; deleting"))
	ns := servicePlan.ObjectMeta.Namespace
	return c.serviceCatalogClient.ServicePlans(ns).Delete(servicePlan.Name, &metav1.DeleteOptions{})
}

func (c *controller) findServiceInstancesOnServicePlan(servicePlan *v1beta1.ServicePlan) (*v1beta1.ServiceInstanceList, error) {
	// ERIK TODO: Need to enhance service instances to support ns service plans?
	fieldSet := fields.Set{
		"spec.servicePlanRef.name": servicePlan.Name,
	}
	fieldSelector := fields.SelectorFromSet(fieldSet).String()
	listOpts := metav1.ListOptions{FieldSelector: fieldSelector}

	return c.serviceCatalogClient.ServiceInstances(metav1.NamespaceAll).List(listOpts)
}
