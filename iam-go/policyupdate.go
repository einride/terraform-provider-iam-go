package iam_go

import (
	"google.golang.org/genproto/googleapis/iam/v1"
	"sync"
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
	l, ok := s.resources[resource]
	if !ok {
		l = &sync.Mutex{}
		s.resources[resource] = l
	}
	s.mutex.Unlock()
	l.Lock()
	return l.Unlock
}
