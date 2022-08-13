package cluster

func NewSupervisor() *Supervisor {
	supervisor := new(Supervisor)
	supervisor.Functions = make(map[string]Cluster)
	supervisor.Configs = make(map[string]Config)
	return supervisor
}

func (s *Supervisor) Register(function string, cluster Cluster, config ...Config) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, found := s.Functions[function]; found {
		return false
	}

	s.Functions[function] = cluster
	if len(config) > 0 {
		s.Configs[function] = config[0]
	}

	return true
}

func (s Supervisor) Function(identifier string) (*Cluster, *Config, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, found := s.Functions[identifier]; !found {
		return nil, nil, false
	}

	cluster := s.Functions[identifier]
	config := s.Configs[identifier]

	return &cluster, &config, true
}
