package quizman

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

	if filter.NoteID > 0 {
		query = query.Where("note_id = ?", filter.NoteID)
	}

	if filter.Keyword != "" {
		query = query.Where("question ILIKE '%%' || ? || '%%'", filter.Keyword)
	}

	if filter.OrderBy != "" {
		query = query.Order(filter.OrderBy)
	}

	return query
}

func (s *Service) Count(filter *Filter) (int, error) {
	query := s.parseFilter(filter)

	var count int64
	if err := query.Model(&Quiz{}).Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}

func (s *Service) GetAll(filter *Filter, page, size int) ([]*Quiz, int, error) {
	query := s.parseFilter(filter)

	var count int64
	if err := query.Model(&Quiz{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if size > 0 {
		query = query.Limit(size)
		if page > 0 {
			query = query.Offset((page - 1) * size)
		}
	}

	var quizs []*Quiz
	if err := query.Find(&quizs).Error; err != nil {
		return nil, 0, err
	}

	return quizs, int(count), nil
}

func (s *Service) Get(data *Quiz) (*Quiz, error) {
	var quiz *Quiz

	if err := s.db.First(&quiz, data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		} else {
			return nil, err
		}
	}

	return quiz, nil
}

func (s *Service) GetByID(ID int) (*Quiz, error) {
	var quiz *Quiz

	if err := s.db.First(&quiz, ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		} else {
			return nil, err
		}
	}

	return quiz, nil
}

func (s *Service) Save(data *Quiz) (*Quiz, error) {
	if err := s.db.Save(data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Service) SaveResult(data *QuizResult) (*QuizResult, error) {
	if err := s.db.Save(data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Service) Delete(id int) error {
	return s.db.Delete(new(Quiz), id).Error
}
