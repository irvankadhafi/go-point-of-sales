package repository

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestAppClientRepository_FindByClientID(t *testing.T) {
	kit, closer := initializeRepoTestKit(t)
	defer closer()
	mock := kit.dbmock
	initializeTest()

	ctx := context.TODO()
	repo := &appClientRepo{
		db:    kit.db,
		cache: kit.cache,
	}

	clientID := "ic-cms"
	client := model.AppClient{
		ID:       int64(1),
		ClientID: clientID,
	}

	t.Run("ok", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "app_clients" WHERE client_id = $1 LIMIT 1`)).
			WithArgs(clientID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(123))

		client, err := repo.FindByClientID(ctx, clientID)
		require.NoError(t, err)
		require.NotNil(t, client)
		require.True(t, kit.miniredis.Exists(repo.newCacheKeyByClientID(clientID)))
	})

	t.Run("success found cache", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		viper.Set("disable_caching", false)
		err := kit.miniredis.Set(
			repo.newCacheKeyByClientID(clientID),
			utils.Dump(client),
		)
		require.NoError(t, err)

		client, err := repo.FindByClientID(ctx, clientID)
		require.NoError(t, err)
		require.NotNil(t, client)
	})

	t.Run("handle error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "app_clients" WHERE client_id = $1 LIMIT 1`)).
			WithArgs(clientID).WillReturnError(errors.New("db error"))

		client, err := repo.FindByClientID(ctx, clientID)
		require.Error(t, err)
		require.Nil(t, client)
		require.False(t, kit.miniredis.Exists(repo.newCacheKeyByClientID(clientID)))
	})
}

func TestAppClientRepository_FindByID(t *testing.T) {
	kit, closer := initializeRepoTestKit(t)
	defer closer()
	mock := kit.dbmock
	initializeTest()

	ctx := context.TODO()
	repo := &appClientRepo{
		db:    kit.db,
		cache: kit.cache,
	}

	appID := int64(1)
	clientID := "ic-cms"
	client := model.AppClient{
		ID:       appID,
		ClientID: clientID,
	}

	t.Run("ok", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "app_clients" WHERE id = $1 LIMIT 1`)).
			WithArgs(appID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(123))

		client, err := repo.FindByID(ctx, appID)
		require.NoError(t, err)
		require.NotNil(t, client)
		require.True(t, kit.miniredis.Exists(repo.newCacheKeyByID(appID)))
	})

	t.Run("success found cache", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		viper.Set("disable_caching", false)
		err := kit.miniredis.Set(
			repo.newCacheKeyByID(appID),
			utils.Dump(client),
		)
		require.NoError(t, err)

		client, err := repo.FindByID(ctx, appID)
		require.NoError(t, err)
		require.NotNil(t, client)
	})

	t.Run("handle error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "app_clients" WHERE id = $1 LIMIT 1`)).
			WithArgs(appID).WillReturnError(errors.New("db error"))

		client, err := repo.FindByID(ctx, appID)
		require.Error(t, err)
		require.Nil(t, client)
		require.False(t, kit.miniredis.Exists(repo.newCacheKeyByID(appID)))
	})
}
