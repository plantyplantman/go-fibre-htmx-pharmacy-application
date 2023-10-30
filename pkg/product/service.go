package product

import (
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Service interface {
	InsertProduct(product *entities.Product) (*entities.Product, error)
	FetchProduct(p *presenter.Product) error
	FetchProducts(scopes ...func(*gorm.DB) *gorm.DB) ([]*presenter.Product, error)
	UpdateProduct(product *entities.Product) (*entities.Product, error)
	CreateProduct(product *entities.Product) (*entities.Product, error)
	RemoveProduct(ID ...uint) error
}

type service struct {
	repository Repository
}

func NewService(r Repository) Service {
	return &service{
		repository: r,
	}
}

func (s *service) InsertProduct(product *entities.Product) (*entities.Product, error) {
	return s.repository.CreateProduct(product)
}

func (s *service) FetchProduct(p *presenter.Product) error {
	var (
		ep  *entities.Product
		err error
	)

	if ep, err = s.repository.GetProduct(p.ToEntity()); err != nil {
		return err
	}
	p.FromEntity(ep)
	return nil
}

func (s *service) FetchProducts(scopes ...func(*gorm.DB) *gorm.DB) ([]*presenter.Product, error) {
	var (
		eps []entities.Product
		err error
	)
	if eps, err = s.repository.GetProducts(scopes...); err != nil {
		return nil, err
	}

	var ps = make([]*presenter.Product, 0)
	for _, ep := range eps {
		p := presenter.Product{}
		p.FromEntity(&ep)
		ps = append(ps, &p)
	}
	return ps, nil
}

func (s *service) CreateProduct(product *entities.Product) (*entities.Product, error) {
	return s.repository.CreateProduct(product)
}

func (s *service) UpdateProduct(product *entities.Product) (*entities.Product, error) {
	return s.repository.UpdateProduct(product)
}

func (s *service) RemoveProduct(IDs ...uint) error {
	return s.repository.DeleteProduct(IDs...)
}

func WithSkus(skus ...string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sku IN ?", skus)
	}
}

func WithStockedOnWeb(ctx *fiber.Ctx) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		q := ctx.Query("includeNotStockedWeb")
		if q == "off" || q == "" {
			return db.Where("on_web", 1)
		}
		return db
	}
}

func WithStockInformation() func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(clause.Associations)
	}
}

func WithDeleted(ctx *fiber.Ctx) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		q := ctx.Query("includeDeleted")
		log.Println(q)
		if q == "on" {
			return db
		}
		return db.Where("name ~ '^[A-Za-z0-9]'")
	}
}

func WithPromoInformation() func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("promo_informations")
	}
}

func WithQuery(c *fiber.Ctx) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		q := strings.ToUpper(c.Query("query"))
		if q == "" {
			return db
		}

		return db.Where("sku LIKE ?", "%"+q+"%").Or("name LIKE ?", "%"+q+"%")
	}
}

func WithPaginate(c *fiber.Ctx) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		var (
			page  int
			limit int
			err   error
		)
		page = c.QueryInt("page", 1)

		if limit, err = strconv.Atoi(c.Query("limit")); err != nil || limit <= 0 {
			limit = 50
		}

		if limit <= 0 {
			limit = 50
		}

		return db.Limit(limit).Offset((page - 1) * limit)
	}
}

func WithOrderBy(c *fiber.Ctx) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		q := c.Query("orderBy")
		if q == "" {
			q = "name"
		}

		desc, err := strconv.ParseBool(c.Query("desc"))
		if err != nil {
			desc = false
		}

		return db.Order(clause.OrderByColumn{
			Column: clause.Column{Name: q},
			Desc:   desc,
		})
	}
}
