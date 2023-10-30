package product

import (
	"errors"

	"github.com/plantyplantman/bcapi/pkg/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	GetProduct(*entities.Product) (*entities.Product, error)
	GetProducts(scopes ...func(*gorm.DB) *gorm.DB) ([]entities.Product, error)
	CreateProduct(*entities.Product) (*entities.Product, error)
	UpdateProduct(*entities.Product) (*entities.Product, error)
	DeleteProduct(ID ...uint) error
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{Collection: db}
}

type repository struct {
	Collection *gorm.DB
}

func (r *repository) CreateProduct(p *entities.Product) (*entities.Product, error) {
	if p == nil {
		return nil, errors.New("product is nil")
	}
	return p, r.Collection.Create(p).Error
}

func (r *repository) GetProduct(qp *entities.Product) (*entities.Product, error) {
	p := &entities.Product{}
	if err := r.Collection.Preload(clause.Associations).First(p, qp).Error; err != nil {
		return nil, err
	}
	return p, nil
}

func (r *repository) GetProducts(scopes ...func(*gorm.DB) *gorm.DB) ([]entities.Product, error) {
	ps := []entities.Product{}
	if err := r.Collection.Scopes(scopes...).Find(&ps).Error; err != nil {
		return nil, err
	}
	return ps, nil
}

func (r *repository) UpdateProduct(p *entities.Product) (*entities.Product, error) {
	if p == nil {
		return nil, errors.New("product is nil")
	}
	return p, r.Collection.Updates(p).Error
}

func (r *repository) DeleteProduct(ID ...uint) error {
	var ps []*entities.Product
	for _, id := range ID {
		ps = append(ps, &entities.Product{Model: gorm.Model{ID: id}})
	}
	return r.Collection.Delete(&ps).Error
}
