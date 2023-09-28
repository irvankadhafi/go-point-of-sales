package model

import (
	"context"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	"gorm.io/gorm"
	"time"
)

var WesternIndonesiaLayout = "02 January 2006 15:04 WIB"

// Product model
type Product struct {
	ID          int64          `json:"id" gorm:"primary_key"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	Price       int64          `json:"price" sql:"type:decimal(20,0)" gorm:"type:numeric(20,0)"`
	Quantity    int64          `json:"quantity"`
	CreatedAt   time.Time      `json:"created_at" sql:"DEFAULT:'now()':::STRING::TIMESTAMP" gorm:"->;<-:create"` // create & read only
	UpdatedAt   time.Time      `json:"updated_at" sql:"DEFAULT:'now()':::STRING::TIMESTAMP"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at"`
}

// ProductRepository repository
type ProductRepository interface {
	FindByID(ctx context.Context, id int64) (*Product, error)
	SearchByPage(ctx context.Context, criteria ProductSearchCriteria) (ids []int64, count int64, err error)
	FindBySlug(ctx context.Context, slug string) (*Product, error)
	Create(ctx context.Context, userID int64, product *Product) error
	Update(ctx context.Context, userID int64, product *Product) error
	Delete(ctx context.Context, userID int64, product *Product) error
}

// ProductUsecase usecase
type ProductUsecase interface {
	FindByID(ctx context.Context, requester *User, id int64) (*Product, error)
	Search(ctx context.Context, requester *User, criteria ProductSearchCriteria) (products AnyProducts, count int64, err error)
	Create(ctx context.Context, requester *User, input CreateProductInput) (*Product, error)
	UpdateByID(ctx context.Context, requester *User, id int64, input UpdateProductInput) (*Product, error)
	DeleteByID(ctx context.Context, requester *User, id int64) error
}

// CreateProductInput create product input
type CreateProductInput struct {
	Name        string `json:"name" validate:"required,min=3,max=60" example:"Pisang Goreng"`
	Description string `json:"description" validate:"max=80" example:"Pisang goreng gurih"`
	Price       int64  `json:"price" validate:"gte=0" example:"5000"`
	Quantity    int64  `json:"quantity" validate:"gt=0" example:"10"`
}

type UpdateProductInput = CreateProductInput

// Validate validate product input
func (c *CreateProductInput) Validate() error {
	return validate.Struct(c)
}

// ProductSortType sort type for product search
type ProductSortType string

const (
	ProductSortTypeCreatedAtAsc  ProductSortType = "CREATED_AT_ASC"
	ProductSortTypeCreatedAtDesc ProductSortType = "CREATED_AT_DESC"
	ProductSortTypePriceAsc      ProductSortType = "PRICE_ASC"
	ProductSortTypePriceDesc     ProductSortType = "PRICE_DESC"
	ProductSortTypeNameAsc       ProductSortType = "NAME_ASC"
	ProductSortTypeNameDesc      ProductSortType = "NAME_DESC"
)

// QueryProductSortByMap sort type to query string map for database ordering
var QueryProductSortByMap = map[ProductSortType]string{
	ProductSortTypeCreatedAtAsc:  "created_at ASC",
	ProductSortTypeCreatedAtDesc: "created_at DESC",
	ProductSortTypePriceAsc:      "price ASC",
	ProductSortTypePriceDesc:     "price DESC",
	ProductSortTypeNameAsc:       "name ASC",
	ProductSortTypeNameDesc:      "name DESC",
}

// ProductSearchCriteria criteria for searching & sorting product
type ProductSearchCriteria struct {
	Query    string          `json:"query" query:"query"`
	Page     int             `json:"page" query:"page"`
	Size     int             `json:"size" query:"size"`
	SortType ProductSortType `json:"sort_type" query:"sortBy"`
}

// SetDefaultValue will set default value for page and size if zero
func (c *ProductSearchCriteria) SetDefaultValue() {
	if c.Page <= 0 {
		c.Page = 1
	}

	if c.Size <= 0 {
		c.Size = 10
	}

	if c.Size >= 20 {
		c.Size = 20
	}

	if c.SortType == "" {
		c.SortType = ProductSortTypeCreatedAtDesc
	}
}

type ProductResponse struct {
	ID          string `json:"id" example:"1695599921375543118"`
	Name        string `json:"name" example:"Pisang Goreng"`
	Slug        string `json:"slug" example:"pisang-goreng"`
	Description string `json:"description" example:"Pisang goreng gurih"`
	Price       string `json:"price" example:"Rp4.000"`
	Quantity    string `json:"quantity" example:"10"`
	CreatedAt   string `json:"created_at" example:"25 September 2023 13:59 WIB"`
	UpdatedAt   string `json:"updated_at" example:"25 September 2023 13:59 WIB"`
}

func (p Product) ToProductResponse() ProductResponse {
	return ProductResponse{
		ID:          utils.Int64ToString(p.ID),
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		Price:       utils.Int64ToRupiah(p.Price),
		Quantity:    utils.Int64ToString(p.Quantity),
		CreatedAt:   utils.FormatToWesternIndonesianTime(WesternIndonesiaLayout, &p.CreatedAt),
		UpdatedAt:   utils.FormatToWesternIndonesianTime(WesternIndonesiaLayout, &p.UpdatedAt),
	}
}

type AnyProducts []*Product

func (ap AnyProducts) ToListProductResponse() (productResponses []ProductResponse) {
	for _, product := range ap {
		productResponses = append(productResponses, product.ToProductResponse())
	}

	return productResponses
}
