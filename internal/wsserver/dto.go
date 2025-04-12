package wsserver

type wsMsg struct {
	IDChat    string `json:"idChat"`
	IPAddress string `json:"ip_address"`
	IDSender  string `json:"idSender"`
	Message   string `json:"message"`
	Time      string `json:"time"`
}
