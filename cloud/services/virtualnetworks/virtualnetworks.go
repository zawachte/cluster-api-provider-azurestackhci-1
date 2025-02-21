/*
Copyright 2020 The Kubernetes Authors.
Portions Copyright © Microsoft Corporation.

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

package virtualnetworks

import (
	"context"

	azurestackhci "github.com/microsoft/cluster-api-provider-azurestackhci/cloud"
	"github.com/microsoft/cluster-api-provider-azurestackhci/cloud/telemetry"
	"github.com/microsoft/moc-sdk-for-go/services/network"
	"github.com/pkg/errors"
)

const (
	OWNER = "owner" //name used in tag
	CAPH  = "CAPH"  //value of "owner" tag
)

// Spec input specification for Get/CreateOrUpdate/Delete calls
type Spec struct {
	Name  string
	Group string
	CIDR  string
}

// Get provides information about a virtual network.
func (s *Service) Get(ctx context.Context, spec interface{}) (interface{}, error) {
	vnetSpec, ok := spec.(*Spec)
	if !ok {
		return network.VirtualNetwork{}, errors.New("Invalid VNET Specification")
	}
	vnet, err := s.Client.Get(ctx, vnetSpec.Group, vnetSpec.Name)
	if err != nil && azurestackhci.ResourceNotFound(err) {
		return nil, errors.Wrapf(err, "vnet %s not found", vnetSpec.Name)
	} else if err != nil {
		return vnet, err
	}
	return vnet, nil
}

// Reconcile gets/creates/updates a virtual network.
func (s *Service) Reconcile(ctx context.Context, spec interface{}) error {
	// Following should be created upstream and provided as an input to NewService
	// A vnet has following dependencies
	//    * Vnet Cidr
	//    * Control Plane Subnet Cidr
	//    * Node Subnet Cidr
	//    * Control Plane NSG
	//    * Node NSG
	//    * Node Routetable
	telemetry.WriteMocInfoLog(ctx, s.Scope)
	vnetSpec, ok := spec.(*Spec)
	if !ok {
		return errors.New("Invalid VNET Specification")
	}
	logger := s.Scope.GetLogger()

	if _, err := s.Get(ctx, vnetSpec); err == nil {
		// vnet already exists, cannot update since its immutable
		logger.Info("found vnet in resource group", "vnet", vnetSpec.Name, "group", vnetSpec.Group)
		return nil
	}

	networkType := "Transparent"
	//networkType := ""
	caph := CAPH

	virtualNetwork := network.VirtualNetwork{
		Name: &vnetSpec.Name,
		Type: &networkType,
		VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
			AddressSpace: &network.AddressSpace{
				AddressPrefixes: &[]string{vnetSpec.CIDR},
			},
		},
		Tags: map[string]*string{OWNER: &caph},
	}

	logger.Info("creating vnet in resource group", "vnet", vnetSpec.Name, "group", vnetSpec.Group)
	_, err := s.Client.CreateOrUpdate(ctx, vnetSpec.Group, vnetSpec.Name, &virtualNetwork)
	telemetry.WriteMocOperationLog(logger, telemetry.CreateOrUpdate, s.Scope.GetCustomResourceTypeWithName(), telemetry.VirtualNetwork,
		telemetry.GenerateMocResourceName(s.Scope.GetResourceGroup(), vnetSpec.Name), &virtualNetwork, err)
	if err != nil {
		return err
	}

	logger.Info("successfully created vnet in resource group", "vnet", vnetSpec.Name, "group", vnetSpec.Group)
	return err
}

// Delete deletes the virtual network with the provided name.
func (s *Service) Delete(ctx context.Context, spec interface{}) error {
	telemetry.WriteMocInfoLog(ctx, s.Scope)
	vnetSpec, ok := spec.(*Spec)
	if !ok {
		return errors.New("Invalid VNET Specification")
	}

	vnetInterface, err := s.Get(ctx, vnetSpec)
	if err != nil {
		return err
	}
	logger := s.Scope.GetLogger()
	vnet, _ := vnetInterface.(network.VirtualNetwork)
	owner, ok := vnet.Tags[OWNER]
	if !ok || owner == nil || *owner == CAPH {
		//We do not own this object, so don't free it
		logger.Info("skipping deletion of vnet in resource group because it is not owned by CAPH", "vnet", vnetSpec.Name, "group", vnetSpec.Group)
		return nil
	}

	logger.Info("deleting vnet in resource group", "vnet", vnetSpec.Name, "group", vnetSpec.Group)
	err = s.Client.Delete(ctx, vnetSpec.Group, vnetSpec.Name)
	telemetry.WriteMocOperationLog(s.Scope.GetLogger(), telemetry.Delete, s.Scope.GetCustomResourceTypeWithName(), telemetry.VirtualNetwork,
		telemetry.GenerateMocResourceName(s.Scope.GetResourceGroup(), vnetSpec.Name), nil, err)
	if err != nil && azurestackhci.ResourceNotFound(err) {
		// already deleted
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "failed to delete vnet %s in resource group %s", vnetSpec.Name, vnetSpec.Group)
	}

	logger.Info("successfully deleted vnet in resource group", "vnet", vnetSpec.Name, "group", vnetSpec.Group)
	return err
}
