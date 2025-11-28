package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"metoda/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// RegisterHandler Функция, в которой мы отдельно регистрируем маршруты
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/catalog")
	})
	router.GET("/catalog", h.GetMedications)
	router.GET("/service/:id", h.GetMedicationByID)
	router.GET("/order", h.GetPrescription)
	router.POST("/add-medication", h.AddMedicationToPrescription)
	router.POST("/delete-prescription", h.DeletePrescription)
}

// RegisterStatic То же самое, что и с маршрутами, регистрируем статику
func (h *Handler) RegisterStatic(router *gin.Engine) {
	// Загружаем все HTML шаблоны
	router.LoadHTMLGlob("templates/*.html")
	logrus.Info("HTML templates loaded from templates/*.html")
	// Регистрируем статические файлы
	router.Static("/static", "./static")
	logrus.Info("Static files registered at /static")
}

// errorHandler для более удобного вывода ошибок
func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

