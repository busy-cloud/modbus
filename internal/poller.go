package internal

import (
	"github.com/busy-cloud/boat/log"
)

type Pollers struct {
	Crontab  string    `json:"crontab,omitempty"`  //定时读取
	Interval uint      `json:"interval,omitempty"` //轮询间隔
	Timeout  uint      `json:"timeout,omitempty"`  //读取超时
	Pollers  []*Poller `json:"pollers"`            //轮询表
}

type Poller struct {
	Code    uint8  `json:"code"`    //功能码 1 2 3 4
	Address uint16 `json:"address"` //地址
	Length  uint16 `json:"length"`  //长度
}

func (p *Poller) Parse(mappers *Mappers, buf []byte, values map[string]any) error {
	switch p.Code {
	case 1:
		for _, m := range mappers.Coils {
			if p.Address <= m.Address && m.Address < p.Address+p.Length {
				ret, err := m.Parse(p.Address, buf)
				if err != nil {
					log.Error(err)
					continue
				}
				values[m.Name] = ret
			}
		}
	case 2:
		for _, m := range mappers.DiscreteInputs {
			if p.Address <= m.Address && m.Address < p.Address+p.Length {
				ret, err := m.Parse(p.Address, buf)
				if err != nil {
					log.Error(err)
					continue
				}
				values[m.Name] = ret
			}
		}
	case 3:
		for _, m := range mappers.HoldingRegisters {
			if p.Address <= m.Address && m.Address < p.Address+p.Length {
				ret, err := m.Parse(p.Address, buf)
				if err != nil {
					log.Error(err)
					continue
				}
				//03 指令 的 位类型
				if rets, ok := ret.(map[string]bool); ok {
					for k, v := range rets {
						values[k] = v
					}
				} else {
					values[m.Name] = ret
				}
			}
		}
	case 4:
		for _, m := range mappers.HoldingRegisters {
			if p.Address <= m.Address && m.Address < p.Address+p.Length {
				ret, err := m.Parse(p.Address, buf)
				if err != nil {
					log.Error(err)
					continue
				}
				//04 指令 的 位类型
				if rets, ok := ret.(map[string]bool); ok {
					for k, v := range rets {
						values[k] = v
					}
				} else {
					values[m.Name] = ret
				}
			}
		}
	}

	return nil
}
