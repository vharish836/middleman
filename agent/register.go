package agent

//go:generate go run gen.go

// RegisterAllAPI ...
func (s *Service) RegisterAllAPI() {
	s.h.RegisterWildCardAPI(s.PassThru)
	s.RegisterGeneratedAPI()
}
