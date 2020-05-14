// Copyright 2017-2020 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"context"
	goerrors "errors"
	"fmt"
	"time"

	"sigs.k8s.io/yaml"

	k8sconst "github.com/cilium/cilium/pkg/k8s/apis/cilium.io"
	"github.com/cilium/cilium/pkg/logging"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/option"
	"github.com/cilium/cilium/pkg/versioncheck"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	// subsysK8s is the value for logfields.LogSubsys
	subsysK8s = "k8s"
)

var (
	// log is the k8s package logger object.
	log = logging.DefaultLogger.WithField(logfields.LogSubsys, subsysK8s)
)

const (
	// CustomResourceDefinitionGroup is the name of the third party resource group
	CustomResourceDefinitionGroup = k8sconst.GroupName

	// CustomResourceDefinitionVersion is the current version of the resource
	CustomResourceDefinitionVersion = "v2"

	// CustomResourceDefinitionSchemaVersion is semver-conformant version of CRD schema
	// Used to determine if CRD needs to be updated in cluster
	CustomResourceDefinitionSchemaVersion = "1.18"

	// CustomResourceDefinitionSchemaVersionKey is key to label which holds the CRD schema version
	CustomResourceDefinitionSchemaVersionKey = "io.cilium.k8s.crd.schema.version"

	// CNPKindDefinition is the kind name for Cilium Network Policy
	CNPKindDefinition = "CiliumNetworkPolicy"

	fqdnNameRegex = `^([-a-zA-Z0-9_]+[.]?)+$`

	fqdnPatternRegex = `^([-a-zA-Z0-9_*]+[.]?)+$`

	// CCNPKindDefinition is the kind name for Cilium Cluster wide Network Policy
	CCNPKindDefinition = "CiliumClusterwideNetworkPolicy"

	// CNKindDefinition is the kind name for Cilium Node
	CNKindDefinition = "CiliumNode"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{
	Group:   CustomResourceDefinitionGroup,
	Version: CustomResourceDefinitionVersion,
}

var (
	comparableCRDSchemaVersion = versioncheck.MustVersion(CustomResourceDefinitionSchemaVersion)
)

// CreateCustomResourceDefinitions creates our CRD objects in the kubernetes
// cluster
func CreateCustomResourceDefinitions(clientset apiextensionsclient.Interface) error {
	if err := createCNPCRD(clientset); err != nil {
		return err
	}

	if err := createCCNPCRD(clientset); err != nil {
		return err
	}

	if err := createCEPCRD(clientset); err != nil {
		return err
	}

	if err := createNodeCRD(clientset); err != nil {
		return err
	}

	if option.Config.IdentityAllocationMode == option.IdentityAllocationModeCRD {
		if err := createIdentityCRD(clientset); err != nil {
			return err
		}
	}

	return nil
}

// createCNPCRD creates and updates the CiliumNetworkPolicies CRD. It should be called
// on agent startup but is idempotent and safe to call again.
func createCNPCRD(clientset apiextensionsclient.Interface) error {
	crdBytes, err := examplesCrdsCiliumnetworkpoliciesYamlBytes()
	if err != nil {
		panic(err)
	}
	ciliumCRD := apiextensionsv1beta1.CustomResourceDefinition{}
	err = yaml.Unmarshal(crdBytes, &ciliumCRD)
	if err != nil {
		panic(err)
	}

	res := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: ciliumCRD.Spec.Names.Plural + "." + SchemeGroupVersion.Group,
			Labels: map[string]string{
				CustomResourceDefinitionSchemaVersionKey: CustomResourceDefinitionSchemaVersion,
			},
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   SchemeGroupVersion.Group,
			Version: SchemeGroupVersion.Version,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:     ciliumCRD.Spec.Names.Plural,
				Singular:   ciliumCRD.Spec.Names.Singular,
				ShortNames: ciliumCRD.Spec.Names.ShortNames,
				Kind:       ciliumCRD.Spec.Names.Kind,
			},
			AdditionalPrinterColumns: ciliumCRD.Spec.AdditionalPrinterColumns,
			Subresources:             ciliumCRD.Spec.Subresources,
			Scope:                    ciliumCRD.Spec.Scope,
			Validation:               ciliumCRD.Spec.Validation,
		},
	}

	return createUpdateCRD(clientset, "CiliumNetworkPolicy/v2", res)
}

// createCGNPCRD creates and updates the CiliumGlobalNetworkPolicies CRD. It should be called
// on agent startup but is idempotent and safe to call again.
func createCCNPCRD(clientset apiextensionsclient.Interface) error {
	crdBytes, err := examplesCrdsCiliumclusterwidenetworkpoliciesYamlBytes()
	if err != nil {
		panic(err)
	}
	ciliumCRD := apiextensionsv1beta1.CustomResourceDefinition{}
	err = yaml.Unmarshal(crdBytes, &ciliumCRD)
	if err != nil {
		panic(err)
	}

	res := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: ciliumCRD.Spec.Names.Plural + "." + SchemeGroupVersion.Group,
			Labels: map[string]string{
				CustomResourceDefinitionSchemaVersionKey: CustomResourceDefinitionSchemaVersion,
			},
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   SchemeGroupVersion.Group,
			Version: SchemeGroupVersion.Version,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:     ciliumCRD.Spec.Names.Plural,
				Singular:   ciliumCRD.Spec.Names.Singular,
				ShortNames: ciliumCRD.Spec.Names.ShortNames,
				Kind:       ciliumCRD.Spec.Names.Kind,
			},
			Subresources: ciliumCRD.Spec.Subresources,
			Scope:        ciliumCRD.Spec.Scope,
			Validation:   ciliumCRD.Spec.Validation,
		},
	}

	return createUpdateCRD(clientset, "CiliumClusterwideNetworkPolicy/v2", res)
}

// createCEPCRD creates and updates the CiliumEndpoint CRD. It should be called
// on agent startup but is idempotent and safe to call again.
func createCEPCRD(clientset apiextensionsclient.Interface) error {
	crdBytes, err := examplesCrdsCiliumendpointsYamlBytes()
	if err != nil {
		panic(err)
	}
	ciliumCRD := apiextensionsv1beta1.CustomResourceDefinition{}
	err = yaml.Unmarshal(crdBytes, &ciliumCRD)
	if err != nil {
		panic(err)
	}

	res := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: ciliumCRD.Spec.Names.Plural + "." + SchemeGroupVersion.Group,
			// TODO(christarazi): Remove me because this is just a workaround
			Labels: map[string]string{
				CustomResourceDefinitionSchemaVersionKey: CustomResourceDefinitionSchemaVersion,
			},
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   SchemeGroupVersion.Group,
			Version: SchemeGroupVersion.Version,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:     ciliumCRD.Spec.Names.Plural,
				Singular:   ciliumCRD.Spec.Names.Singular,
				ShortNames: ciliumCRD.Spec.Names.ShortNames,
				Kind:       ciliumCRD.Spec.Names.Kind,
			},
			AdditionalPrinterColumns: ciliumCRD.Spec.AdditionalPrinterColumns,
			Subresources:             ciliumCRD.Spec.Subresources,
			Scope:                    ciliumCRD.Spec.Scope,
			Validation:               ciliumCRD.Spec.Validation,
		},
	}

	return createUpdateCRD(clientset, "v2.CiliumEndpoint", res)
}

// createNodeCRD creates and updates the CiliumNode CRD. It should be called on
// agent startup but is idempotent and safe to call again.
func createNodeCRD(clientset apiextensionsclient.Interface) error {
	crdBytes, err := examplesCrdsCiliumnodesYamlBytes()
	if err != nil {
		panic(err)
	}
	ciliumCRD := apiextensionsv1beta1.CustomResourceDefinition{}
	err = yaml.Unmarshal(crdBytes, &ciliumCRD)
	if err != nil {
		panic(err)
	}

	res := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: ciliumCRD.Spec.Names.Plural + "." + SchemeGroupVersion.Group,
			// TODO(christarazi): Remove me because this is just a workaround
			Labels: map[string]string{
				CustomResourceDefinitionSchemaVersionKey: CustomResourceDefinitionSchemaVersion,
			},
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   SchemeGroupVersion.Group,
			Version: SchemeGroupVersion.Version,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:     ciliumCRD.Spec.Names.Plural,
				Singular:   ciliumCRD.Spec.Names.Singular,
				ShortNames: ciliumCRD.Spec.Names.ShortNames,
				Kind:       ciliumCRD.Spec.Names.Kind,
			},
			Subresources: ciliumCRD.Spec.Subresources,
			Scope:        ciliumCRD.Spec.Scope,
			Validation:   ciliumCRD.Spec.Validation,
		},
	}

	return createUpdateCRD(clientset, "v2.CiliumNode", res)
}

// createIdentityCRD creates and updates the CiliumIdentity CRD. It should be
// called on agent startup but is idempotent and safe to call again.
func createIdentityCRD(clientset apiextensionsclient.Interface) error {
	crdBytes, err := examplesCrdsCiliumidentitiesYamlBytes()
	if err != nil {
		panic(err)
	}
	ciliumCRD := apiextensionsv1beta1.CustomResourceDefinition{}
	err = yaml.Unmarshal(crdBytes, &ciliumCRD)
	if err != nil {
		panic(err)
	}

	res := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: ciliumCRD.Spec.Names.Plural + "." + SchemeGroupVersion.Group,
			// TODO(christarazi): Remove me because this is just a workaround
			Labels: map[string]string{
				CustomResourceDefinitionSchemaVersionKey: CustomResourceDefinitionSchemaVersion,
			},
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   SchemeGroupVersion.Group,
			Version: SchemeGroupVersion.Version,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:     ciliumCRD.Spec.Names.Plural,
				Singular:   ciliumCRD.Spec.Names.Singular,
				ShortNames: ciliumCRD.Spec.Names.ShortNames,
				Kind:       ciliumCRD.Spec.Names.Kind,
			},
			Subresources: ciliumCRD.Spec.Subresources,
			Scope:        ciliumCRD.Spec.Scope,
		},
	}

	return createUpdateCRD(clientset, "v2.CiliumIdentity", res)
}

// createUpdateCRD ensures the CRD object is installed into the k8s cluster. It
// will create or update the CRD and it's validation when needed
func createUpdateCRD(clientset apiextensionsclient.Interface,
	crdName string,
	crd *apiextensionsv1beta1.CustomResourceDefinition) error {

	scopedLog := log.WithField("name", crdName)

	clusterCRD, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(
		context.TODO(),
		crd.ObjectMeta.Name,
		metav1.GetOptions{})
	if errors.IsNotFound(err) {
		scopedLog.Info("Creating CRD (CustomResourceDefinition)...")

		clusterCRD, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(
			context.TODO(),
			crd,
			metav1.CreateOptions{})
		// This occurs when multiple agents race to create the CRD. Since another has
		// created it, it will also update it, hence the non-error return.
		if errors.IsAlreadyExists(err) {
			return nil
		}
	}
	if err != nil {
		return err
	}

	scopedLog.Debug("Checking if CRD (CustomResourceDefinition) needs update...")

	if crd.Spec.Validation != nil &&
		clusterCRD.Labels[CustomResourceDefinitionSchemaVersionKey] != "" &&
		needsUpdate(clusterCRD) {

		scopedLog.Info("Updating CRD (CustomResourceDefinition)...")

		// Update the CRD with the validation schema.
		err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
			clusterCRD, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(
				context.TODO(),
				crd.ObjectMeta.Name,
				metav1.GetOptions{})
			if err != nil {
				return false, err
			}

			// This seems too permissive but we only get here if the version is
			// different per needsUpdate above. If so, we want to update on any
			// validation change including adding or removing validation.
			if needsUpdate(clusterCRD) {
				scopedLog.Debug("CRD validation is different, updating it...")

				clusterCRD.ObjectMeta.Labels = crd.ObjectMeta.Labels
				clusterCRD.Spec = crd.Spec

				_, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Update(
					context.TODO(),
					clusterCRD,
					metav1.UpdateOptions{})
				if err == nil {
					return true, nil
				}

				scopedLog.WithError(err).Debug("Unable to update CRD validation")
				return false, err
			}

			return true, nil
		})
		if err != nil {
			scopedLog.WithError(err).Error("Unable to update CRD")
			return err
		}
	}

	// wait for the CRD to be established
	scopedLog.Debug("Waiting for CRD (CustomResourceDefinition) to be available...")
	err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		for _, cond := range clusterCRD.Status.Conditions {
			switch cond.Type {
			case apiextensionsv1beta1.Established:
				if cond.Status == apiextensionsv1beta1.ConditionTrue {
					return true, err
				}
			case apiextensionsv1beta1.NamesAccepted:
				if cond.Status == apiextensionsv1beta1.ConditionFalse {
					scopedLog.WithError(goerrors.New(cond.Reason)).Error("Name conflict for CRD")
					return false, err
				}
			}
		}
		clusterCRD, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(
			context.TODO(),
			crd.ObjectMeta.Name,
			metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return false, err
	})
	if err != nil {
		deleteErr := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(
			context.TODO(),
			crd.ObjectMeta.Name,
			metav1.DeleteOptions{})
		if deleteErr != nil {
			return fmt.Errorf("unable to delete k8s %s CRD %s. Deleting CRD due: %s",
				crdName, deleteErr, err)
		}
		return err
	}

	scopedLog.Info("CRD (CustomResourceDefinition) is installed and up-to-date")
	return nil
}

func needsUpdate(clusterCRD *apiextensionsv1beta1.CustomResourceDefinition) bool {
	if clusterCRD.Spec.Validation == nil {
		// no validation detected
		return true
	}

	v, ok := clusterCRD.Labels[CustomResourceDefinitionSchemaVersionKey]
	if !ok {
		// no schema version detected
		return true
	}

	clusterVersion, err := versioncheck.Version(v)
	if err != nil || clusterVersion.LT(comparableCRDSchemaVersion) {
		// version in cluster is either unparsable or smaller than current version
		return true
	}

	return false
}

var (
	// cepCRV is a minimal validation for CEP objects. Since only the agent is
	// creating them, it is better to be permissive and have some data, if buggy,
	// than to have no data in k8s.
	cepCRV = apiextensionsv1beta1.CustomResourceValidation{
		OpenAPIV3Schema: &apiextensionsv1beta1.JSONSchemaProps{},
	}

	cnpCRV = apiextensionsv1beta1.CustomResourceValidation{
		OpenAPIV3Schema: &apiextensionsv1beta1.JSONSchemaProps{},
	}
)
