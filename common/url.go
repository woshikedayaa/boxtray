package common

import (
	"net/url"
)

func CombineArgs(values ...url.Values) url.Values {
	if len(values) == 0 {
		return nil
	}
	if len(values) == 1 {
		return values[0]
	}
	ret := url.Values{}
	// shit code
	for _, value := range values {
		if value == nil {
			continue
		}
		for k, v := range value {
			for _, vv := range v {
				ret.Add(k, vv)
			}
		}
	}
	return ret
}
