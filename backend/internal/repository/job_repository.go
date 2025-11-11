package repository

import (
	"inventory/backend/internal/domain"

	"gorm.io/gorm"
)

type JobRepository struct {
	DB *gorm.DB
}

func NewJobRepository(db *gorm.DB) *JobRepository {
	return &JobRepository{DB: db}
}

func (r *JobRepository) CreateJob(job *domain.Job) error {
	return r.DB.Create(job).Error
}

func (r *JobRepository) GetJob(jobID uint) (*domain.Job, error) {
	var job domain.Job
	err := r.DB.First(&job, jobID).Error
	return &job, err
}

func (r *JobRepository) UpdateJob(job *domain.Job) error {
	return r.DB.Save(job).Error
}
