package iamgo

import (
	"sync"

	"google.golang.org/genproto/googleapis/iam/v1"
)

type resourceName string

type policyUpdate struct {
	mutex     *sync.Mutex
	resources map[resourceName]*sync.Mutex
	client    iam.IAMPolicyClient
}

func newPolicyUpdate(client iam.IAMPolicyClient) *policyUpdate {
	return &policyUpdate{
		mutex:     &sync.Mutex{},
		resources: make(map[resourceName]*sync.Mutex),
		client:    client,
	}
}

func (s *policyUpdate) lock(resource resourceName) func() {
	s.mutex.Lock()
	mutex, ok := s.resources[resource]
	if !ok {
		mutex = &sync.Mutex{}
		s.resources[resource] = mutex
	}
	s.mutex.Unlock()
	mutex.Lock()
	return mutex.Unlock
}
