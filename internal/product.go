package internal

import (
	"errors"
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/lib"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/iot/types"
	"xorm.io/xorm/schemas"
)

type ProductMappers struct {
	types.ProductConfig `xorm:"extends"`
	Content             *Mappers `json:"content,omitempty"`
}

func (p *ProductMappers) TableName() string {
	return "product_config"
}

type ProductPollers struct {
	types.ProductConfig `xorm:"extends"`
	Content             *Pollers `json:"content,omitempty"`
}

func (p *ProductPollers) TableName() string {
	return "product_config"
}

type Product struct {
	types.Product

	mappers *Mappers
	pollers *Pollers
}

func (p *Product) Load() error {

	var mapper ProductMappers
	has, err := db.Engine().ID(schemas.PK{p.Id, "config"}).Get(&mapper)
	if err != nil {
		return err
	}
	if !has {
		return errors.New("缺少映射")
	}
	p.mappers = mapper.Content

	var poller ProductPollers
	has, err = db.Engine().ID(schemas.PK{p.Id, "config"}).Get(&poller)
	if err != nil {
		return err
	}
	if !has {
		log.Info(p.Id, "缺少轮询")
		//return errors.New("缺少轮询")
	}
	p.pollers = poller.Content

	return nil
}

var products lib.Map[Product]

func LoadProduct(id string) (*Product, error) {
	var product Product
	has, err := db.Engine().ID(id).Get(&product.Product)
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
