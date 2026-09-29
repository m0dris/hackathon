package core_http_response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	core_errors "github.com/m0dris/hackathon/internal/core/errors"
	core_logger "github.com/m0dris/hackathon/internal/core/logger"
	"go.uber.org/zap"
)

type Responder struct {
	log *core_logger.Logger
	rw  http.ResponseWriter
}

func NewResponder(log *core_logger.Logger, rw http.ResponseWriter) *Responder {
	return &Responder{
		log: log,
		rw:  rw,
	}
}

func (h *Responder) JSONResponse(
	responseBody any,
	statusCode int,
) {
	h.rw.WriteHeader(statusCode)

	if err := json.NewEncoder(h.rw).Encode(responseBody); err != nil {
		h.log.Error("write HTTP response", zap.Error(err))
	}
}

func (h *Responder) NoContentResponse() {
	h.rw.WriteHeader(http.StatusNoContent)
}

func (h *Responder) ErrorResponse(err error, msg string) {
	var (
		statusCode int
		logFunc    func(string, ...zap.Field)
	)

	switch {
	case errors.Is(err, core_errors.ErrInvalidArgument):
		statusCode = http.StatusBadRequest
		logFunc = h.log.Warn

	case errors.Is(err, core_errors.ErrNotFound):
		statusCode = http.StatusNotFound
		logFunc = h.log.Debug

	case errors.Is(err, core_errors.ErrConflict):
		statusCode = http.StatusConflict
		logFunc = h.log.Warn

	case errors.Is(err, core_errors.ErrTimeLimit):
		statusCode = http.StatusTooManyRequests
		logFunc = h.log.Debug

	case errors.Is(err, core_errors.ErrTimesUp):
		statusCode = http.StatusBadRequest
		logFunc = h.log.Warn

	case errors.Is(err, core_errors.ErrFrozen):
		statusCode = http.StatusTooManyRequests
		logFunc = h.log.Warn

	default:
		statusCode = http.StatusInternalServerError
		logFunc = h.log.Error
	}

	logFunc(msg, zap.Error(err))

	h.errorResponse(statusCode, err, msg)
}

func (h *Responder) PanicResponse(p any, msg string) {
	statusCode := http.StatusInternalServerError
	err := fmt.Errorf("unexpected panic: %v", p)

	h.log.Error(msg, zap.Error(err))

	h.errorResponse(statusCode, err, msg)
}

func (h *Responder) errorResponse(
	statusCode int,
	err error,
	msg string,
) {
	response := map[string]string{
		"message": msg,
		"error":   err.Error(),
	}

	h.JSONResponse(
		response,
		statusCode,
	)
}
