package capi

type Version struct {
	Meta    bool   `json:"meta"`
	Premium bool   `json:"premium"`
	Version string `json:"version"`
}
