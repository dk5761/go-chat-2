package postgres

import (
	"context"

	"github.com/chat-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type groupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) *groupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) Create(ctx context.Context, group *models.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

func (r *groupRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Group, error) {
	var group models.Group
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *groupRepository) Update(ctx context.Context, group *models.Group) error {
	return r.db.WithContext(ctx).Save(group).Error
}

func (r *groupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete group members first
		if err := tx.WithContext(ctx).Where("group_id = ?", id).Delete(&models.GroupMember{}).Error; err != nil {
			return err
		}
		// Then delete the group
		return tx.WithContext(ctx).Delete(&models.Group{}, id).Error
	})
}

func (r *groupRepository) AddMember(ctx context.Context, groupID, userID uuid.UUID, role string) error {
	member := &models.GroupMember{
		GroupID: groupID,
		UserID:  userID,
		Role:    role,
	}
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *groupRepository) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&models.GroupMember{}).
		Error
}

func (r *groupRepository) GetMembers(ctx context.Context, groupID uuid.UUID) ([]models.GroupMember, error) {
	var members []models.GroupMember
	err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Find(&members).
		Error
	return members, err
}

func (r *groupRepository) GetUserGroups(ctx context.Context, userID uuid.UUID) ([]models.Group, error) {
	var groups []models.Group
	err := r.db.WithContext(ctx).
		Joins("JOIN group_members ON groups.id = group_members.group_id").
		Where("group_members.user_id = ?", userID).
		Find(&groups).
		Error
	return groups, err
}

func (r *groupRepository) UpdateMemberRole(ctx context.Context, groupID, userID uuid.UUID, role string) error {
	return r.db.WithContext(ctx).
		Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Update("role", role).
		Error
}
