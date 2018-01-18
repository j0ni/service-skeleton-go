package dao

import (
	"github.com/j0ni/service-skeleton-go/olio/common/filter"
	"github.com/j0ni/service-skeleton-go/olio/common/models"
)

type PermissionsDAO struct {
	baseDAO DAO
}

func NewPermissionsDAO(connectionManager ConnectionProvider, baseDAO DAO) *PermissionsDAO {
	dao := PermissionsDAO{baseDAO}

	return &dao
}

func (self *PermissionsDAO) Find(filter *filters.AccessTokenFilter) ([]models.AccessToken, error) {
	db := self.baseDAO.GetConnectionManager().GetDb()

	if filter.Token != "" {
		db = db.Where("token = ?", filter.Token)
	}

	results := []models.AccessToken{}
	db = db.Find(&results)

	return results, db.Error
}

func (self *PermissionsDAO) Insert(accessToken *models.AccessToken) error {
	return self.baseDAO.Insert(accessToken)
}
