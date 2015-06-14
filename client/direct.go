package client

type Direct struct {
	url string
}

func (d *Direct) GetEndpoint(service, method string) string {
	return concat(d.url, "/", service, "/", method)
}
