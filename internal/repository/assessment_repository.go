package repository

import (
	"errors"
	"inkwell-backend-V2.0/internal/db"
	"inkwell-backend-V2.0/internal/model"
)

type AssessmentRepository interface {
	CreateAssessment(assessment *model.Assessment) error
	GetAssessments() ([]model.Assessment, error)
	GetAssessmentBySessionID(sessionID string) (*model.Assessment, error)
	SaveAnswer(answer *model.Answer) error
	GetRandomQuestions(topic string, limit int) ([]model.Question, error)
}

type assessmentRepository struct{}

func NewAssessmentRepository() AssessmentRepository {
	return &assessmentRepository{}
}

func (r *assessmentRepository) CreateAssessment(assessment *model.Assessment) error {
	return db.GetDB().Create(assessment).Error
}

func (r *assessmentRepository) GetAssessments() ([]model.Assessment, error) {
	var assessments []model.Assessment
	err := db.GetDB().Preload("Questions").Find(&assessments).Error
	return assessments, err
}

func (r *assessmentRepository) GetAssessmentBySessionID(sessionID string) (*model.Assessment, error) {
	var assessment model.Assessment
	err := db.GetDB().Preload("Questions").Where("session_id = ?", sessionID).First(&assessment).Error
	if err != nil {
		return nil, errors.New("assessment not found")
	}
	return &assessment, nil
}

func (r *assessmentRepository) SaveAnswer(answer *model.Answer) error {
	return db.GetDB().Create(answer).Error
}

func (r *assessmentRepository) GetRandomQuestions(topic string, limit int) ([]model.Question, error) {
	var questions []model.Question
	err := db.GetDB().Raw(`SELECT * FROM questions WHERE topic = ? ORDER BY RANDOM() LIMIT ?`, topic, limit).Scan(&questions).Error
	if err != nil {
		return nil, err
	}
	return questions, nil
}
