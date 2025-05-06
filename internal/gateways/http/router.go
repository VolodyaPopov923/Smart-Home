package http

import (
	"encoding/json"
	"errors"
	"homework/internal/usecase"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func setupRouter(r *gin.Engine, uc UseCases, ws *WebSocketHandler) {
	gin.SetMode(gin.ReleaseMode)
	r.HandleMethodNotAllowed = true

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, ErrorResponse{Reason: "Not Found"})
	})

	r.NoMethod(func(c *gin.Context) {
		var allowedMethods string
		path := c.Request.URL.Path

		switch path {
		case "/events":
			allowedMethods = "POST,OPTIONS"
		case "/sensors":
			allowedMethods = "GET,HEAD,POST,OPTIONS"
		case "/users":
			allowedMethods = "POST,OPTIONS"
		default:
			if strings.HasPrefix(path, "/sensors/") {
				allowedMethods = "GET,HEAD,OPTIONS"
			} else if strings.HasPrefix(path, "/users/") && strings.HasSuffix(path, "/sensors") {
				allowedMethods = "GET,HEAD,POST,OPTIONS"
			}
		}

		if allowedMethods != "" {
			c.Header("Allow", allowedMethods)
		}

		c.JSON(http.StatusMethodNotAllowed, ErrorResponse{Reason: "Method Not Allowed"})
	})

	setupEventsRoutes(r, uc)
	setupSensorsRoutes(r, uc, ws)
	setupUsersRoutes(r, uc)

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
}

func setupEventsRoutes(r *gin.Engine, uc UseCases) {
	eventsGroup := r.Group("/events")
	{
		eventsGroup.POST("", func(c *gin.Context) {
			if !checkContentTypeJSON(c) {
				return
			}

			var eventReq SensorEventRequest
			if err := c.ShouldBindJSON(&eventReq); err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Reason: "Invalid request body"})
				return
			}

			if eventReq.SensorSerialNumber == "" || len(eventReq.SensorSerialNumber) != 10 || !isNumeric(eventReq.SensorSerialNumber) {
				c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Reason: "Invalid sensor serial number"})
				return
			}

			event := eventToDomain(eventReq)
			err := uc.Event.ReceiveEvent(c.Request.Context(), event)
			if err != nil {
				handleError(c, err)
				return
			}

			c.Status(http.StatusCreated)
		})

		eventsGroup.OPTIONS("", func(c *gin.Context) {
			setAllowHeader(c, "POST,OPTIONS")
		})
	}
}

func setupSensorsRoutes(r *gin.Engine, uc UseCases, ws *WebSocketHandler) {
	sensorsGroup := r.Group("/sensors")
	{
		sensorsGroup.GET("", func(c *gin.Context) {
			if !checkAcceptJSON(c) {
				return
			}

			sensors, err := uc.Sensor.GetSensors(c.Request.Context())
			if err != nil {
				handleError(c, err)
				return
			}

			c.JSON(http.StatusOK, sensorsToResponse(sensors))
		})

		sensorsGroup.HEAD("", func(c *gin.Context) {
			if !checkAcceptJSON(c) {
				return
			}

			sensors, err := uc.Sensor.GetSensors(c.Request.Context())
			if err != nil {
				handleStatusOnlyError(c, err)
				return
			}

			setContentLength(c, sensorsToResponse(sensors))
			c.Status(http.StatusOK)
		})

		sensorsGroup.POST("", func(c *gin.Context) {
			if !checkContentTypeJSON(c) {
				return
			}

			var sensorCreate SensorCreateRequest
			if err := c.ShouldBindJSON(&sensorCreate); err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Reason: "Invalid request body"})
				return
			}

			if !validateSensorData(c, sensorCreate) {
				return
			}

			sensor := sensorToDomain(sensorCreate)
			result, err := uc.Sensor.RegisterSensor(c.Request.Context(), sensor)
			if err != nil {
				handleError(c, err)
				return
			}

			c.JSON(http.StatusOK, sensorToResponse(result))
		})

		sensorsGroup.OPTIONS("", func(c *gin.Context) {
			setAllowHeader(c, "GET,HEAD,POST,OPTIONS")
		})

		setupSensorByIDRoutes(sensorsGroup, uc, ws)
	}
}

func setupSensorByIDRoutes(rg *gin.RouterGroup, uc UseCases, ws *WebSocketHandler) {
	rg.GET("/:sensor_id", func(c *gin.Context) {
		if !checkAcceptJSON(c) {
			return
		}

		id, err := strconv.ParseInt(c.Param("sensor_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Reason: "Invalid sensor ID"})
			return
		}

		sensor, err := uc.Sensor.GetSensorByID(c.Request.Context(), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, sensorToResponse(sensor))
	})

	rg.HEAD("/:sensor_id", func(c *gin.Context) {
		if !checkAcceptJSON(c) {
			return
		}
		id, err := strconv.ParseInt(c.Param("sensor_id"), 10, 64)
		if err != nil {
			c.Status(http.StatusUnprocessableEntity)
			return
		}

		sensor, err := uc.Sensor.GetSensorByID(c.Request.Context(), id)
		if err != nil {
			handleStatusOnlyError(c, err)
			return
		}

		setContentLength(c, sensorToResponse(sensor))
		c.Status(http.StatusOK)
	})

	rg.GET("/:sensor_id/events", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("sensor_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Reason: "Invalid sensor ID"})
			return
		}

		err = ws.Handle(c, id)
		if err != nil {
			return
		}
	})

	rg.GET("/:sensor_id/history", func(c *gin.Context) {
		if !checkAcceptJSON(c) {
			return
		}

		id, err := strconv.ParseInt(c.Param("sensor_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Reason: "Invalid sensor ID"})
			return
		}

		startDateStr := c.Query("start_date")
		endDateStr := c.Query("end_date")

		var startDate, endDate time.Time
		var parseErr error

		if startDateStr != "" {
			startDate, parseErr = time.Parse(time.RFC3339, startDateStr)
			if parseErr != nil {
				c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Reason: "Invalid start_date format. Use RFC3339 format."})
				return
			}
		}

		if endDateStr != "" {
			endDate, parseErr = time.Parse(time.RFC3339, endDateStr)
			if parseErr != nil {
				c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Reason: "Invalid end_date format. Use RFC3339 format."})
				return
			}
		}

		if endDate.IsZero() {
			endDate = time.Now()
		}

		if startDate.IsZero() {
			startDate = endDate.AddDate(0, -1, 0)
		}

		events, err := uc.Event.GetSensorHistory(c.Request.Context(), id, startDate, endDate)
		if err != nil {
			handleError(c, err)
			return
		}

		metadata := SensorHistoryMetadata{
			RequestTime:     time.Now().Format("2006-01-02 15:04:05"),
			RequestedByUser: "VolodyaPopov923",
		}

		c.JSON(http.StatusOK, eventsToHistoryResponse(events, metadata))
	})

	rg.OPTIONS("/:sensor_id", func(c *gin.Context) {
		setAllowHeader(c, "GET,HEAD,OPTIONS")
	})
	rg.OPTIONS("/:sensor_id/history", func(c *gin.Context) {
		setAllowHeader(c, "GET,OPTIONS")
	})
}

func setupUsersRoutes(r *gin.Engine, uc UseCases) {
	usersGroup := r.Group("/users")
	{
		usersGroup.POST("", func(c *gin.Context) {
			if !checkContentTypeJSON(c) {
				return
			}

			var userCreate UserCreateRequest
			if err := c.ShouldBindJSON(&userCreate); err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Reason: "Invalid request body"})
				return
			}

			if userCreate.Name == "" {
				c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Reason: "User name cannot be empty"})
				return
			}

			user := userToDomain(userCreate)
			result, err := uc.User.RegisterUser(c.Request.Context(), user)
			if err != nil {
				handleError(c, err)
				return
			}

			c.JSON(http.StatusOK, userToResponse(result))
		})

		usersGroup.OPTIONS("", func(c *gin.Context) {
			setAllowHeader(c, "POST,OPTIONS")
		})

		setupUserSensorsRoutes(usersGroup, uc)
	}
}

func setupUserSensorsRoutes(rg *gin.RouterGroup, uc UseCases) {
	userSensorsGroup := rg.Group("/:user_id/sensors")
	{
		userSensorsGroup.GET("", func(c *gin.Context) {
			if !checkAcceptJSON(c) {
				return
			}

			id, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Reason: "Invalid user ID"})
				return
			}

			sensors, err := uc.User.GetUserSensors(c.Request.Context(), id)
			if err != nil {
				if errors.Is(err, usecase.ErrUserNotFound) || strings.Contains(err.Error(), "user not found") {
					c.JSON(http.StatusNotFound, ErrorResponse{Reason: "User not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, ErrorResponse{Reason: "Internal server error"})
				return
			}

			c.JSON(http.StatusOK, sensorsToResponse(sensors))
		})

		userSensorsGroup.HEAD("", func(c *gin.Context) {
			if !checkAcceptJSON(c) {
				return
			}

			id, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
			if err != nil {
				c.Status(http.StatusUnprocessableEntity)
				return
			}

			sensors, err := uc.User.GetUserSensors(c.Request.Context(), id)
			if err != nil {
				if errors.Is(err, usecase.ErrUserNotFound) || strings.Contains(err.Error(), "user not found") {
					c.Status(http.StatusNotFound)
					return
				}
				c.Status(http.StatusInternalServerError)
				return
			}

			setContentLength(c, sensorsToResponse(sensors))
			c.Status(http.StatusOK)
		})

		userSensorsGroup.POST("", func(c *gin.Context) {
			if !checkContentTypeJSON(c) {
				return
			}
			id, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Reason: "Invalid user ID"})
				return
			}

			var binding SensorBindingRequest
			if err := c.ShouldBindJSON(&binding); err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Reason: "Invalid request body"})
				return
			}

			if binding.SensorID <= 0 {
				c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Reason: "Invalid sensor ID"})
				return
			}

			err = uc.User.AttachSensorToUser(c.Request.Context(), id, binding.SensorID)
			if err != nil {
				if errors.Is(err, usecase.ErrUserNotFound) || strings.Contains(err.Error(), "user not found") {
					c.JSON(http.StatusNotFound, ErrorResponse{Reason: "User not found"})
					return
				}
				handleError(c, err)
				return
			}

			c.Status(http.StatusCreated)
		})

		userSensorsGroup.OPTIONS("", func(c *gin.Context) {
			setAllowHeader(c, "GET,HEAD,POST,OPTIONS")
		})
	}
}

func checkContentTypeJSON(c *gin.Context) bool {
	contentType := c.Request.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		c.Status(http.StatusUnsupportedMediaType)
		return false
	}
	return true
}

func checkAcceptJSON(c *gin.Context) bool {
	accept := c.Request.Header.Get("Accept")
	if !isAcceptableType(accept) {
		c.Status(http.StatusNotAcceptable)
		return false
	}
	return true
}

func validateSensorData(c *gin.Context, sensor SensorCreateRequest) bool {
	if sensor.SerialNumber == "" || len(sensor.SerialNumber) != 10 || !isNumeric(sensor.SerialNumber) {
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Reason: "Invalid serial number"})
		return false
	}
	if sensor.Description == "" {
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Reason: "Description is required"})
		return false
	}
	return true
}

func setAllowHeader(c *gin.Context, methods string) {
	c.Header("Allow", methods)
	c.Status(http.StatusNoContent)
}

func setContentLength(c *gin.Context, data interface{}) {
	jsonData, _ := json.Marshal(data)
	c.Header("Content-Length", strconv.Itoa(len(jsonData)))
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func isAcceptableType(accept string) bool {
	if accept == "" {
		return true
	}
	return strings.Contains(accept, "*/*") || strings.Contains(accept, "application/json")
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrSensorNotFound) || errors.Is(err, usecase.ErrUserNotFound) || strings.Contains(err.Error(), "not found"):
		c.JSON(http.StatusNotFound, ErrorResponse{Reason: err.Error()})
	case errors.Is(err, usecase.ErrWrongSensorSerialNumber) ||
		errors.Is(err, usecase.ErrWrongSensorType) ||
		errors.Is(err, usecase.ErrInvalidUserName) ||
		errors.Is(err, usecase.ErrInvalidEventTimestamp):

		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Reason: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{Reason: "Internal server error"})
	}
}

func handleStatusOnlyError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrSensorNotFound) || errors.Is(err, usecase.ErrUserNotFound) || strings.Contains(err.Error(), "not found"):
		c.Status(http.StatusNotFound)
	case errors.Is(err, usecase.ErrWrongSensorSerialNumber) ||
		errors.Is(err, usecase.ErrWrongSensorType) ||
		errors.Is(err, usecase.ErrInvalidUserName) ||
		errors.Is(err, usecase.ErrInvalidEventTimestamp):
		c.Status(http.StatusUnprocessableEntity)
	default:
		c.Status(http.StatusInternalServerError)
	}
}
