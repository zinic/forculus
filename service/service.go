package service

import "sync"

type Service interface {
	Start(waitGroup *sync.WaitGroup)
	Stop()
}

func NewManager() *Manager {
	return &Manager{
		waitGroup: &sync.WaitGroup{},
	}
}

type Manager struct {
	services  []Service
	waitGroup *sync.WaitGroup
}

func (s *Manager) Start(service Service) {
	s.services = append(s.services, service)
	service.Start(s.waitGroup)
}

func (s *Manager) Stop() {
	for _, service := range s.services {
		service.Stop()
	}

	s.waitGroup.Wait()
}
