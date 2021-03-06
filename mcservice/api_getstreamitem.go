package mcservice

import (
	"log"
)

func (s *MCService) getstreamitem(req *JSONRequest) (*JSONResponse, error) {
	if len(req.Params) < 2 {
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
	rsp, err := s.platformAPI(req)
	if rsp.Result != nil {
		rdata, ok := rsp.Result.(map[string]interface{})
		if ok != true {
			log.Printf("could not type assert to object")
			return nil, errInternal
		}
		data := rdata["data"].(string)
		if err != nil {
			log.Printf("could not decode hex: %s", err)
			return nil, errInternal
		}
		plaintext, err := s.boxer.UnBox(data)
		if err != nil {
			log.Printf("could not decrypt: %s", err)
			return nil, errInternal
		}

		rdata["data"] = plaintext
	}
	return rsp, err
}
