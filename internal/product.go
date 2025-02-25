package internal

import (
	"errors"
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/lib"
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

var products lib.Map[Product]

func LoadProduct(id string) (*Product, error) {
	var product Product
	has, err := db.Engine.ID(id).Get(&product.Product)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("product %s not found", id)
	}
	err = product.Load()
	if err != nil {
		return nil, err
	}
	products.Store(id, &product)

	return &product, nil
}

func EnsureProduct(id string) (*Product, error) {
	prod := products.Load(id)
	if prod != nil {
		return prod, nil
	}
	return LoadProduct(id)
}
