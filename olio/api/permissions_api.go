package api

import (
	"time"

	"github.com/j0ni/service-skeleton-go/olio/common/filter"
	"github.com/j0ni/service-skeleton-go/olio/common/models"
	"github.com/j0ni/service-skeleton-go/olio/dao"
)

const PERMISSION_TYPE_INCLUDES string = "+"
const PERMISSION_TYPE_EXCLUDES string = "_"
const PERMISSION_OPERATION_CREATE string = "C"
const PERMISSION_OPERATION_READ string = "R"
const PERMISSION_OPERATION_UPDATE string = "U"
const PERMISSION_OPERATION_ACCEPT string = "A"
const PERMISSION_OPERATION_DELETE string = "D"
const PERMISSION_OPERATION_ALL string = "*"

const PERMISSION_OBJECT_TYPE_ALL string = "*"
const PERMISSION_OBJECT_TYPE_PASSWORD string = "PASSWORD"
const PERMISSION_OBJECT_TYPE_USER string = "USER"

type PermissionsAPI struct {
	dao *dao.PermissionsDAO
}

////////////////////////////////////////////////////////////////////////////////////////
// public
////////////////////////////////////////////////////////////////////////////////////////

//
// Constructor
//
func NewPermissionsAPI(dao *dao.PermissionsDAO) *PermissionsAPI {
	api := PermissionsAPI{
		dao,
	}
	return &api
}

func (self *PermissionsAPI) Permitted(isSuperAccessUser bool, Permissions []*models.Permission, accessType string, operation string, objectType string, objectID string) bool {
	if isSuperAccessUser {
		// short circuit: if system user, its permitted
		return true
	}

	if Permissions == nil {
		// short circuit: if no permissions were specified, then assume its a non-restricted access context
		return true
	}

	// inspect permissions in the access context
	for _, permission := range Permissions {
		// short circuit: if the ALL operation for ALL object types is present, then assume its a non-restricted access context
		if permission.Operation == PERMISSION_OPERATION_ALL && permission.ObjectType == PERMISSION_OBJECT_TYPE_ALL {
			return true
		}

		var foundAccessType bool
		var foundOperation bool
		var foundObjectType bool
		var foundObjectID bool

		// matched access type (include or exclude)
		if accessType == permission.Type {
			foundAccessType = true
		}

		// matched on operation type (create, update, etc.), or "all"
		if operation == permission.Operation || permission.Operation == PERMISSION_OBJECT_TYPE_ALL {
			foundOperation = true
		}

		// matched on object type (tag, time record, etc.), or "all"
		if objectType == "" || objectType == permission.ObjectType || permission.ObjectType == PERMISSION_OBJECT_TYPE_ALL {
			foundObjectType = true
		}

		// matched on object id or empty (which means "all")
		if objectID == "" || objectID == permission.ObjectID || permission.ObjectID == "" {
			foundObjectID = true
		}

		// if matched on all criteria, the operation is permitted
		if foundAccessType && foundOperation && foundObjectType && foundObjectID {
			return true
		}
	}

	return false
}

func (self *PermissionsAPI) BlacklistToken(token string, expirationDate *time.Time) *Exception {
	accessToken := &models.AccessToken{
		Token:          token,
		ExpirationDate: expirationDate,
	}

	if err := self.dao.Insert(accessToken); err != nil {
		return NewRuntimeException(err.Error())
	}

	return nil
}

func (self *PermissionsAPI) IsTokenBlacklisted(token string) (bool, *Exception) {
	tokenFilter := filters.AccessTokenFilter{Token: token}
	results, err := self.dao.Find(&tokenFilter)
	if err != nil {
		return false, NewRuntimeException(err.Error())
	}

	return len(results) > 0, nil
}
