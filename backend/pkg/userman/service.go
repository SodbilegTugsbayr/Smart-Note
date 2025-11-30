package userman

import (
	"errors"
	"log"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	db       *gorm.DB
}

func NewService(db *gorm.DB, infoLog, errorLog *log.Logger) *Service {
	return &Service{
		infoLog:  infoLog,
		errorLog: errorLog,
		db:       db,
	}
}

func (s *Service) parseFilter(filter *Filter) *gorm.DB {
	query := s.db

	if filter == nil {
		return query
	}

	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}

	if filter.Keyword != "" {
		query = query.Where("first_name ILIKE '%%' || ? || '%%' OR last_name ILIKE '%%' || ? || '%%' OR email ILIKE '%%' || ? || '%%'",
			filter.Keyword, filter.Keyword, filter.Keyword)
	}

	if filter.Role != "" {
		query = query.Where("role=?", filter.Role)
	}

	if filter.Email != "" {
		query = query.Where("email ILIKE ? || '%%'", filter.Email)
	}

	if len(filter.Emails) > 0 {
		query = query.Where("email in ?", filter.Emails)
	}

	return query
}

func (s *Service) Count(filter *Filter) (int, error) {
	query := s.parseFilter(filter)

	var count int64
	if err := query.Model(&User{}).Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}

func (s *Service) GetAll(filter *Filter, page, size int) ([]*User, int, error) {
	query := s.parseFilter(filter)

	var count int64
	if err := query.Model(&User{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if size > 0 {
		query = query.Limit(size)
		if page > 0 {
			query = query.Offset((page - 1) * size)
		}
	}

	var users []*User
	if err := query.Order("role, email").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, int(count), nil
}

func (s *Service) Get(data *User) (*User, error) {
	var user *User

	if err := s.db.Where("self_deleted_at IS NULL").First(&user, data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		} else {
			return nil, err
		}
	}

	return user, nil
}

func (s *Service) GetWithAuthTypes(data *User, authTypes []string) (*User, error) {
	var user *User

	if err := s.db.Where("auth_type IN (?) AND self_deleted_at IS NULL", authTypes).First(&user, data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		} else {
			return nil, err
		}
	}

	return user, nil
}

func (s *Service) GetByID(ID int) (*User, error) {
	var user *User

	if err := s.db.Where("self_deleted_at IS NULL").First(&user, ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		} else {
			return nil, err
		}
	}

	return user, nil
}

func (s *Service) GetRecentlyDeleted(data *User, authTypes []string) (*User, error) {
	var user *User
	yesterday := time.Now().AddDate(0, 0, -1)

	if err := s.db.Where("self_deleted_at IS NOT NULL AND self_deleted_at<?", yesterday).Where("auth_type IN (?)", authTypes).First(&user, data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		} else {
			return nil, err
		}
	}

	return user, nil
}

func (s *Service) Save(data *User) (*User, error) {
	if err := s.db.Save(data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Service) Delete(id int) error {
	return s.db.Delete(new(User), id).Error
}
