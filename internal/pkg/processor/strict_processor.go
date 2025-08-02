package processor

import (
	validationErrors "k8s.io/kube-openapi/pkg/validation/errors"
	"k8s.io/kube-openapi/pkg/validation/strfmt"
	"k8s.io/kube-openapi/pkg/validation/validate"
	"sigs.k8s.io/kustomize/kyaml/errors"
	fw "sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	k8syaml "sigs.k8s.io/yaml"
)

// SimpleProcessor is a wrapper around kustomize's SimpleProcessor that allows to set a strict mode
// for configuration unmarshalling.
type SimpleProcessor struct {
	fw.SimpleProcessor
	// Strict indicates whether the configuration unmarshalling should be strict.
	// If set to true, it will return an error if the configuration contains unknown fields.
	Strict bool
}

// NewSimpleProcessor creates a new SimpleProcessor with the given configuration and filter.
func NewSimpleProcessor(config interface{}, filter kio.Filter, strict bool) SimpleProcessor {
	return SimpleProcessor{
		SimpleProcessor: fw.SimpleProcessor{
			Config: config,
			Filter: filter,
		},
		Strict: strict,
	}
}

// Process makes SimpleProcessor implement the ResourceListProcessor interface.
// It loads the ResourceList.functionConfig into the provided Config type, applying
// defaulting and validation if supported by Config. It then executes the processor's filter.
func (p SimpleProcessor) Process(rl *fw.ResourceList) error {
	if err := LoadFunctionConfig(rl.FunctionConfig, p.Config, p.Strict); err != nil {
		return errors.WrapPrefixf(err, "loading function config")
	}
	return errors.WrapPrefixf(rl.Filter(p.Filter), "processing filter")
}

// LoadFunctionConfig reads a configuration resource from YAML into the provided data structure
// and then prepares it for use by running defaulting and validation on it, if supported.
// ResourceListProcessors should use this function to load ResourceList.functionConfig.
func LoadFunctionConfig(src *yaml.RNode, api interface{}, strict bool) error {
	if api == nil {
		return nil
	}
	// Run this before unmarshalling to avoid nasty unmarshal failure error messages
	var schemaValidationError error
	if s, ok := api.(fw.ValidationSchemaProvider); ok {
		schema, err := s.Schema()
		if err != nil {
			return errors.WrapPrefixf(err, "loading provided schema")
		}
		schemaValidationError = errors.Wrap(validate.AgainstSchema(schema, src, strfmt.Default))
		// don't return it yet--try to make it to custom validation stage to combine errors
	}

	var err error
	if strict {
		err = k8syaml.UnmarshalStrict([]byte(src.MustString()), api)
	} else {
		err = k8syaml.Unmarshal([]byte(src.MustString()), api)
	}

	if err != nil {
		if schemaValidationError != nil {
			// if we got a validation error, report it instead as it is likely a nicer version of the same message
			return schemaValidationError
		}
		return errors.Wrap(err)
	}

	if d, ok := api.(fw.Defaulter); ok {
		if err := d.Default(); err != nil {
			return errors.Wrap(err)
		}
	}

	if v, ok := api.(fw.Validator); ok {
		return combineErrors(schemaValidationError, v.Validate())
	}
	return schemaValidationError
}

// combineErrors produces a CompositeValidationError for the given schemaErr and givenErr.
// If either is already a CompsiteError, its constituent errors become part of the new
// composite error. If both given errors are nil, this function returns nil.
func combineErrors(schemaErr, customErr error) error {
	combined := validationErrors.CompositeValidationError()
	if compositeSchemaErr, ok := schemaErr.(*validationErrors.CompositeError); ok {
		combined.Errors = append(combined.Errors, compositeSchemaErr.Errors...)
	} else if schemaErr != nil {
		combined.Errors = append(combined.Errors, schemaErr)
	}
	if compositeCustomErr, ok := customErr.(*validationErrors.CompositeError); ok {
		combined.Errors = append(combined.Errors, compositeCustomErr.Errors...)
	} else if customErr != nil {
		combined.Errors = append(combined.Errors, customErr)
	}
	if len(combined.Errors) > 0 {
		return combined
	}
	return nil
}
