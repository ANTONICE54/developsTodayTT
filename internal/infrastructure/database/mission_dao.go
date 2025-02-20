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
	missionRepository struct {
		logger logger.Logger
		*sql.DB
	}
)

func NewMissonRepository(customLogger logger.Logger, r *sql.DB) *missionRepository {
	return &missionRepository{
		logger: customLogger,
		DB:     r,
	}
}

func (r *missionRepository) Add(mission models.Mission) (*models.Mission, error) {

	var res models.Mission
	res.TargetList = make([]models.Target, 0)

	query := "INSERT INTO missions (name) VALUES($1) RETURNING id, name, is_completed, created_at;"

	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout)
	defer cancel()

	row := r.QueryRowContext(ctx, query, mission.Name)

	err := row.Scan(
		&res.ID,
		&res.Name,
		&res.IsCompleted,
		&res.CreatedAt,
	)

	if err != nil {
		r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
		return nil, apperrors.ErrDatabase
	}

	for _, v := range mission.TargetList {
		query := "INSERT INTO targets (name, country, notes, mission_id) VALUES ($1, $2, $3, $4) RETURNING id, mission_id, name, country, notes, is_completed, created_at ;"

		row := r.QueryRowContext(ctx, query, v.Name, v.Country, v.Notes, res.ID)

		var target models.Target

		err := row.Scan(
			&target.ID,
			&target.MissionID,
			&target.Name,
			&target.Country,
			&target.Notes,
			&target.IsCompleted,
			&target.CreatedAt,
		)

		if err != nil {
			r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
			return nil, apperrors.ErrDatabase
		}
		res.TargetList = append(res.TargetList, target)
	}

	//r.logger.Warnf("%v", *mission)

	return &res, nil
}

func (r *missionRepository) AssignToCat(missionId, catId uint) error {
	query := "UPDATE missions SET cat_id = $1 WHERE id = $2;"

	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout)
	defer cancel()

	_, err := r.ExecContext(ctx, query, catId, missionId)

	if err != nil {

		r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
		return apperrors.ErrDatabase
	}

	return nil
}

func (r *missionRepository) GetByID(id uint) (*models.Mission, error) {
	var mission models.Mission
	mission.TargetList = make([]models.Target, 0)
	notFound := true
	query := `
		SELECT 
		    m.id, 
		    m.name, 
		    m.cat_id, 
		    m.is_completed, 
		    m.created_at, 
		    t.id, 
		    t.mission_id, 
		    t.name, 
		    t.country, 
		    t.notes, 
		    t.is_completed, 
		    t.created_at
		FROM missions m
		JOIN targets t ON m.id = t.mission_id
		WHERE m.id = $1;
	`

	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout)
	defer cancel()
	rows, err := r.QueryContext(ctx, query, id)

	if err != nil {
		r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
		return nil, apperrors.ErrDatabase
	}

	defer rows.Close()

	for rows.Next() {
		notFound = false
		var target models.Target

		err := rows.Scan(
			&mission.ID,
			&mission.Name,
			&mission.CatId,
			&mission.IsCompleted,
			&mission.CreatedAt,
			&target.ID,
			&target.MissionID,
			&target.Name,
			&target.Country,
			&target.Notes,
			&target.IsCompleted,
			&target.CreatedAt,
		)
		if err != nil {
			r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
			return nil, apperrors.ErrDatabase
		}

		mission.TargetList = append(mission.TargetList, target)
	}

	if notFound {
		return nil, nil
	}

	return &mission, nil
}

func (r *missionRepository) GetByCatID(catID uint) (*models.Mission, error) {
	var mission models.Mission
	mission.TargetList = make([]models.Target, 0)
	notFound := true

	query := `
		SELECT 
		    m.id, 
		    m.name, 
		    m.cat_id, 
		    m.is_completed, 
		    m.created_at, 
		    t.id, 
		    t.mission_id, 
		    t.name, 
		    t.country, 
		    t.notes, 
		    t.is_completed, 
		    t.created_at
		FROM missions m
		JOIN targets t ON m.id = t.mission_id
		WHERE m.cat_id = $1;
	`

	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout)
	defer cancel()

	rows, err := r.QueryContext(ctx, query, catID)

	if err != nil {
		r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
		return nil, apperrors.ErrDatabase
	}

	defer rows.Close()

	for rows.Next() {
		var target models.Target
		notFound = false
		err := rows.Scan(
			&mission.ID,
			&mission.Name,
			&mission.CatId,
			&mission.IsCompleted,
			&mission.CreatedAt,
			&target.ID,
			&target.MissionID,
			&target.Name,
			&target.Country,
			&target.Notes,
			&target.IsCompleted,
			&target.CreatedAt,
		)
		if err != nil {
			r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
			return nil, apperrors.ErrDatabase

		}

		mission.TargetList = append(mission.TargetList, target)
	}

	if notFound {
		return nil, nil
	}

	return &mission, nil
}

func (r *missionRepository) Delete(id uint) error {
	query := "DELETE FROM missions WHERE id = $1;"

	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout)
	defer cancel()

	_, err := r.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
		return apperrors.ErrDatabase
	}

	return nil
}

func (r *missionRepository) List() ([]models.Mission, error) {
	query := `
		SELECT 
		    m.id AS mission_id, 
		    m.name AS mission_name, 
		    m.cat_id, 
		    m.is_completed AS mission_completed, 
		    m.created_at AS mission_created_at, 
		    t.id AS target_id,
		    t.mission_id, 
		    t.name AS target_name, 
		    t.country, 
		    t.notes, 
		    t.is_completed AS target_completed, 
		    t.created_at AS target_created_at
		FROM missions m
		JOIN targets t ON m.id = t.mission_id
		ORDER BY m.id;
	`

	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout)
	defer cancel()

	rows, err := r.QueryContext(ctx, query)
	if err != nil {
		r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
		return nil, apperrors.ErrDatabase
	}
	defer rows.Close()

	missionMap := make(map[uint]*models.Mission)
	missions := make([]models.Mission, 0)

	for rows.Next() {
		var target models.Target
		var mission models.Mission

		err := rows.Scan(
			&mission.ID,
			&mission.Name,
			&mission.CatId,
			&mission.IsCompleted,
			&mission.CreatedAt,
			&target.ID,
			&target.MissionID,
			&target.Name,
			&target.Country,
			&target.Notes,
			&target.IsCompleted,
			&target.CreatedAt,
		)
		if err != nil {
			r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
			return nil, apperrors.ErrDatabase
		}

		if _, exists := missionMap[mission.ID]; !exists {
			missionMap[mission.ID] = &models.Mission{
				ID:          mission.ID,
				CatId:       mission.CatId,
				Name:        mission.Name,
				IsCompleted: mission.IsCompleted,
				CreatedAt:   mission.CreatedAt,
				TargetList:  []models.Target{},
			}
		}

		missionMap[mission.ID].TargetList = append(missionMap[mission.ID].TargetList, target)
	}

	for _, mission := range missionMap {
		missions = append(missions, *mission)
	}

	return missions, nil
}

func (r *missionRepository) Update(id uint, completed bool) error {
	query := "UPDATE missions SET is_completed = $1 WHERE id = $2;"

	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout)
	defer cancel()

	_, err := r.ExecContext(ctx, query, completed, id)

	if err != nil {
		r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
		return apperrors.ErrDatabase
	}

	return nil
}

func (r *missionRepository) GetTarget(id uint) (*models.Target, error) {
	var target models.Target
	query := "SELECT * FROM targets WHERE id = $1"

	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout)
	defer cancel()

	row := r.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&target.ID,
		&target.MissionID,
		&target.Name,
		&target.Country,
		&target.Notes,
		&target.IsCompleted,
		&target.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
		return nil, apperrors.ErrDatabase
	}

	return &target, nil
}

func (r *missionRepository) DeleteTarget(id uint) error {
	query := "DELETE FROM targets WHERE id = $1;"

	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout)
	defer cancel()

	_, err := r.ExecContext(ctx, query, id)
	return err
}

func (r *missionRepository) AddTarget(missionId uint, target models.Target) (*models.Target, error) {
	query := "INSERT INTO targets (name, country, notes, mission_id) VALUES ($1, $2, $3, $4) RETURNING id, mission_id, name, country, notes, is_completed, created_at ;"

	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout)
	defer cancel()

	row := r.QueryRowContext(ctx, query, target.Name, target.Country, target.Notes, missionId)

	var res models.Target

	err := row.Scan(
		&res.ID,
		&res.MissionID,
		&res.Name,
		&res.Country,
		&res.Notes,
		&res.IsCompleted,
		&res.CreatedAt,
	)

	if err != nil {
		r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
		return nil, apperrors.ErrDatabase
	}
	return &res, nil
}

func (r *missionRepository) CompleteTarget(id uint) (*models.Target, error) {
	query := "UPDATE targets SET is_completed = TRUE WHERE id = $1 RETURNING id, mission_id, name, country, notes, is_completed, created_at;"

	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout)
	defer cancel()

	row := r.QueryRowContext(ctx, query, id)
	var res models.Target

	err := row.Scan(
		&res.ID,
		&res.MissionID,
		&res.Name,
		&res.Country,
		&res.Notes,
		&res.IsCompleted,
		&res.CreatedAt,
	)

	if err != nil {
		r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
		return nil, apperrors.ErrDatabase
	}
	return &res, nil
}

func (r *missionRepository) UpdateTargetNotes(id uint, notes string) (*models.Target, error) {
	query := "UPDATE targets SET notes = $1 WHERE id = $2 RETURNING id, mission_id, name, country, notes, is_completed, created_at;"

	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout)
	defer cancel()

	row := r.QueryRowContext(ctx, query, notes, id)
	var res models.Target

	err := row.Scan(
		&res.ID,
		&res.MissionID,
		&res.Name,
		&res.Country,
		&res.Notes,
		&res.IsCompleted,
		&res.CreatedAt,
	)

	if err != nil {
		r.logger.Warnf(apperrors.ErrDatabaseMsg(err.Error()))
		return nil, apperrors.ErrDatabase
	}
	return &res, nil
}
