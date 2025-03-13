package common

func MapSlice[E any, D any, S ~[]E, DS ~[]D](sl S, f func(idx int, source E) D) DS {
	ret := make(DS, len(sl))
	for i, v := range sl {
		ret[i] = f(i, v)
	}
	return ret
}
