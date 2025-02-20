package database

import (
	"context"
	"database/sql"
	"errors"
	"spyCatAgency/internal/domain/models"
	"spyCatAgency/internal/infrastructure/apperrors"
	"spyCatAgency/internal/infrastructure/logger"
)

type (
	catRepository struct {
		logger logger.Logger
		*sql.DB
	}
)

func NewCatRepository(customLogger logger.Logger, r *sql.DB) *catRepository {
	return &catRepository{
		logger: customLogger,
		DB:     r,
	}
}

func (r *catRepository) Add(cat models.Cat) (*models.Cat, error) {
	query := "INSERT INTO cats(name, years_of_experience, breed, salary ) VALUES ($1, $2, $3, $4) RETURNING id, name, years_of_experience, breed, salary, created_at;"

	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout)
	defer cancel()

	row := r.QueryRowContext(ctx, query, cat.Name, cat.YearsOfExperience, cat.Breed, cat.Salary)

	var result models.Cat

	err := row.Scan(
		&result.ID,
		&result.Name,
		&result.YearsOfExperience,
		&result.Breed,
		&result.Salary,
		&result.CreatedAt,
	)

	if err != nil {
		r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
		return nil, apperrors.ErrDatabase
	}

	return &result, nil
}

func (r *catRepository) Delete(id uint) error {
	query := "DELETE FROM cats WHERE id = $1;"

	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout)
	defer cancel()

	_, err := r.ExecContext(ctx, query, id)

	return err
}

func (r *catRepository) Update(id uint, salary float64) (*models.Cat, error) {
	query := "UPDATE cats SET salary = $1 WHERE id = $2 RETURNING id, name, years_of_experience, breed, salary, created_at;"

	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout)
	defer cancel()

	row := r.QueryRowContext(ctx, query, salary, id)

	var updatedCat models.Cat

	err := row.Scan(
		&updatedCat.ID,
		&updatedCat.Name,
		&updatedCat.YearsOfExperience,
		&updatedCat.Breed,
		&updatedCat.Salary,
		&updatedCat.CreatedAt,
	)

	if err != nil {
		r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
		return nil, apperrors.ErrDatabase
	}
	return &updatedCat, nil
}

func (r *catRepository) List() ([]models.Cat, error) {
	list := []models.Cat{}
	query := "SELECT * FROM cats;"

	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout)
	defer cancel()

	rows, err := r.QueryContext(ctx, query)

	if err != nil {
		r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
		return nil, apperrors.ErrDatabase
	}

	defer rows.Close()

	for rows.Next() {
		var cat models.Cat
		if err := rows.Scan(
			&cat.ID,
			&cat.Name,
			&cat.YearsOfExperience,
			&cat.Breed,
			&cat.Salary,
			&cat.CreatedAt,
		); err != nil {
			r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
			return nil, apperrors.ErrDatabase
		}
		list = append(list, cat)
	}

	if err := rows.Close(); err != nil {
		r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
		return nil, apperrors.ErrDatabase
	}

	if err := rows.Err(); err != nil {
		r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
		return nil, apperrors.ErrDatabase
	}

	return list, nil
}

func (r *catRepository) Get(id uint) (*models.Cat, error) {
	var res models.Cat
	query := "SELECT * FROM cats WHERE id = $1;"

	row := r.QueryRow(query, id)

	err := row.Scan(
		&res.ID,
		&res.Name,
		&res.YearsOfExperience,
		&res.Breed,
		&res.Salary,
		&res.CreatedAt,
	)

	if err != nil {
		r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))

		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, apperrors.ErrDatabase
	}

	return &res, nil

}
