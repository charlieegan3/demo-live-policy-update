package handlers

import "github.com/charlieegan3/demo-live-policy-update/pkg/opa"

type Options struct {
	OPAManager *opa.Manager

	DevMode bool

	EtagScript string
	EtagStyles string
}
