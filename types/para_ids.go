package types

// ParamValues holds the values of the configuration parameters, as read by
// a single provider. The datastruct maps flagset => parameter name => value.
type ParamValues map[string]map[string]string
