package internal

type Options struct {
	Tcp             bool  `json:"tcp,omitempty"`              //TCP模式，默认RTU
	Timeout         int64 `json:"timeout,omitempty"`          //读取超时
	Polling         bool  `json:"polling,omitempty"`          //开启轮询
	PollingInterval int64 `json:"polling_interval,omitempty"` //轮询间隔(s)
}
