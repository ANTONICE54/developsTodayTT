package server

import (
	"encoding/json"
	"io"
	"net/http"
	"spyCatAgency/internal/infrastructure/logger"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type (
	CatHandlerInterface interface {
		Hire(ctx *gin.Context)
		Fire(ctx *gin.Context)
		UpdateSalary(ctx *gin.Context)
		List(ctx *gin.Context)
		Get(ctx *gin.Context)
	}

	MissionHandlerInterface interface {
		Add(ctx *gin.Context)
		Update(ctx *gin.Context)
		Get(ctx *gin.Context)
		Delete(ctx *gin.Context)
		List(ctx *gin.Context)
		GetTarget(ctx *gin.Context)
		DeleteTarget(ctx *gin.Context)
		AddTarget(ctx *gin.Context)
		UpdateTarget(ctx *gin.Context)
	}

	server struct {
		logger         logger.Logger
		router         *gin.Engine
		catHandler     CatHandlerInterface
		missionHandler MissionHandlerInterface
	}
)

func New(customLogger logger.Logger, catH CatHandlerInterface, missionH MissionHandlerInterface) *server {

	s := &server{
		logger:         customLogger,
		router:         gin.Default(),
		catHandler:     catH,
		missionHandler: missionH,
	}

	s.setUpRoutes()
	s.addBreedValidator()

	return s
}

func (s *server) setUpRoutes() {
	catRoutes := s.router.Group("/cats")
	catRoutes.POST("", s.catHandler.Hire)
	catRoutes.DELETE("/:id", s.catHandler.Fire)
	catRoutes.GET("", s.catHandler.List)
	catRoutes.GET("/:id", s.catHandler.Get)
	catRoutes.PATCH("/:id", s.catHandler.UpdateSalary)

	missionRoutes := s.router.Group("/missions")
	missionRoutes.POST("", s.missionHandler.Add)
	missionRoutes.PATCH("/:id", s.missionHandler.Update)
	missionRoutes.GET("/:id", s.missionHandler.Get)
	missionRoutes.DELETE("/:id", s.missionHandler.Delete)
	missionRoutes.GET("", s.missionHandler.List)

	targetRoutes := s.router.Group("targets")
	targetRoutes.GET("/:id", s.missionHandler.GetTarget)
	targetRoutes.DELETE("/:id", s.missionHandler.DeleteTarget)
	targetRoutes.POST("", s.missionHandler.AddTarget)
	targetRoutes.PATCH("/:id", s.missionHandler.UpdateTarget)

}

func (s *server) Run(serverPort string) {
	if err := s.router.Run(":" + serverPort); err != nil {
		s.logger.Fatalf("Failed to run server %v", err.Error())
	}
}

type BreedName struct {
	Name string `json:"name"`
}

func (s *server) addBreedValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		breedList, err := s.fetchBreedList()
		if err != nil {
			s.logger.Warnf("Failed to fetch breed info: %s", err.Error())
			return
		}

		v.RegisterValidation("breed", func(fl validator.FieldLevel) bool {
			breed, ok := fl.Field().Interface().(string)
			if !ok {
				return false
			}

			for _, v := range breedList {
				if v.Name == breed {
					return true
				}
			}
			return false
		})
	}
}

func (s *server) fetchBreedList() ([]BreedName, error) {
	url := "https://api.thecatapi.com/v1/breeds"

	client := &http.Client{Timeout: 5 * time.Second}
	queryRes, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer queryRes.Body.Close()
	body, err := io.ReadAll(queryRes.Body)
	if err != nil {
		return nil, err
	}

	var breedList []BreedName
	err = json.Unmarshal(body, &breedList)
	if err != nil {
		return nil, err
	}

	return breedList, nil
}
