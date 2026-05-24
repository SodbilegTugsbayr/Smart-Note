package courseman

import (
	"errors"
	"log"

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

	if filter.UserID > 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if filter.Keyword != "" {
		query = query.Where("title ILIKE '%%' || ? || '%%' OR description ILIKE '%%' || ? || '%%'",
			filter.Keyword, filter.Keyword)
	}

	if filter.OrderBy != "" {
		query = query.Order(filter.OrderBy)
	}

	return query
}

func (s *Service) Count(filter *Filter) (int, error) {
	query := s.parseFilter(filter)

	var count int64
	if err := query.Model(&Course{}).Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}

func (s *Service) GetAll(filter *Filter, page, size int, preloads ...string) ([]*Course, int, error) {
	query := s.parseFilter(filter)

	var count int64
	if err := query.Model(&Course{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	if size > 0 {
		query = query.Limit(size)
		if page > 0 {
			query = query.Offset((page - 1) * size)
		}
	}

	var courses []*Course
	if err := query.Find(&courses).Error; err != nil {
		return nil, 0, err
	}

	return courses, int(count), nil
}

func (s *Service) Get(data *Course) (*Course, error) {
	var course *Course

	if err := s.db.First(&course, data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		} else {
			return nil, err
		}
	}

	return course, nil
}

func (s *Service) GetByID(ID int, preloads ...string) (*Course, error) {
	query := s.db
	var course *Course

	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	if err := query.First(&course, ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		} else {
			return nil, err
		}
	}

	return course, nil
}

func (s *Service) Save(data *Course) (*Course, error) {
	if err := s.db.Save(data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Service) Delete(id int) error {
	return s.db.Delete(new(Course), id).Error
}
