package public

var (
	ChMqtt          chan string
	ChCmd           chan string
	IsMountedSdCard bool
	ICheckSdCardCnt int
	AppConfig       AgentConfig
	Config          map[string]string
)
