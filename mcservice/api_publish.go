package mcservice

import (
	"log"
)

func (s *MCService) publish(req *JSONRequest) (*JSONResponse, error) {
	if len(req.Params) < 3 {
		return nil, errNumParameter
	}
	_, ok := req.Params[0].(string)
	if ok != true {
		return nil, errParameter
	}
	_, ok = req.Params[1].(string)
	if ok != true {
		return nil, errParameter
	}
	data, ok := req.Params[2].(string)
	if ok != true {
		return nil, errParameter
	}
	hexstr, err := s.boxer.Box([]byte(data), s.cfg.NativeEntity)
	if err != nil {
		log.Printf("could not encode: %s", err)
		return nil, errInternal
	}
	req.Params[2] = hexstr
	return s.platformAPI(req)
}
