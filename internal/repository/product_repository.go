package repository

import (
	"context"
	"fmt"
	"github.com/irvankadhafi/go-point-of-sales/cacher"
	"github.com/irvankadhafi/go-point-of-sales/internal/config"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type productRepository struct {
	db        *gorm.DB
	cache     cacher.CacheManager
	auditRepo model.AuditRepository
}

// NewProductRepository create new repository
func NewProductRepository(
	db *gorm.DB,
	cache cacher.CacheManager,
	auditRepo model.AuditRepository,
) model.ProductRepository {
	return &productRepository{
		db:        db,
		cache:     cache,
		auditRepo: auditRepo,
	}
}

// FindByID find product by id
func (p *productRepository) FindByID(ctx context.Context, id int64) (*model.Product, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":       utils.DumpIncomingContext(ctx),
		"productID": id,
	})

	cacheKey := p.newCacheKeyByID(id)
	if !config.DisableCaching() {
		reply, mu, err := cacher.FindFromCacheByKey[*model.Product](p.cache, cacheKey)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		defer cacher.SafeUnlock(mu)

		if mu == nil {
			return reply, nil
		}
	}

	product := &model.Product{}
	err := p.db.WithContext(ctx).Take(product, "id = ?", id).Error
	switch err {
	case nil:
	case gorm.ErrRecordNotFound:
		cacher.StoreNil(ctx, p.cache, cacheKey)
		return nil, nil
	default:
		logger.Error(err)
		return nil, err
	}

	if err := p.cache.StoreWithoutBlocking(cacher.NewItem(cacheKey, utils.Dump(product))); err != nil {
		logger.Error(err)
	}

	return product, nil
}

// FindBySlug find product with specific slug
func (p *productRepository) FindBySlug(ctx context.Context, slug string) (*model.Product, error) {
	logger := logrus.WithFields(logrus.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"slug":    slug})

	cacheKey := p.newCacheKeyBySlug(slug)
	if !config.DisableCaching() {
		reply, mu, err := cacher.FindFromCacheByKey[int64](p.cache, cacheKey)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		defer cacher.SafeUnlock(mu)

		if mu == nil {
			return p.FindByID(ctx, reply)
		}
	}

	var id int64
	err := p.db.WithContext(ctx).Model(model.Product{}).Select("id").Take(&id, "slug = ?", slug).Error
	switch err {
	case nil:
	case gorm.ErrRecordNotFound:
		cacher.StoreNil(ctx, p.cache, cacheKey)
		return nil, nil
	default:
		logger.Error(err)
		return nil, err
	}

	if err = p.cache.StoreWithoutBlocking(cacher.NewItem(cacheKey, id)); err != nil {
		logger.Error(err)
	}

	return p.FindByID(ctx, id)
}

// Create product
func (p *productRepository) Create(ctx context.Context, userID int64, product *model.Product) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":     utils.DumpIncomingContext(ctx),
		"product": utils.Dump(product),
	})

	product.UpdatedAt = time.Now()

	err := p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(product).Error; err != nil {
			logger.Error(err)
			return err
		}

		if err := p.auditRepo.Audit(ctx, tx, product, &model.Audit{
			UserID:        userID,
			AuditableType: p.name(),
			AuditableID:   product.ID,
			Action:        model.AuditActionCreate,
			CreatedAt:     time.Now(),
		}); err != nil {
			logger.Error(err)
			return err
		}

		return nil
	})
	if err != nil {
		logger.Error(err)
		return err
	}

	if err := p.deleteCaches(product); err != nil {
		logger.Error(err)
	}

	return nil
}

func (p *productRepository) Update(ctx context.Context, userID int64, product *model.Product) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":     utils.DumpIncomingContext(ctx),
		"product": utils.Dump(product),
	})

	err := p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Updates(product).Error; err != nil {
			logger.Error(err)
			return err
		}

		if err := p.auditRepo.Audit(ctx, tx, product, &model.Audit{
			UserID:        userID,
			AuditableType: p.name(),
			AuditableID:   product.ID,
			Action:        model.AuditActionUpdate,
			CreatedAt:     time.Now(),
		}); err != nil {
			logger.Error(err)
			return err
		}

		return nil
	})
	if err != nil {
		logger.Error(err)
		return err
	}

	if err := p.deleteCaches(product); err != nil {
		logger.Error(err)
	}

	return nil
}

// Delete soft delete a product
func (p *productRepository) Delete(ctx context.Context, userID int64, product *model.Product) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":     utils.DumpIncomingContext(ctx),
		"userID":  userID,
		"product": utils.Dump(product),
	})

	err := p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(product).Error; err != nil {
			logger.Error(err)
			return err
		}

		err := p.auditRepo.Audit(ctx, tx, product, &model.Audit{
			UserID:        userID,
			AuditableType: p.name(),
			AuditableID:   product.ID,
			Action:        model.AuditActionDelete,
			CreatedAt:     time.Now(),
		})
		if err != nil {
			logger.Error(err)
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error(err)
		return err
	}

	if err := p.deleteCaches(product); err != nil {
		logger.Error(err)
	}

	return nil
}

// SearchByPage find all product with specific criteria
func (p *productRepository) SearchByPage(ctx context.Context, criteria model.ProductSearchCriteria) (ids []int64, count int64, err error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":      utils.DumpIncomingContext(ctx),
		"criteria": utils.Dump(criteria),
	})

	count, err = p.countAll(ctx, criteria)
	if err != nil {
		logger.Error(err)
		return nil, 0, err
	}

	if count <= 0 {
		return nil, 0, nil
	}

	ids, err = p.findAllIDsByCriteria(ctx, criteria)
	switch err {
	case nil:
	case gorm.ErrRecordNotFound:
		return nil, 0, nil
	default:
		logger.Error(err)
		return nil, 0, err
	}

	return ids, count, nil
}

func (p *productRepository) findAllIDsByCriteria(ctx context.Context, criteria model.ProductSearchCriteria) ([]int64, error) {
	var scopes []func(*gorm.DB) *gorm.DB
	scopes = append(scopes, scopeByPageAndLimit(criteria.Page, criteria.Size))
	if criteria.Query != "" {
		scopes = append(scopes, scopeMatchTSQuery(criteria.Query))
	}

	var ids []int64
	err := p.db.WithContext(ctx).
		Model(model.Product{}).
		Scopes(scopes...).
		Order(orderByProductSortType(criteria.SortType)).
		Pluck("id", &ids).Error
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"ctx":      utils.DumpIncomingContext(ctx),
			"criteria": utils.Dump(criteria),
		}).Error(err)
		return nil, err
	}

	return ids, nil
}

func (p *productRepository) countAll(ctx context.Context, criteria model.ProductSearchCriteria) (int64, error) {
	var scopes []func(*gorm.DB) *gorm.DB
	if criteria.Query != "" {
		scopes = append(scopes, scopeMatchTSQuery(criteria.Query))
	}

	var count int64
	err := p.db.WithContext(ctx).Model(model.Product{}).
		Scopes(scopes...).
		Count(&count).
		Error
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"ctx":      utils.DumpIncomingContext(ctx),
			"criteria": utils.Dump(criteria),
		}).Error(err)
		return 0, err
	}

	return count, nil
}

func (p *productRepository) newCacheKeyByID(id int64) string {
	return fmt.Sprintf("cache:object:product:id:%d", id)
}

func (p *productRepository) newCacheKeyBySlug(slug string) string {
	return fmt.Sprintf("cache:object:product:slug:%s", slug)
}

func (p *productRepository) newProductCacheKeyByCriteria(criteria model.ProductSearchCriteria) string {
	key := fmt.Sprintf("cache:object:productMultiValue:page:%d:size:%d:sortType:%s", criteria.Page, criteria.Size, string(criteria.SortType))

	if criteria.Query != "" {
		return key + ":query:" + criteria.Query
	}

	return key
}

// deleteCaches delete related cache
func (p *productRepository) deleteCaches(product *model.Product) error {
	if product == nil {
		return nil
	}

	return p.cache.DeleteByKeys([]string{
		p.newCacheKeyByID(product.ID),
		p.newCacheKeyBySlug(product.Slug),
	})
}

func (p *productRepository) name() string {
	return "product"
}

func orderByProductSortType(sortType model.ProductSortType) string {
	if orderBy, ok := model.QueryProductSortByMap[sortType]; ok {
		return orderBy
	}

	return model.QueryProductSortByMap[model.ProductSortTypeCreatedAtDesc]
}
