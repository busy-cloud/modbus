package internal

type Modbus struct {
	Crontab  string    `json:"crontab,omitempty"`
	Interval uint      `json:"interval,omitempty"`
	Timeout  uint      `json:"timeout,omitempty"`
	Mapper   *Mappers  `json:"mapper,omitempty"`
	Pollers  []*Poller `json:"pollers,omitempty"`
}
