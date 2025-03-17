package internal

type Options struct {
	Mapper   *Mapper   `json:"mapper,omitempty"`   //映射表
	Crontab  string    `json:"crontab,omitempty"`  //定时读取
	Interval uint      `json:"interval,omitempty"` //轮询间隔
	Timeout  uint      `json:"timeout,omitempty"`  //读取超时
	Pollers  []*Poller `json:"pollers"`            //轮询表
}
