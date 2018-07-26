package server

import (
	"github.com/Tanibox/tania-core/src/user/domain"
	"github.com/Tanibox/tania-core/src/user/storage"
)

func MapToUserRead(user *domain.User) storage.UserRead {
	userRead := storage.UserRead{}
	userRead.UID = user.UID
	userRead.Email = user.Email
	userRead.Role = user.Role
	userRead.Status = user.Status
	userRead.OrganizationUID = user.OrganizationUID
	userRead.CreatedDate = user.CreatedDate
	userRead.LastUpdated = user.LastUpdated

	return userRead
}

func MapToOrganizationRead(org *domain.Organization) storage.OrganizationRead {
	orgRead := storage.OrganizationRead{}
	orgRead.UID = org.UID
	orgRead.Name = org.Name
	orgRead.Email = org.Email
	orgRead.Status = org.Status
	orgRead.Type = org.Type
	orgRead.TotalMember = org.TotalMember
	orgRead.Province = org.Province
	orgRead.City = org.City
	orgRead.CreatedDate = org.CreatedDate

	return orgRead
}
