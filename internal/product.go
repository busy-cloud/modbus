package internal

import (
	"errors"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/iot/types"
	"xorm.io/xorm/schemas"
)

type ProductMapper struct {
	types.ProductConfig `xorm:"extends"`
	Content             *Mapper `json:"content,omitempty"`
}

func (p *ProductMapper) TableName() string {
	return "product_config"
}

type ProductPoller struct {
	types.ProductConfig `xorm:"extends"`
	Content             []*Poller `json:"content,omitempty"`
}

func (p *ProductPoller) TableName() string {
	return "product_config"
}

type Product struct {
	types.Product

	mapper *Mapper
	poller []*Poller
}

func (p *Product) Load() error {

	var mapper ProductMapper
	has, err := db.Engine.ID(schemas.PK{p.Id, "config"}).Get(&mapper)
	if err != nil {
		return err
	}
	if !has {
		return errors.New("缺少映射")
	}
	p.mapper = mapper.Content

	var poller ProductPoller
	has, err = db.Engine.ID(schemas.PK{p.Id, "config"}).Get(&poller)
	if err != nil {
		return err
	}
	if !has {
		return errors.New("缺少轮询")
	}
	p.poller = poller.Content

	return nil
}
