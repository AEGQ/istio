// Copyright Istio Authors
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

package validate

import (
	"github.com/ghodss/yaml"
	"istio.io/istio/operator/pkg/apis/istio/v1alpha1"
	"istio.io/istio/operator/pkg/util"
)

//"gateways.istio\\-egressgateway.autoscaleEnabled": validateDefaultPDB,

var (
	// DefaultValuesValidations maps a data path to a validation function.
	DefaultValuesValidations = map[string]ValidatorFunc{
		"global.proxy.includeIPRanges":              validateIPRangesOrStar,
		"global.proxy.excludeIPRanges":              validateIPRangesOrStar,
		"global.proxy.includeInboundPorts":          validateStringList(validatePortNumberString),
		"global.proxy.excludeInboundPorts":          validateStringList(validatePortNumberString),
		"meshConfig":                                validateMeshConfig,
		"global.defaultPodDisruptionBudget.enabled": checkEnabled,
		"pilot.autoscaleEnabled":                    checkEnabled,
		"pilot.autoscaleMin":                        checkAutoscaleMin,
		// TODO: what is the right path for ingress and egress gateway?
		// "gateways.[name:istio-egressgateway].autoscaleEnabled": checkEnabled,
	}
)

// CheckValues validates the values in the given tree, which follows the Istio values.yaml schema.
func CheckValues(root interface{}) util.Errors {
	vs, err := yaml.Marshal(root)
	if err != nil {
		return util.Errors{err}
	}
	val := &v1alpha1.Values{}
	if err := util.UnmarshalWithJSONPB(string(vs), val, false); err != nil {
		return util.Errors{err}
	}

	if err := ValuesValidate(DefaultValuesValidations, root, nil); err !=  nil {
		return util.Errors{err}
	}

	// Validate HA mode
	if err := validateDefaultPDB(checkEnabledMap, checkAutoscaleMinMap); err != nil {
		return util.Errors{err}
	}
	return nil
}

// ValuesValidate validates the values of the tree using the supplied Func
func ValuesValidate(validations map[string]ValidatorFunc, node interface{}, path util.Path) (errs util.Errors) {
	pstr := path.String()
	scope.Debugf("ValuesValidate %s", pstr)
	vf := validations[pstr]
	if vf != nil {
		errs = util.AppendErrs(errs, vf(path, node))
	}

	nn, ok := node.(map[string]interface{})
	if !ok {
		// Leaf, nothing more to recurse.
		return errs
	}
	for k, v := range nn {
		errs = util.AppendErrs(errs, ValuesValidate(validations, v, append(path, k)))
	}
	return errs
}
