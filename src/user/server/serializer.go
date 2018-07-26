package server

import (
	"github.com/Tanibox/tania-core/src/user/domain"
	"github.com/Tanibox/tania-core/src/user/storage"
	uuid "github.com/satori/go.uuid"
)

type Organization struct {
	*storage.OrganizationRead
	*UserOrganization
}

type UserOrganization struct {
	UID   uuid.UUID `json:"user_id"`
	Email string    `json:"user_email"`
}

func MapToUserRead(user *domain.User) storage.UserRead {
	userRead := storage.UserRead{}
	userRead.UID = user.UID
	userRead.Email = user.Email
	userRead.Role = user.Role
	userRead.Status = user.Status
	userRead.Name = user.Name
	userRead.Gender = user.Gender
	userRead.BirthDate = user.BirthDate
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

func MapOrganizationWithUser(org *domain.Organization, user *domain.User) Organization {
	o := Organization{}
	o.OrganizationRead = &storage.OrganizationRead{}
	o.OrganizationRead.UID = org.UID
	o.OrganizationRead.Name = org.Name
	o.OrganizationRead.Email = org.Email
	o.OrganizationRead.Status = org.Status
	o.OrganizationRead.Type = org.Type
	o.OrganizationRead.TotalMember = org.TotalMember
	o.OrganizationRead.Province = org.Province
	o.OrganizationRead.City = org.City
	o.OrganizationRead.CreatedDate = org.CreatedDate

	o.UserOrganization = &UserOrganization{}
	o.UserOrganization.UID = user.UID
	o.UserOrganization.Email = user.Email

	return o
}
