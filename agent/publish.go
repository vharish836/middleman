package agent

import "log"

// Publish ...
func (s *Service) Publish(req *JSONRequest) (*JSONResponse, error) {
	if len(req.Params) < 3 {
		return &JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Need exactly three arguments, refer to help",
		}}, nil
	}
	_, ok := req.Params[0].(string)
	if ok != true {
		return &JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Invalid stream-identifier",
		}}, nil
	}
	entty, ok := req.Params[1].(string)
	if ok != true {
		return &JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Invalid key",
		}}, nil
	}
	data, ok := req.Params[2].(string)
	if ok != true {
		return &JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Invalid data",
		}}, nil
	}
	var key []byte
	var hexstr string
	var err error
	k, ok := s.entityKeys.Load(entty)
	if ok != true {
		return &JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Invalid key",
		}}, nil
	}
	key = []byte(k.(string))
	if s.cfg.AESMode == GCMMode {
		hexstr, err = aesEncryptGCM([]byte(data), key)
	} else {
		hexstr, err = aesEncryptCBC([]byte(data), key)
	}

	if err != nil {
		log.Printf("could not encode: %s", err)
		return &JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Internal Server Error",
		}}, nil
	}
	req.Params[2] = hexstr
	return s.PlatformAPI(req)
}
