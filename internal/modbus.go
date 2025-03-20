package internal

type ModbusConfig struct {
	Crontab  string    `json:"crontab,omitempty"`
	Interval uint      `json:"interval,omitempty"`
	Timeout  uint      `json:"timeout,omitempty"`
	Mapper   *Mapper   `json:"mapper,omitempty"`
	Pollers  []*Poller `json:"pollers,omitempty"`
}
