package chaos

import "github.com/ricjcosme/kube-monkey/deployments"

type ChaosResult struct {
	chaos *Chaos
	err   error
}

func (r *ChaosResult) Deployment() *deployments.Deployment {
	return r.chaos.Deployment()
}

func (r *ChaosResult) Error() error {
	return r.err
}
