package domain

import "fmt"

type UserRole string

const (
	RoleOwner   UserRole = "owner"
	RoleAdmin   UserRole = "admin"
	RoleManager UserRole = "manager"
	RoleViewer  UserRole = "viewer"
)

type UserTenant struct {
	UserID   int64
	TenantID int64
	Role     UserRole
}

type UserTenantDetail struct {
	TenantID   int64
	TenantName string
	Role       UserRole
}

func ParseUserRole(s string) (UserRole, error) {
	switch s {
	case "owner":
		return RoleOwner, nil
	case "admin":
		return RoleAdmin, nil
	case "manager":
		return RoleManager, nil
	case "viewer":
		return RoleViewer, nil
	default:
		return "", fmt.Errorf("неизвестная роль пользователя: %s", s)
	}
}
