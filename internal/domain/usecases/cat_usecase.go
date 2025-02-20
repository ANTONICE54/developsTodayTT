package usecases

import (
	"spyCatAgency/internal/domain/models"
	"spyCatAgency/internal/infrastructure/apperrors"
	"spyCatAgency/internal/infrastructure/logger"

	"github.com/lib/pq"
)

type (
	CatRepositoryInterface interface {
		Add(cat models.Cat) (*models.Cat, error)
		Delete(id uint) error
		Update(id uint, salary float64) (*models.Cat, error)
		List() ([]models.Cat, error)
		Get(id uint) (*models.Cat, error)
	}

	catUseCase struct {
		logger        logger.Logger
		catRepository CatRepositoryInterface
	}
)

func NewCatUseCase(customLogger logger.Logger, catRepo CatRepositoryInterface) *catUseCase {
	return &catUseCase{
		logger:        customLogger,
		catRepository: catRepo,
	}
}

func (uc *catUseCase) HireCat(cat models.Cat) (*models.Cat, error) {
	hiredCat, err := uc.catRepository.Add(cat)

	if err != nil {
		return nil, err
	}

	return hiredCat, nil

}

func (uc *catUseCase) FireCat(catID uint) error {
	cat, err := uc.catRepository.Get(catID)

	if cat == nil && err == nil {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("There is no cat with such id"))
		return apperrors.ErrBadRequestf("There is no cat with such id")
	}

	if err != nil {
		return err
	}

	err = uc.catRepository.Delete(catID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation":

				return apperrors.ErrBadRequestf("You cannot fire cat, while it is on mission")
			}
		}

		return apperrors.ErrDatabase
	}

	return nil

}

func (uc *catUseCase) UpdateSalary(catID uint, salary float64) (*models.Cat, error) {
	cat, _ := uc.catRepository.Get(catID)

	if cat == nil {

		return nil, apperrors.ErrBadRequestf("There is no cat with such id")
	}

	updatedCat, err := uc.catRepository.Update(catID, salary)

	return updatedCat, err

}

func (uc *catUseCase) List() ([]models.Cat, error) {
	list, err := uc.catRepository.List()

	if err != nil {
		return nil, err
	}

	return list, nil
}

func (uc *catUseCase) Get(catID uint) (*models.Cat, error) {

	cat, err := uc.catRepository.Get(catID)

	if cat == nil && err == nil {
		uc.logger.Warnf(apperrors.ErrBadRequestMsg("There is no cat with such id"))
		return nil, apperrors.ErrBadRequestf("There is no cat with such id")
	}

	if err != nil {
		return nil, err
	}

	return cat, nil

}
