package internal

type Options struct {
	Timeout         int64 `json:"timeout"`
	Polling         bool  `json:"polling,omitempty"`          //开启轮询
	PollingInterval int64 `json:"polling_interval,omitempty"` //轮询间隔(s)
}
