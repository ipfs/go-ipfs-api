package options

// DagStatSettings is a set of Dag stat options.
type DagStatSettings struct {
	Progress bool
}

// DagStatOption is a single Dag option.
type DagStatOption func(opts *DagStatSettings) error

// DagStatOptions applies the given option to a DagStatSettings instance.
func DagStatOptions(opts ...DagStatOption) (*DagStatSettings, error) {
	options := &DagStatSettings{
		Progress: false,
	}

	for _, opt := range opts {
		err := opt(options)
		if err != nil {
			return nil, err
		}
	}

	return options, nil
}

// Progress is an option for Dag.Stat which returns progressive data while reading through the DAG
func (dagOpts) Progress(progress bool) DagStatOption {
	return func(opts *DagStatSettings) error {
		opts.Progress = progress
		return nil
	}
}
