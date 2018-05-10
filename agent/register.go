package agent

import ()

// RegisterAllAPI ...
func (s *Service) RegisterAllAPI() {
	s.h.RegisterWildCardAPI(s.PassThru)
	s.h.RegisterAPI("publish", s.Publish)
	s.h.RegisterAPI("getstreamitem", s.GetStreamItem)
	s.h.RegisterAPI("liststreamkeyitems",s.ListStreamKeyItems)
}
