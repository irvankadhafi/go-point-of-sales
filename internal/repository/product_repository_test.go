package repository

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"testing"
)

func TestProductRepository_Create(t *testing.T) {
	initializeTest()
	kit, closer := initializeRepoTestKit(t)
	defer closer()
	mock := kit.dbmock

	ctx := context.TODO()
	repo := &productRepository{
		db:    kit.db,
		cache: kit.cache,
	}

	userID := int64(111)
	product := &model.Product{
		ID:          utils.GenerateID(),
		Name:        "Apple iPhone 14 Pro Max",
		Slug:        "apple-iphone-14-pro-max",
		Price:       19000000,
		Description: "Apple iPhone 14 Pro Max",
		Quantity:    20,
	}

	t.Run("ok", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id",
			"name",
			"slug",
			"price",
			"description",
			"quantity",
		})
		rows.AddRow(product.ID,
			product.Name,
			product.Slug,
			product.Price,
			product.Description,
			product.Quantity)

		mock.ExpectBegin()
		mock.ExpectQuery(`^INSERT INTO "products"`).WillReturnRows(rows)
		mock.ExpectCommit()

		err := repo.Create(context.Background(), userID, product)
		require.NoError(t, err)
	})

	t.Run("failed - create product return err", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`^INSERT INTO "products"`).WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Create(ctx, userID, product)
		require.Error(t, err)
	})
}

func TestProductRepository_FindByID(t *testing.T) {
	initializeTest()
	kit, closer := initializeRepoTestKit(t)
	defer closer()
	mock := kit.dbmock

	ctx := context.TODO()
	repo := &productRepository{
		db:    kit.db,
		cache: kit.cache,
	}

	product := &model.Product{
		ID:          utils.GenerateID(),
		Name:        "Apple iPhone 14 Pro Max",
		Slug:        "apple-iphone-14-pro-max",
		Price:       19000000,
		Description: "Apple iPhone 14 Pro Max",
		Quantity:    20,
	}

	t.Run("ok", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		rows := sqlmock.NewRows([]string{
			"id",
			"name",
			"slug",
			"price",
			"description",
			"quantity",
		})
		rows.AddRow(product.ID,
			product.Name,
			product.Slug,
			product.Price,
			product.Description,
			product.Quantity)

		mock.ExpectQuery("^SELECT .+ FROM \"products\"").WillReturnRows(rows)

		res, err := repo.FindByID(ctx, product.ID)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.True(t, kit.miniredis.Exists(repo.newCacheKeyByID(product.ID)))
	})

	t.Run("failed - return err", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"products\"").WillReturnError(errors.New("db error"))

		res, err := repo.FindByID(ctx, product.ID)
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("failed - not found", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"products\"").
			WillReturnError(gorm.ErrRecordNotFound)

		res, err := repo.FindByID(ctx, product.ID)
		require.NoError(t, err)
		require.Nil(t, res)

		cacheVal, err := kit.miniredis.Get(repo.newCacheKeyByID(product.ID))
		require.NoError(t, err)
		require.Equal(t, `null`, cacheVal)
	})
}

func TestProductRepository_SearchByPage(t *testing.T) {
	initializeTest()
	kit, closer := initializeRepoTestKit(t)
	defer closer()
	mock := kit.dbmock

	ctx := context.TODO()
	repo := &productRepository{
		db:    kit.db,
		cache: kit.cache,
	}

	productIDs := []int64{int64(111), int64(222), int64(333), int64(444)}
	expectedCount := int64(len(productIDs))
	criteria := model.ProductSearchCriteria{
		Page:     1,
		Size:     10,
		SortType: model.ProductSortTypeNameDesc,
	}

	t.Run("success", func(t *testing.T) {
		defer kit.miniredis.FlushAll()
		mock.ExpectQuery(`^SELECT count(.*) FROM "products"`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(len(productIDs)))

		rows := sqlmock.NewRows([]string{"id"})
		for _, id := range productIDs {
			rows.AddRow(id)
		}
		mock.ExpectQuery(`^SELECT .+ FROM "products"`).
			WillReturnRows(rows)

		actualProductIDs, count, err := repo.SearchByPage(ctx, criteria)
		require.NoError(t, err)
		require.Equal(t, len(productIDs), len(actualProductIDs))
		require.Equal(t, expectedCount, count)
	})
}

func TestProductRepository_FindBySlug(t *testing.T) {
	initializeTest()
	kit, closer := initializeRepoTestKit(t)
	defer closer()
	mock := kit.dbmock

	ctx := context.TODO()
	repo := &productRepository{
		db:    kit.db,
		cache: kit.cache,
	}

	product := &model.Product{
		ID:          utils.GenerateID(),
		Name:        "Apple iPhone 14 Pro Max",
		Slug:        "apple-iphone-14-pro-max",
		Price:       19000000,
		Description: "Apple iPhone 14 Pro Max",
		Quantity:    20,
	}

	t.Run("ok", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		rows := sqlmock.NewRows([]string{
			"id",
			"name",
			"slug",
			"price",
			"description",
			"quantity",
		})
		rows.AddRow(product.ID,
			product.Name,
			product.Slug,
			product.Price,
			product.Description,
			product.Quantity)

		productIDRes := sqlmock.NewRows([]string{"id"}).AddRow(product.ID)

		mock.ExpectQuery("^SELECT .+ FROM \"products\"").WillReturnRows(productIDRes)
		mock.ExpectQuery("^SELECT .+ FROM \"products\"").WillReturnRows(rows)

		res, err := repo.FindBySlug(ctx, product.Slug)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.True(t, kit.miniredis.Exists(repo.newCacheKeyBySlug(product.Slug)))
	})

	t.Run("success, from cache", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		_ = kit.miniredis.Set(repo.newCacheKeyByID(product.ID), utils.Dump(product))
		_ = kit.miniredis.Set(repo.newCacheKeyBySlug(product.Slug), utils.Int64ToString(product.ID))

		res, err := repo.FindBySlug(context.TODO(), product.Slug)
		assert.NoError(t, err)
		assert.EqualValues(t, product, res)
	})

	t.Run("not found", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		rows := sqlmock.NewRows([]string{
			"id",
			"name",
			"slug",
			"price",
			"description",
			"quantity",
		})

		mock.ExpectQuery("^SELECT .+ FROM \"products\"").WillReturnRows(rows)

		res, err := repo.FindBySlug(ctx, product.Slug)
		assert.NoError(t, err)
		assert.Nil(t, res)

		kit.miniredis.Exists(repo.newCacheKeyBySlug(product.Slug))
	})

	t.Run("error on db", func(t *testing.T) {
		mock.ExpectQuery("^SELECT .+ FROM \"products\"").WillReturnError(gorm.ErrInvalidValue)

		_, err := repo.FindBySlug(ctx, product.Slug)
		assert.Equal(t, gorm.ErrInvalidValue, err)
	})
}
