package message

type Message struct {
	Header map[string]string
	Body   Body
}

type Body struct {
	ID     string      `josn:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

type Request struct {
	ID     string      `josn:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

type Response struct {
	ID     interface{} `json:"id"`
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
}
