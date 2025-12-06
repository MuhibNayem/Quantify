package services

import (
	"errors"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
)

type RoleService interface {
	ListRoles() ([]domain.Role, error)
	GetRoleByID(id uint) (*domain.Role, error)
	CreateRole(name, description string) (*domain.Role, error)
	UpdateRole(id uint, name, description string) (*domain.Role, error)
	DeleteRole(id uint) error
	ListPermissions() ([]domain.Permission, error)
	UpdateRolePermissions(roleID uint, permissionIDs []uint) error
}

type roleService struct {
	repo repository.RoleRepository
}

func NewRoleService(repo repository.RoleRepository) RoleService {
	return &roleService{repo: repo}
}

func (s *roleService) ListRoles() ([]domain.Role, error) {
	return s.repo.ListRoles()
}

func (s *roleService) GetRoleByID(id uint) (*domain.Role, error) {
	return s.repo.GetRoleByID(id)
}

func (s *roleService) CreateRole(name, description string) (*domain.Role, error) {
	// Check for existing
	existing, _ := s.repo.GetRoleByName(name)
	if existing != nil && existing.ID != 0 {
		return nil, errors.New("role with this name already exists")
	}

	role := &domain.Role{
		Name:        name,
		Description: description,
		IsSystem:    false,
	}

	if err := s.repo.CreateRole(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleService) UpdateRole(id uint, name, description string) (*domain.Role, error) {
	role, err := s.repo.GetRoleByID(id)
	if err != nil {
		return nil, err
	}
	if role.IsSystem && name != role.Name {
		return nil, errors.New("cannot rename system roles")
	}

	role.Name = name
	role.Description = description

	if err := s.repo.UpdateRole(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleService) DeleteRole(id uint) error {
	role, err := s.repo.GetRoleByID(id)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return errors.New("cannot delete system roles")
	}

	return s.repo.DeleteRole(id)
}

func (s *roleService) ListPermissions() ([]domain.Permission, error) {
	return s.repo.ListPermissions()
}

func (s *roleService) UpdateRolePermissions(roleID uint, permissionIDs []uint) error {
	// Admin role should always have all permissions, but for now we enforce manual assignment.
	// Or we can block stripping admins.
	role, err := s.repo.GetRoleByID(roleID)
	if err != nil {
		return err
	}
	if role.Name == "Admin" {
		// Validating that we don't accidentally remove all permissions from admin?
		// User is "Admin" role, so they can re-add them.
		// Let's allow it but warn? Or just allow.
	}

	return s.repo.AssignPermissions(roleID, permissionIDs)
}
