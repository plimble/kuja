package reqresp

type AddReq struct {
	A int
	B int
}

type AddResp struct {
	C int
}

type AddService struct{}

func (s *AddService) Add(req *AddReq) (*AddResp, error) {
	resp := &AddResp{}
	resp.C = req.A + req.B
	return resp, nil
}
