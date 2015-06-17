package client

type Direct struct {
	url string
}

func (d *Direct) GetAddress(service, method string) (string, error) {
	return d.url + "/" + service + "/" + method, nil
	// return concat(d.url, "/", service, "/", method), nil
}
