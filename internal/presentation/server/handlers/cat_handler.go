package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"spyCatAgency/internal/domain/models"
	"spyCatAgency/internal/infrastructure/apperrors"
	"spyCatAgency/internal/infrastructure/logger"
	"strconv"

	"github.com/gin-gonic/gin"
)

type (
	CatUseCaseInterface interface {
		HireCat(cat models.Cat) (*models.Cat, error)
		FireCat(catID uint) error
		UpdateSalary(catID uint, salary float64) (*models.Cat, error)
		List() ([]models.Cat, error)
		Get(catID uint) (*models.Cat, error)
	}

	catHandler struct {
		logger     logger.Logger
		catUseCase CatUseCaseInterface
	}

	HireCatRequest struct {
		Name              string  `json:"name" binding:"required,alpha"`
		YearsOfExperience uint    `json:"years_of_experience" binding:"required,numeric"`
		Breed             string  `json:"breed" binding:"required,alpha,breed"`
		Salary            float64 `json:"salary" binding:"required,numeric"`
	}

	CatResponse struct {
		ID                uint    `json:"id"`
		Name              string  `json:"name"`
		YearsOfExperience uint    `json:"years_of_experience"`
		Breed             string  `json:"breed"`
		Salary            float64 `json:"salary"`
	}

	UpdateSalaryRequest struct {
		Salary float64 `json:"salary" binding:"required,numeric,gt=0"`
	}

	ListCatsResponse struct {
		List []CatResponse `json:"list"`
	}
)

func NewCatHandler(customLogger logger.Logger, catUC CatUseCaseInterface) *catHandler {
	return &catHandler{
		logger:     customLogger,
		catUseCase: catUC,
	}
}

func (h *catHandler) Hire(ctx *gin.Context) {

	var catInfo HireCatRequest

	if err := ctx.ShouldBindJSON(&catInfo); err != nil {
		h.logger.Warnf("Couldn't bind request: %s", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Couldn't bind request: %s", err.Error()),
		})
		return
	}

	cat, err := h.catUseCase.HireCat(catInfo.mapToCatObj())
	if err != nil {
		var httpErr *apperrors.AppError
		if errors.As(err, &httpErr) {
			ctx.JSON(httpErr.Status(), httpErr.Message)
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	var resp CatResponse

	resp.parseFromCatObj(cat)

	ctx.JSON(http.StatusOK, &resp)
}

func (req *HireCatRequest) mapToCatObj() models.Cat {
	return models.Cat{
		Name:              req.Name,
		YearsOfExperience: req.YearsOfExperience,
		Breed:             req.Breed,
		Salary:            req.Salary,
	}
}

func (resp *CatResponse) parseFromCatObj(cat *models.Cat) {
	resp.ID = cat.ID
	resp.Name = cat.Name
	resp.YearsOfExperience = cat.YearsOfExperience
	resp.Breed = cat.Breed
	resp.Salary = cat.Salary
}

func (h *catHandler) Fire(ctx *gin.Context) {

	catIDstr := ctx.Param("id")
	catID, err := strconv.ParseUint(catIDstr, 10, 32)

	if err != nil {
		h.logger.Warnf("Failed to parse cat id to integer:%s", err.Error())
		ctx.JSON(apperrors.ErrInternal.Status(), apperrors.ErrInternal.Message)
		return
	}

	err = h.catUseCase.FireCat(uint(catID))

	if err != nil {
		var httpErr *apperrors.AppError
		if errors.As(err, &httpErr) {
			ctx.JSON(httpErr.Status(), httpErr.Message)
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	ctx.Status(http.StatusOK)
}

func (h *catHandler) UpdateSalary(ctx *gin.Context) {

	catIDstr := ctx.Param("id")
	catID, err := strconv.ParseUint(catIDstr, 10, 32)

	if err != nil {
		h.logger.Warnf("Failed to parse cat id to integer:%s", err.Error())
		ctx.JSON(apperrors.ErrInternal.Status(), apperrors.ErrInternal.Message)
		return
	}

	var req UpdateSalaryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("Couldn't bind request: %s", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Couldn't bind request",
		})
		return
	}

	updatedCat, err := h.catUseCase.UpdateSalary(uint(catID), req.Salary)
	if err != nil {
		var httpErr *apperrors.AppError
		if errors.As(err, &httpErr) {
			ctx.JSON(httpErr.Status(), httpErr.Message)
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	var resp CatResponse
	resp.parseFromCatObj(updatedCat)
	ctx.JSON(http.StatusOK, &resp)
}

func (h *catHandler) List(ctx *gin.Context) {
	var resp ListCatsResponse

	list, err := h.catUseCase.List()

	if err != nil {
		var httpErr *apperrors.AppError
		if errors.As(err, &httpErr) {
			ctx.JSON(httpErr.Status(), httpErr.Message)
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	respList := make([]CatResponse, 0)
	for _, cat := range list {

		var catResp CatResponse
		catResp.parseFromCatObj(&cat)
		respList = append(respList, catResp)
	}

	resp.List = respList
	ctx.JSON(http.StatusOK, &resp)
}

func (h *catHandler) Get(ctx *gin.Context) {
	catIDstr := ctx.Param("id")
	catID, err := strconv.ParseUint(catIDstr, 10, 32)

	if err != nil {
		h.logger.Warnf("Failed to parse cat id to integer:%s", err.Error())
		ctx.JSON(apperrors.ErrInternal.Status(), apperrors.ErrInternal.Message)
		return
	}

	cat, err := h.catUseCase.Get(uint(catID))

	if err != nil {
		var httpErr *apperrors.AppError
		if errors.As(err, &httpErr) {
			ctx.JSON(httpErr.Status(), httpErr.Message)
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	var resp CatResponse

	resp.parseFromCatObj(cat)

	ctx.JSON(http.StatusOK, &resp)
}
