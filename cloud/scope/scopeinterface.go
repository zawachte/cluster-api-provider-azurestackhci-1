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

package scope

import (
	"github.com/go-logr/logr"
	"github.com/microsoft/moc/pkg/auth"
)

// ScopeInterface allows multiple scope types to be used for cloud services
type ScopeInterface interface {
	GetResourceGroup() string
	GetCloudAgentFqdn() string
	GetAuthorizer() auth.Authorizer
	GetCustomResourceTypeWithName() string
	GetLogger() logr.Logger
}
