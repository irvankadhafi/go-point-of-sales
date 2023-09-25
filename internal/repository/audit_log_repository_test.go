package repository

import (
	"context"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type auditLogRepositoryMocks struct {
	Ctrl *gomock.Controller
}

func initializeAuditLogRepository(t *testing.T) (ar model.AuditRepository, am auditLogRepositoryMocks) {
	initializeTest()
	am.Ctrl = gomock.NewController(t)
	ar = NewAuditRepository()
	return
}
func TestAuditLogRepository_Audit(t *testing.T) {
	db, dbmock := initializeCockroachMockConn()
	ar, _ := initializeAuditLogRepository(t)
	user := &model.User{}
	audit := &model.Audit{}
	t.Run("Success", func(t *testing.T) {
		dbmock.ExpectBegin()
		dbmock.ExpectExec("INSERT INTO \"audits\"").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
		dbmock.ExpectCommit()

		err := ar.Audit(context.TODO(), db, user, audit)
		require.NoError(t, err)
	})
	t.Run("Failed Create Error", func(t *testing.T) {
		dbmock.ExpectExec("INSERT INTO \"audits\"").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(gorm.ErrInvalidData)
		err := ar.Audit(context.TODO(), db, user, audit)
		require.Error(t, err)
	})
	t.Run("Failed Marshal Error", func(t *testing.T) {
		channel := make(chan string)
		err := ar.Audit(context.TODO(), db, channel, audit)
		require.Error(t, err)
	})
}
