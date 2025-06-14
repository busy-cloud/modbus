package internal

type ModbusConfig struct {
	Crontab  string    `json:"crontab,omitempty"`
	Interval uint      `json:"interval,omitempty"`
	Timeout  uint      `json:"timeout,omitempty"`
	Mapper   *Mapper   `json:"mapper,omitempty"`
	Pollers  []*Poller `json:"pollers,omitempty"`
	Actions  []*Action `json:"actions,omitempty"`
}

type Options struct {
	Tcp             bool  `json:"tcp,omitempty"`              //TCP模式，默认RTU
	Timeout         int64 `json:"timeout,omitempty"`          //读取超时
	Polling         bool  `json:"polling,omitempty"`          //开启轮询
	PollingInterval int64 `json:"polling_interval,omitempty"` //轮询间隔(s)
}

type Action struct {
	Name      string      `json:"name,omitempty"`
	Label     string      `json:"label,omitempty"`
	Operators []*Operator `json:"operators,omitempty"`
}

type Operator struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"` //表达式
	Delay int64  `json:"delay,omitempty"` //延时
}
