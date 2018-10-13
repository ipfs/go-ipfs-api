package options

// DagPutOption is a single routing option.
type DagPutOption func(opts *DagPutOptions) error

// DagPutOptions is a set of routing options
type DagPutOptions struct {
	Pin   string
	Other map[interface{}]interface{}
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

// Expired is an option that tells the routing system to return expired records
// when no newer records are known.
var Pin DagPutOption = func(opts *DagPutOptions) error {
	opts.Pin = "true"
	return nil
}
