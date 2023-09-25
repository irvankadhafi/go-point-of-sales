package usecase

import (
	"context"
	"errors"
	"github.com/gosimple/slug"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/irvankadhafi/go-point-of-sales/rbac"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	"github.com/sirupsen/logrus"
	"sync"
)

type productUsecase struct {
	productRepo model.ProductRepository
}

// NewProductUsecase instantiate a new product usecase
func NewProductUsecase(productRepo model.ProductRepository) model.ProductUsecase {
	return &productUsecase{
		productRepo: productRepo,
	}
}

// FindByID find product by specific id
func (p *productUsecase) FindByID(ctx context.Context, requester *model.User, id int64) (*model.Product, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":       utils.DumpIncomingContext(ctx),
		"requester": utils.Dump(requester),
		"productID": id,
	})

	if !requester.HasAccess(rbac.ResourceProduct, rbac.ActionViewAny) {
		return nil, ErrPermissionDenied
	}

	product, err := p.findByID(ctx, id)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return product, nil
}

// Create product from input
func (p *productUsecase) Create(ctx context.Context, requester *model.User, input model.CreateProductInput) (*model.Product, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":       utils.DumpIncomingContext(ctx),
		"requester": utils.Dump(requester),
		"input":     utils.Dump(input),
	})

	if !requester.HasAccess(rbac.ResourceProduct, rbac.ActionCreateAny) {
		return nil, ErrPermissionDenied
	}

	if err := input.Validate(); err != nil {
		logger.Error(err)
		return nil, err
	}

	product := &model.Product{
		ID:          utils.GenerateID(),
		Name:        input.Name,
		Slug:        slug.Make(input.Name),
		Price:       input.Price,
		Description: input.Description,
		Quantity:    input.Quantity,
	}

	existingProduct, err := p.productRepo.FindBySlug(ctx, product.Slug)
	if err != nil && !errors.Is(err, ErrNotFound) {
		logger.Error(err)
		return nil, err
	}
	if existingProduct != nil {
		return nil, ErrAlreadyExist
	}

	if err := p.productRepo.Create(ctx, requester.ID, product); err != nil {
		logger.Error(err)
		return nil, err
	}

	return product, nil
}

// Search product with given search criteria
func (p *productUsecase) Search(ctx context.Context, requester *model.User, criteria model.ProductSearchCriteria) (products []*model.Product, count int64, err error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":      utils.DumpIncomingContext(ctx),
		"criteria": utils.Dump(criteria),
	})

	if !requester.HasAccess(rbac.ResourceProduct, rbac.ActionViewAny) {
		err = ErrPermissionDenied
		return
	}

	criteria.SetDefaultValue()
	productIDs, count, err := p.productRepo.SearchByPage(ctx, criteria)
	if err != nil {
		logger.Error(err)
		return nil, 0, err
	}
	if len(productIDs) <= 0 || count <= 0 {
		return nil, 0, err
	}

	products = p.findAllByIDs(ctx, productIDs)
	if len(products) <= 0 {
		err = ErrNotFound
		return
	}

	return
}

func (p *productUsecase) findByID(ctx context.Context, id int64) (*model.Product, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx": utils.DumpIncomingContext(ctx),
		"id":  id,
	})

	product, err := p.productRepo.FindByID(ctx, id)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if product == nil {
		return nil, ErrNotFound
	}

	return product, nil
}

// findAllByIDs find all products with IDs
func (p *productUsecase) findAllByIDs(ctx context.Context, ids []int64) []*model.Product {
	logger := logrus.WithFields(logrus.Fields{
		"ctx": utils.DumpIncomingContext(ctx),
		"ids": ids,
	})

	// WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup
	// Creating channel to receive found products
	c := make(chan *model.Product, len(ids))

	// Iterating through received ids
	for _, id := range ids {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			product, err := p.findByID(ctx, id)
			if err != nil {
				logger.Error(err)
				return
			}
			c <- product
		}(id)
	}
	wg.Wait()
	close(c)

	if len(c) <= 0 {
		return nil
	}

	// put all products in a map with product id as key
	rs := map[int64]*model.Product{}
	for product := range c {
		if product != nil {
			rs[product.ID] = product
		}
	}

	// sort products based on the order of received ids
	var products []*model.Product
	for _, id := range ids {
		if product, ok := rs[id]; ok {
			products = append(products, product)
		}
	}

	return products
}
