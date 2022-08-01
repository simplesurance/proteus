package types

// ParamValues holds the values of the configuration parameters, as read by
// a single provider. The datastruct maps paramset => parameter name => value.
type ParamValues map[string]map[string]string

// Copy returns an independent copy.
func (p ParamValues) Copy() ParamValues {
	ret := make(ParamValues, len(p))
	for setName, set := range p {
		newSet := make(map[string]string, len(set))

		for paramName, value := range set {
			newSet[paramName] = value
		}

		ret[setName] = newSet
	}

	return ret
}

func (p ParamValues) Get(setName, paramName string) *string {
	if set, ok := p[setName]; ok {
		if ret, ok := set[paramName]; ok {
			return &ret
		}
	}

	return nil
}
