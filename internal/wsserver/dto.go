package wsserver

type wsMsg struct {
	IPAddress string `json:"ip_address"`
	Message   string `json:"message"`
	Time      string `json:"time"`
}
