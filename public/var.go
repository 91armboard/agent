package public

import "sync"

var (
	ChMqtt     chan string
	ChActivity chan string
	ChUpload   chan string
	ChCmd      chan string
	ChStatus   chan string
	ChHttp     chan string

	IsMountedSdCard bool
	IsSdCardNotFind bool
	IsUseTmpFolder  bool
	IUploadErrCnt   int
	IUploadMTUnum   int
	IsUploadBigfile bool

	INoSdCardCnt    int
	ICheckSdCardCnt int
	Tmpsize_mutex   sync.Mutex

	IsPubDualDoorMod bool

	Config map[string]string
)
