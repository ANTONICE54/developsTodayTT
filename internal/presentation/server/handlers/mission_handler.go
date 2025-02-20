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
	MissionUseCaseInterface interface {
		Create(mission models.Mission) (*models.Mission, error)
		Assign(missionId, catID uint) (*models.Mission, error)
		Get(id uint) (*models.Mission, error)
		Delete(id uint) error
		ListMissions() ([]models.Mission, error)
		Update(id uint, completed bool) (*models.Mission, error)
		GetTarget(id uint) (*models.Target, error)
		DeleteTarget(id uint) error
		AddTarget(missionId uint, target models.Target) (*models.Target, error)
		CompleteTarget(id uint) (*models.Target, error)
		UpdateTargetNotes(id uint, notes string) (*models.Target, error)
	}

	misionHandler struct {
		logger         logger.Logger
		missionUseCase MissionUseCaseInterface
	}

	TargetRequest struct {
		MissionID uint   `json:"mission_id" `
		Name      string `json:"name" binding:"required,alpha"`
		Country   string `json:"country" binding:"required,alpha"`
		Notes     string `json:"notes"`
	}

	AddMissionRequest struct {
		Name       string          `json:"name" binding:"required,alpha"`
		CatId      *uint           `json:"cat_id"`
		TargetList []TargetRequest `json:"target_list" binding:"required"`
	}

	TargetResponse struct {
		ID          uint   `json:"id"`
		MissionID   uint   `json:"mission_id" `
		Name        string `json:"name" binding:"required,alpha"`
		Country     string `json:"country" binding:"required,alpha"`
		Notes       string `json:"notes"`
		IsCompleted bool   `json:"is_completed"`
	}

	MissionResponse struct {
		ID          uint             `json:"id"`
		Name        string           `json:"name" binding:"required,alpha"`
		CatId       *uint            `json:"cat_id"`
		TargetList  []TargetResponse `json:"target_list" binding:"required"`
		IsCompleted bool             `json:"is_completed"`
	}

	PatchRequest struct {
		CatID       *uint `json:"cat_id,omitempty"`
		IsCompleted *bool `json:"is_completed,omitempty" `
	}

	ListMissionsResponse struct {
		List []MissionResponse `json:"list"`
	}

	AddTargetRequest struct {
		MissionID uint          `json:"mission_id" binding:"required,numeric,gt=0"`
		TargetObj TargetRequest `json:"target" binding:"required"`
	}

	UpdateTargetRequest struct {
		Notes *string `json:"notes,omitempty"`
	}
)

func NewMisionHandler(customLogger logger.Logger, missionUC MissionUseCaseInterface) *misionHandler {
	return &misionHandler{
		logger:         customLogger,
		missionUseCase: missionUC,
	}
}

func (h *misionHandler) Add(ctx *gin.Context) {
	var req AddMissionRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("Couldn't bind request: %s", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Couldn't bind request: %s", err.Error()),
		})
		return
	}

	createdMission, err := h.missionUseCase.Create(*req.mapToMissionObj())

	if err != nil {
		var httpErr *apperrors.AppError
		if errors.As(err, &httpErr) {
			ctx.JSON(httpErr.Status(), httpErr.Message)
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	var resp MissionResponse
	resp.parseFromMissionObj(*createdMission)

	ctx.JSON(http.StatusOK, &resp)

}

func (h *misionHandler) Update(ctx *gin.Context) {

	missionIDstr := ctx.Param("id")
	missionID, err := strconv.ParseUint(missionIDstr, 10, 32)

	if err != nil {
		h.logger.Warnf("Failed to parse mission id to integer:%s", err.Error())
		ctx.JSON(apperrors.ErrInternal.Status(), apperrors.ErrInternal.Message)
		return
	}

	var req PatchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("Couldn't bind request: %s", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Couldn't bind request: %s", err.Error()),
		})
		return
	}

	if req.CatID == nil && req.IsCompleted == nil {
		h.logger.Warnf("Bad request: cat id and is completed fields missing in patch request")
		ctx.JSON(apperrors.ErrBadRequest.Status(), apperrors.ErrBadRequest.Message)
		return

	} else if req.CatID != nil && req.IsCompleted != nil {
		h.logger.Warnf("Bad request: cat id AND is completed fields in patch request")
		ctx.JSON(apperrors.ErrBadRequest.Status(), apperrors.ErrBadRequest.Message)
		return
	} else if req.CatID != nil {

		mission, err := h.missionUseCase.Assign(uint(missionID), *req.CatID)

		if err != nil {
			var httpErr *apperrors.AppError
			if errors.As(err, &httpErr) {
				ctx.JSON(httpErr.Status(), httpErr.Message)
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			return
		}

		var resp MissionResponse

		resp.parseFromMissionObj(*mission)

		ctx.JSON(http.StatusOK, &resp)
	} else if req.IsCompleted != nil {
		mission, err := h.missionUseCase.Update(uint(missionID), *req.IsCompleted)

		if err != nil {
			var httpErr *apperrors.AppError
			if errors.As(err, &httpErr) {
				ctx.JSON(httpErr.Status(), httpErr.Message)
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			return
		}

		var resp MissionResponse

		resp.parseFromMissionObj(*mission)

		ctx.JSON(http.StatusOK, &resp)

	}

}

func (h *misionHandler) Get(ctx *gin.Context) {

	missionIDstr := ctx.Param("id")
	missionID, err := strconv.ParseUint(missionIDstr, 10, 32)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	mission, err := h.missionUseCase.Get(uint(missionID))

	if err != nil {
		var httpErr *apperrors.AppError
		if errors.As(err, &httpErr) {
			ctx.JSON(httpErr.Status(), httpErr.Message)
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	var resp MissionResponse

	resp.parseFromMissionObj(*mission)

	ctx.JSON(http.StatusOK, &resp)
}

func (h *misionHandler) Delete(ctx *gin.Context) {

	missionIDstr := ctx.Param("id")
	missionID, err := strconv.ParseUint(missionIDstr, 10, 32)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	err = h.missionUseCase.Delete(uint(missionID))
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

func (h *misionHandler) List(ctx *gin.Context) {
	var resp ListMissionsResponse

	list, err := h.missionUseCase.ListMissions()
	if err != nil {
		var httpErr *apperrors.AppError
		if errors.As(err, &httpErr) {
			ctx.JSON(httpErr.Status(), httpErr.Message)
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	respList := make([]MissionResponse, 0)
	for _, mission := range list {
		var missionResp MissionResponse
		missionResp.parseFromMissionObj(mission)
		respList = append(respList, missionResp)
	}

	resp.List = respList

	ctx.JSON(http.StatusOK, &resp)
}

func (h *misionHandler) GetTarget(ctx *gin.Context) {

	targetIDstr := ctx.Param("id")
	targetID, err := strconv.ParseUint(targetIDstr, 10, 32)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	target, err := h.missionUseCase.GetTarget(uint(targetID))

	if err != nil {
		var httpErr *apperrors.AppError
		if errors.As(err, &httpErr) {
			ctx.JSON(httpErr.Status(), httpErr.Message)
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	var resp TargetResponse

	resp.parseFromTargetObj(*target)

	ctx.JSON(http.StatusOK, &resp)
}

func (h *misionHandler) DeleteTarget(ctx *gin.Context) {

	targetIDstr := ctx.Param("id")
	targetID, err := strconv.ParseUint(targetIDstr, 10, 32)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	err = h.missionUseCase.DeleteTarget(uint(targetID))

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

func (h *misionHandler) AddTarget(ctx *gin.Context) {
	var req AddTargetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("Couldn't bind request: %s", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Couldn't bind request: %s", err.Error()),
		})
		return
	}

	targetObj := req.TargetObj.mapToTargetObj()

	target, err := h.missionUseCase.AddTarget(req.MissionID, *targetObj)

	if err != nil {
		var httpErr *apperrors.AppError
		if errors.As(err, &httpErr) {
			ctx.JSON(httpErr.Status(), httpErr.Message)
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	var resp TargetResponse

	resp.parseFromTargetObj(*target)

	ctx.JSON(http.StatusOK, &resp)
}

func (h *misionHandler) UpdateTarget(ctx *gin.Context) {
	targetIDstr := ctx.Param("id")
	targetID, err := strconv.ParseUint(targetIDstr, 10, 32)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var req UpdateTargetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("Couldn't bind request: %s", err.Error())

	}

	if req.Notes == nil {

		target, err := h.missionUseCase.CompleteTarget(uint(targetID))
		if err != nil {
			var httpErr *apperrors.AppError
			if errors.As(err, &httpErr) {
				ctx.JSON(httpErr.Status(), httpErr.Message)
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			return
		}

		var resp TargetResponse

		resp.parseFromTargetObj(*target)

		ctx.JSON(http.StatusOK, &resp)
	} else {

		target, err := h.missionUseCase.UpdateTargetNotes(uint(targetID), *req.Notes)
		if err != nil {
			var httpErr *apperrors.AppError
			if errors.As(err, &httpErr) {
				ctx.JSON(httpErr.Status(), httpErr.Message)
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			return
		}
		var resp TargetResponse

		resp.parseFromTargetObj(*target)

		ctx.JSON(http.StatusOK, &resp)
	}

}

func (req *AddMissionRequest) mapToMissionObj() *models.Mission {

	var mission models.Mission

	mission.Name = req.Name
	mission.CatId = req.CatId

	mission.TargetList = make([]models.Target, 0)

	for _, target := range req.TargetList {

		targetToAppend := models.Target{
			MissionID: target.MissionID,
			Name:      target.Name,
			Country:   target.Country,
			Notes:     target.Notes,
		}

		mission.TargetList = append(mission.TargetList, targetToAppend)

	}

	return &mission
}

func (resp *MissionResponse) parseFromMissionObj(mission models.Mission) {

	resp.ID = mission.ID
	resp.Name = mission.Name
	resp.CatId = mission.CatId
	resp.IsCompleted = mission.IsCompleted

	targetResponseList := make([]TargetResponse, 0)

	for _, target := range mission.TargetList {

		targetToAppend := TargetResponse{
			ID:          target.ID,
			MissionID:   target.MissionID,
			Name:        target.Name,
			Country:     target.Country,
			Notes:       target.Notes,
			IsCompleted: target.IsCompleted,
		}
		targetResponseList = append(targetResponseList, targetToAppend)

	}

	resp.TargetList = targetResponseList

}

func (req *TargetRequest) mapToTargetObj() *models.Target {
	return &models.Target{
		MissionID: req.MissionID,
		Name:      req.Name,
		Country:   req.Country,
		Notes:     req.Notes,
	}
}

func (resp *TargetResponse) parseFromTargetObj(target models.Target) {
	resp.ID = target.ID
	resp.MissionID = target.MissionID
	resp.Name = target.Name
	resp.Country = target.Country
	resp.Notes = target.Notes
	resp.IsCompleted = target.IsCompleted
}
