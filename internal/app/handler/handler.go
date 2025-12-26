package handler

import (
	"net/http"

	"metoda/internal/app/auth"
	"metoda/internal/app/repository"
	"metoda/internal/app/storage"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Repository *repository.Repository
	Minio      *storage.MinioClient
}

func NewHandler(r *repository.Repository, minioClient *storage.MinioClient) *Handler {
	return &Handler{
		Repository: r,
		Minio:      minioClient,
	}
}

// RegisterHandler Функция, в которой мы отдельно регистрируем маршруты
func (h *Handler) RegisterHandler(router *gin.Engine) {
	// Обработчик корневого маршрута
	router.GET("/", h.RootHandler)

	// API маршруты с middleware авторизации
	api := router.Group("/api")
	api.Use(auth.AuthMiddleware()) // Добавляем middleware для всех API маршрутов
	{
		// Домен услуги (Medication)
		medications := api.Group("/medications")
		{
			medications.GET("", h.GetMedicationsAPI)                         // GET список с фильтрацией
			medications.GET("/:id", h.GetMedicationByIDAPI)                  // GET одна запись
			medications.POST("", h.CreateMedicationAPI)                      // POST добавление (без изображения)
			medications.PUT("/:id", h.UpdateMedicationAPI)                   // PUT изменение
			medications.DELETE("/:id", h.DeleteMedicationAPI)                // DELETE удаление
			medications.POST("/:id/add-to-draft", h.AddMedicationToDraftAPI) // POST добавление в заявку-черновик
			medications.POST("/:id/image", h.UploadMedicationImageAPI)       // POST добавление изображения
		}

		// Домен заявки (Prescription)
		prescriptions := api.Group("/prescriptions")
		{
			prescriptions.GET("/cart-icon", h.GetCartIconAPI) // GET иконки корзины
			prescriptions.GET("", h.GetPrescriptionsAPI)      // GET список с фильтрацией
			// Специфичные маршруты должны быть ПЕРЕД общими маршрутами с :id
			prescriptions.PUT("/:id/form", h.FormPrescriptionAPI)                                      // PUT сформировать создателем
			prescriptions.PUT("/:id/complete", h.CompleteOrRejectPrescriptionAPI)                      // PUT завершить модератором
			prescriptions.PUT("/:id/reject", h.CompleteOrRejectPrescriptionAPI)                        // PUT отклонить модератором
			prescriptions.DELETE("/:id/medications/:medication_id", h.DeletePrescriptionMedicationAPI) // DELETE удаление из заявки
			prescriptions.PUT("/:id/medications/:medication_id", h.UpdatePrescriptionMedicationAPI)    // PUT изменение
			// Общие маршруты с :id должны быть в конце
			prescriptions.GET("/:id", h.GetPrescriptionByIDAPI)   // GET одна запись
			prescriptions.PUT("/:id", h.UpdatePrescriptionAPI)    // PUT изменения полей
			prescriptions.DELETE("/:id", h.DeletePrescriptionAPI) // DELETE удаление
		}

		// Домен пользователь (User)
		users := api.Group("/users")
		{
			users.POST("/register", h.RegisterUserAPI)    // POST регистрация
			users.GET("/profile", h.GetUserProfileAPI)    // GET полей пользователя
			users.PUT("/profile", h.UpdateUserProfileAPI) // PUT пользователя
			users.POST("/auth", h.AuthenticateUserAPI)    // POST аутентификация
			users.POST("/logout", h.LogoutUserAPI)        // POST деавторизация
		}
	}
}

// RootHandler - обработчик корневого маршрута
func (h *Handler) RootHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "API Server is running",
		"version": "1.0",
		"endpoints": gin.H{
			"api_base": "/api",
			"medications": gin.H{
				"GET    /api/medications":                    "Список услуг с фильтрацией",
				"GET    /api/medications/:id":                "Одна запись услуги",
				"POST   /api/medications":                    "Добавление услуги (без изображения)",
				"PUT    /api/medications/:id":                 "Изменение услуги",
				"DELETE /api/medications/:id":                "Удаление услуги",
				"POST   /api/medications/:id/add-to-draft":  "Добавление в заявку-черновик",
				"POST   /api/medications/:id/image":         "Добавление изображения",
			},
			"prescriptions": gin.H{
				"GET    /api/prescriptions/cart-icon":                    "Иконка корзины",
				"GET    /api/prescriptions":                               "Список заявок с фильтрацией",
				"GET    /api/prescriptions/:id":                           "Одна запись заявки",
				"PUT    /api/prescriptions/:id":                           "Изменение полей заявки",
				"PUT    /api/prescriptions/:id/form":                      "Сформировать создателем",
				"PUT    /api/prescriptions/:id/complete":                  "Завершить модератором",
				"PUT    /api/prescriptions/:id/reject":                   "Отклонить модератором",
				"DELETE /api/prescriptions/:id":                          "Удаление заявки",
				"DELETE /api/prescriptions/:id/medications/:medication_id": "Удаление из заявки",
				"PUT    /api/prescriptions/:id/medications/:medication_id": "Изменение в м-м",
			},
			"users": gin.H{
				"POST /api/users/register": "Регистрация",
				"GET  /api/users/profile":   "Поля пользователя",
				"PUT  /api/users/profile":  "Обновление пользователя",
				"POST /api/users/auth":     "Аутентификация",
				"POST /api/users/logout":   "Деавторизация",
			},
		},
	})
}

