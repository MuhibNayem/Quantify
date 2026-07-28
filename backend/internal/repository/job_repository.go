package repository

import (
	"inventory/backend/internal/domain"

	"gorm.io/gorm"
)

type jobRepository struct {
	DB *gorm.DB
}

func NewJobRepository(db *gorm.DB) JobRepository {
	return &jobRepository{DB: db}
}

func (r *jobRepository) CreateJob(job *domain.Job) error {
	return r.DB.Create(job).Error
}

func (r *jobRepository) GetJob(jobID uint) (*domain.Job, error) {
	var job domain.Job
	err := r.DB.First(&job, jobID).Error
	return &job, err
}

func (r *jobRepository) UpdateJob(job *domain.Job) error {
	return r.DB.Save(job).Error
}

func (r *jobRepository) ListJobs(limit int) ([]domain.Job, error) {
	var jobs []domain.Job
	err := r.DB.Order("created_at desc").Limit(limit).Find(&jobs).Error
	return jobs, err
}
