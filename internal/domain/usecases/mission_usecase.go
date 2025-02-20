package usecases

import (
	"spyCatAgency/internal/domain/models"
	"spyCatAgency/internal/infrastructure/apperrors"
	"spyCatAgency/internal/infrastructure/logger"
)

type (
	MissionRepositoryInterface interface {
		Add(mission models.Mission) (*models.Mission, error)
		AssignToCat(missionId, catId uint) error
		GetByID(id uint) (*models.Mission, error)
		GetByCatID(catID uint) (*models.Mission, error)
		Delete(id uint) error
		List() ([]models.Mission, error)
		Update(id uint, completed bool) error
		GetTarget(id uint) (*models.Target, error)
		DeleteTarget(id uint) error
		AddTarget(missionId uint, target models.Target) (*models.Target, error)
		CompleteTarget(id uint) (*models.Target, error)
		UpdateTargetNotes(id uint, notes string) (*models.Target, error)
	}

	missionUseCase struct {
		logger            logger.Logger
		missionRepository MissionRepositoryInterface
		catRepository     CatRepositoryInterface
	}
)

func NewMissionUseCase(customLogger logger.Logger, missionRepo MissionRepositoryInterface, catRepo CatRepositoryInterface) *missionUseCase {
	return &missionUseCase{
		logger:            customLogger,
		missionRepository: missionRepo,
		catRepository:     catRepo,
	}
}

func (uc *missionUseCase) Create(mission models.Mission) (*models.Mission, error) {

	if len(mission.TargetList) > 3 || len(mission.TargetList) < 1 {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("Target limit exceeded"))
		return nil, apperrors.ErrBadRequestf("Target limit exceeded")
	}

	if mission.CatId != nil {

		cat, err := uc.catRepository.Get(*mission.CatId)

		if cat == nil && err == nil {
			uc.logger.Warnf(apperrors.ErrBadRequestMsg("There is no cat with such id"))
			return nil, apperrors.ErrBadRequestf("There is no cat with such id")
		}

		catMission, err := uc.missionRepository.GetByCatID(*mission.CatId)

		if err != nil {
			return nil, err
		}

		if catMission != nil {
			uc.logger.Warnf(apperrors.ErrBadRequestMsg("This cat has already been assigned a mission"))
			return nil, apperrors.ErrBadRequestf("This cat has already been assigned a mission")
		}
	}

	createdMission, err := uc.missionRepository.Add(mission)
	if err != nil {
		return nil, err
	}

	if mission.CatId != nil {

		err = uc.missionRepository.AssignToCat(createdMission.ID, *mission.CatId)
		if err != nil {
			return nil, err
		}

		createdMission.CatId = mission.CatId
	}

	return createdMission, nil
}

func (uc *missionUseCase) Assign(missionId, catID uint) (*models.Mission, error) {
	mission, err := uc.missionRepository.GetByID(missionId)
	if err == nil && mission == nil {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("There is no mission with such id"))
		return nil, apperrors.ErrBadRequestf("There is no mission with such id")

	} else if err != nil {
		return nil, err
	}

	if mission.CatId != nil {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("Already assigned"))
		return nil, apperrors.ErrBadRequestf("Already assigned")
	}

	cat, err := uc.catRepository.Get(catID)

	if cat == nil && err == nil {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("There is no cat with such id"))
		return nil, apperrors.ErrBadRequestf("There is no cat with such id")
	}

	catMission, err := uc.missionRepository.GetByCatID(catID)

	uc.logger.Warnf("Cat mission %v", catMission)

	if err != nil {

		return nil, err
	}

	if catMission != nil {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("This cat has already been assigned a mission"))
		return nil, apperrors.ErrBadRequestf("This cat has already been assigned a mission")
	}

	err = uc.missionRepository.AssignToCat(missionId, catID)

	if err != nil {
		return nil, err
	}

	mission, err = uc.missionRepository.GetByID(missionId)

	if err != nil {
		return nil, err
	}

	return mission, nil
}

func (uc *missionUseCase) Get(id uint) (*models.Mission, error) {
	mission, err := uc.missionRepository.GetByID(id)

	if err == nil && mission == nil {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("There is no mission with such id"))
		return nil, apperrors.ErrBadRequestf("There is no mission with such id")

	} else if err != nil {
		return nil, err
	}
	return mission, nil
}

func (uc *missionUseCase) Delete(id uint) error {

	mission, err := uc.missionRepository.GetByID(id)
	if err == nil && mission == nil {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("There is no mission with such id"))
		return apperrors.ErrBadRequestf("There is no mission with such id")

	} else if err != nil {
		return err
	}

	if mission.CatId != nil {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("Assigned mission cannot be deleted"))
		return apperrors.ErrBadRequestf("Assigned mission cannot be deleted")
	}

	err = uc.missionRepository.Delete(id)

	if err != nil {
		return err
	}

	return nil
}

func (uc *missionUseCase) ListMissions() ([]models.Mission, error) {
	list, err := uc.missionRepository.List()

	if err != nil {
		return nil, err
	}

	return list, nil
}

func (uc *missionUseCase) Update(id uint, completed bool) (*models.Mission, error) {
	mission, err := uc.missionRepository.GetByID(id)
	if err == nil && mission == nil {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("There is no mission with such id"))
		return nil, apperrors.ErrBadRequestf("There is no mission with such id")

	} else if err != nil {
		return nil, err
	}

	if mission.IsCompleted {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("Completed mission cannot be updated"))
		return nil, apperrors.ErrBadRequestf("Completed mission cannot be updated")
	}

	err = uc.missionRepository.Update(id, completed)

	if err != nil {
		return nil, err
	}

	mission, err = uc.missionRepository.GetByID(id)

	if err != nil {
		return nil, err
	}

	return mission, nil
}

func (uc *missionUseCase) GetTarget(id uint) (*models.Target, error) {
	target, err := uc.missionRepository.GetTarget(id)

	if err == nil && target == nil {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("There is no target with such id"))
		return nil, apperrors.ErrBadRequestf("There is no target with such id")
	} else if err != nil {
		return nil, err
	}

	return target, nil
}

func (uc *missionUseCase) DeleteTarget(id uint) error {

	target, err := uc.missionRepository.GetTarget(id)

	if err == nil && target == nil {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("There is no target with such id"))
		return apperrors.ErrBadRequestf("There is no target with such id")
	} else if err != nil {
		return err
	}
	if target.IsCompleted {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("Completed target cannot be deleted"))
		return apperrors.ErrBadRequestf("Completed target cannot be deleted")
	}

	mission, err := uc.missionRepository.GetByID(target.MissionID)

	if err != nil {
		return err
	}

	if mission.IsCompleted {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("Targets cannot be deleted from completed missions"))
		return apperrors.ErrBadRequestf("Targets cannot be deleted from completed missions")
	}

	if len(mission.TargetList) == 1 {

		uc.logger.Warnf(apperrors.ErrBadRequestMsg("You cannot delete last targer of the mission"))
		return apperrors.ErrBadRequestf("You cannot delete last targer of the mission")

	}

	err = uc.missionRepository.DeleteTarget(id)

	if err != nil {
		return err
	}

	return nil
}

func (uc *missionUseCase) AddTarget(missionId uint, target models.Target) (*models.Target, error) {
	mission, err := uc.missionRepository.GetByID(missionId)
	if err == nil && mission == nil {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("There is no mission with such id"))
		return nil, apperrors.ErrBadRequestf("There is no mission with such id")

	} else if err != nil {
		return nil, err
	}

	if mission.IsCompleted {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("Completed mission cannot be updated with new targets"))
		return nil, apperrors.ErrBadRequestf("Completed mission cannot be updated with new targets")
	}

	if len(mission.TargetList) == 3 {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("Target limit exceeded"))
		return nil, apperrors.ErrBadRequestf("Target limit exceeded")
	}

	createdTarget, err := uc.missionRepository.AddTarget(missionId, target)

	if err != nil {
		return nil, err
	}

	return createdTarget, nil
}

func (uc *missionUseCase) CompleteTarget(id uint) (*models.Target, error) {
	var allTargetsCompleted = true

	target, err := uc.missionRepository.GetTarget(id)

	if err == nil && target == nil {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("There is no target with such id"))
		return nil, apperrors.ErrBadRequestf("There is no target with such id")
	} else if err != nil {
		return nil, err
	}

	if target.IsCompleted {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("Completed target cannot be updated"))
		return nil, apperrors.ErrBadRequestf("Completed target cannot be updated")
	}

	mission, err := uc.missionRepository.GetByID(target.MissionID)
	if err == nil && mission == nil {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("There is no mission with such id"))
		return nil, apperrors.ErrBadRequestf("There is no mission with such id")

	} else if err != nil {
		return nil, err
	}

	if mission.IsCompleted {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("Target of completed mission cannot be updated"))
		return nil, apperrors.ErrBadRequestf("Target of completed mission cannot be updated")
	}

	updatedTarget, err := uc.missionRepository.CompleteTarget(id)

	if err != nil {
		return nil, err
	}

	for _, v := range mission.TargetList {
		if v.ID == id {
			continue
		}
		if !v.IsCompleted {
			allTargetsCompleted = false
			break
		}
	}

	uc.logger.Warnf("%v", allTargetsCompleted)

	if allTargetsCompleted {
		err = uc.missionRepository.Update(mission.ID, true)

		if err != nil {
			return nil, err
		}
	}

	return updatedTarget, nil
}

func (uc *missionUseCase) UpdateTargetNotes(id uint, notes string) (*models.Target, error) {
	target, err := uc.missionRepository.GetTarget(id)

	if err == nil && target == nil {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("There is no target with such id"))
		return nil, apperrors.ErrBadRequestf("There is no target with such id")
	} else if err != nil {
		return nil, err
	}

	if target.IsCompleted {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("Completed target cannot be updated"))
		return nil, apperrors.ErrBadRequestf("Completed target cannot be updated")
	}

	mission, err := uc.missionRepository.GetByID(id)
	if err == nil && mission == nil {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("There is no mission with such id"))
		return nil, apperrors.ErrBadRequestf("There is no mission with such id")

	} else if err != nil {
		return nil, err
	}

	if mission.IsCompleted {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("Completed mission cannot be updated"))
		return nil, apperrors.ErrBadRequestf("Completed mission cannot be updated")
	}

	target, err = uc.missionRepository.UpdateTargetNotes(id, notes)

	if err != nil {
		return nil, err
	}

	return target, nil

}
