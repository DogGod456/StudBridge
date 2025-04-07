package wsserver

type wsMsg struct {
	IPAdress string `json:"ip_address"`
	Message  string `json:"message"`
	Time     string `json:"time"`
}
