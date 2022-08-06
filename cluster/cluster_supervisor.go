package cluster

func NewSupervisor() *Supervisor {
	supervisor := new(Supervisor)
	supervisor.functions = make(map[string]Cluster)
	supervisor.configs = make(map[string]Config)
	return supervisor
}

func (s *Supervisor) Register(function string, cluster Cluster, config ...Config) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, found := s.functions[function]; found {
		return false
	}

	s.functions[function] = cluster
	if len(config) > 0 {
		s.configs[function] = config[0]
	}

	return true
}

func (s Supervisor) Function(identifier string) (*Cluster, *Config, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, found := s.functions[identifier]; !found {
		return nil, nil, false
	}

	cluster := s.functions[identifier]
	config := s.configs[identifier]

	return &cluster, &config, true
}
