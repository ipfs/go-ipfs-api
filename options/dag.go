package options

// DagPutOption is a single routing option.
type DagPutOption func(opts *DagPutOptions) error

// DagPutOptions is a set of routing options
type DagPutOptions struct {
	Pin      string
	InputEnc string
	Kind     string
	Other    map[interface{}]interface{}
}

// Apply applies the given options to this Options
func (opts *DagPutOptions) Apply(options ...DagPutOption) error {
	for _, o := range options {
		if err := o(opts); err != nil {
			return err
		}
	}
	return nil
}

// ToOption converts this Options to a single Option.
func (opts *DagPutOptions) ToOption() DagPutOption {
	return func(nopts *DagPutOptions) error {
		*nopts = *opts
		if opts.Other != nil {
			nopts.Other = make(map[interface{}]interface{}, len(opts.Other))
			for k, v := range opts.Other {
				nopts.Other[k] = v
			}
		}
		return nil
	}
}

type dagOpts struct{}

var Dag dagOpts

// Pin is an option for Dag.Put which specifies whether to pin the added
// dags. Default is false.
func (dagOpts) Pin(pin string) DagPutOption {
	return func(opts *DagPutOptions) error {
		opts.Pin = pin
		return nil
	}
}

// InputEnc is an option for Dag.Put which specifies the input encoding of the
// data. Default is "json", most formats/codecs support "raw".
func (dagOpts) InputEnc(enc string) DagPutOption {
	return func(opts *DagPutOptions) error {
		opts.InputEnc = enc
		return nil
	}
}

// Kind is an option for Dag.Put which specifies the format that the dag
// will be added as. Default is cbor.
func (dagOpts) Kind(kind string) DagPutOption {
	return func(opts *DagPutOptions) error {
		opts.Kind = kind
		return nil
	}
}
