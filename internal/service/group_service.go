package service

import (
	"context"
	"errors"
	"time"

	"github.com/chat-backend/internal/models"
	"github.com/chat-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/sourcegraph/conc"
)

const (
	RoleAdmin  = "admin"
	RoleMember = "member"
)

type GroupService struct {
	groupRepo repository.GroupRepository
	userRepo  repository.UserRepository
}

func NewGroupService(groupRepo repository.GroupRepository, userRepo repository.UserRepository) *GroupService {
	return &GroupService{
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

type CreateGroupInput struct {
	Name        string      `json:"name" binding:"required"`
	Description string      `json:"description"`
	CreatorID   uuid.UUID   `json:"creator_id" binding:"required"`
	Members     []uuid.UUID `json:"members"`
}

func (s *GroupService) CreateGroup(ctx context.Context, input CreateGroupInput) (*models.Group, error) {
	// Create group
	group := &models.Group{
		ID:          uuid.New(),
		Name:        input.Name,
		Description: input.Description,
		CreatorID:   input.CreatorID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	// Add creator as admin
	if err := s.groupRepo.AddMember(ctx, group.ID, input.CreatorID, RoleAdmin); err != nil {
		return nil, err
	}

	// Add members concurrently
	if len(input.Members) > 0 {
		var wg conc.WaitGroup
		errors := make(chan error, len(input.Members))

		for _, memberID := range input.Members {
			memberID := memberID // Create new variable for goroutine
			wg.Go(func() {
				if err := s.groupRepo.AddMember(ctx, group.ID, memberID, RoleMember); err != nil {
					errors <- err
				}
			})
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			if err != nil {
				return group, err
			}
		}
	}

	return group, nil
}

func (s *GroupService) GetGroup(ctx context.Context, id uuid.UUID) (*models.Group, error) {
	return s.groupRepo.GetByID(ctx, id)
}

func (s *GroupService) UpdateGroup(ctx context.Context, group *models.Group) error {
	group.UpdatedAt = time.Now()
	return s.groupRepo.Update(ctx, group)
}

func (s *GroupService) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	return s.groupRepo.Delete(ctx, id)
}

func (s *GroupService) AddMember(ctx context.Context, groupID, userID uuid.UUID, role string) error {
	// Validate user exists
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		return errors.New("user not found")
	}

	// Check if user is already a member
	members, err := s.groupRepo.GetMembers(ctx, groupID)
	if err != nil {
		return err
	}

	for _, member := range members {
		if member.UserID == userID {
			return errors.New("user is already a member of this group")
		}
	}

	return s.groupRepo.AddMember(ctx, groupID, userID, role)
}

func (s *GroupService) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	// Check if user is the last admin
	members, err := s.groupRepo.GetMembers(ctx, groupID)
	if err != nil {
		return err
	}

	adminCount := 0
	isMemberAdmin := false
	for _, member := range members {
		if member.Role == RoleAdmin {
			adminCount++
		}
		if member.UserID == userID && member.Role == RoleAdmin {
			isMemberAdmin = true
		}
	}

	if adminCount == 1 && isMemberAdmin {
		return errors.New("cannot remove the last admin from the group")
	}

	return s.groupRepo.RemoveMember(ctx, groupID, userID)
}

func (s *GroupService) GetMembers(ctx context.Context, groupID uuid.UUID) ([]models.GroupMember, error) {
	return s.groupRepo.GetMembers(ctx, groupID)
}

func (s *GroupService) GetUserGroups(ctx context.Context, userID uuid.UUID) ([]models.Group, error) {
	return s.groupRepo.GetUserGroups(ctx, userID)
}

func (s *GroupService) UpdateMemberRole(ctx context.Context, groupID, userID uuid.UUID, newRole string) error {
	// Validate role
	if newRole != RoleAdmin && newRole != RoleMember {
		return errors.New("invalid role")
	}

	// Check if user is a member
	members, err := s.groupRepo.GetMembers(ctx, groupID)
	if err != nil {
		return err
	}

	isMember := false
	adminCount := 0
	for _, member := range members {
		if member.Role == RoleAdmin {
			adminCount++
		}
		if member.UserID == userID {
			isMember = true
			// If demoting an admin, check if they're the last one
			if member.Role == RoleAdmin && newRole == RoleMember && adminCount == 1 {
				return errors.New("cannot demote the last admin")
			}
		}
	}

	if !isMember {
		return errors.New("user is not a member of this group")
	}

	return s.groupRepo.UpdateMemberRole(ctx, groupID, userID, newRole)
}
