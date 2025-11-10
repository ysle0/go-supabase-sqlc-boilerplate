package httputil

import (
	"net/http"
	"os"

	"github.com/go-chi/render"
)

func OkNoDataRaw(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("RUN_INTEGRATION_TESTS") == "false" {
		logger.Info("Ok No Data")
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, render.M{
		"status": "ok",
	})
}

func OkNoDataWithMsgRaw(w http.ResponseWriter, r *http.Request, msg string) {
	if os.Getenv("RUN_INTEGRATION_TESTS") == "false" {
		logger.Info("Ok No Data")
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, render.M{
		"status":  "ok",
		"message": msg,
	})
}

func OkWithMsgRaw(w http.ResponseWriter, r *http.Request, msg string, data any) {
	if os.Getenv("RUN_INTEGRATION_TESTS") == "false" {
		logger.Info("Ok", "data", data)
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, render.M{
		"status":  "ok",
		"message": msg,
		"data":    data,
	})
}

func OkRaw(w http.ResponseWriter, r *http.Request, data any) {
	if os.Getenv("RUN_INTEGRATION_TESTS") == "false" {
		logger.Info("Ok", "data", data)
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, render.M{
		"status":  "ok",
		"message": "ok",
		"data":    data,
	})
}

func FailRaw(w http.ResponseWriter, r *http.Request, msg string) {
	if os.Getenv("RUN_INTEGRATION_TESTS") == "false" {
		logger.Error("Fail", "message", msg)
	}
	render.Status(r, http.StatusInternalServerError)
	render.JSON(w, r, render.M{
		"status": "fail",
		"msg":    msg,
	})
}

func ErrWithMsgRaw(w http.ResponseWriter, r *http.Request, err error, msg string) {
	if os.Getenv("RUN_INTEGRATION_TESTS") == "false" {
		logger.Error("Err", "error", err, "message", msg)
	}
	render.Status(r, http.StatusInternalServerError)
	render.JSON(w, r, render.M{
		"status": "fail",
		"msg":    msg,
	})
}

func ErrRaw(w http.ResponseWriter, r *http.Request, err error) {
	if os.Getenv("RUN_INTEGRATION_TESTS") == "false" {
		logger.Error("Err", "error", err)
	}
	render.Status(r, http.StatusInternalServerError)
	render.JSON(w, r, render.M{
		"status":  "fail",
		"message": "internal server error",
	})
}

func OkNoData(c *HttpRequestContext) {
	OkNoDataRaw(c.Writer, c.Request)
}

func OkNoDataWithMsg(c *HttpRequestContext, msg string) {
	OkNoDataWithMsgRaw(c.Writer, c.Request, msg)
}

func OkWithMsg(c *HttpRequestContext, msg string, data any) {
	OkWithMsgRaw(c.Writer, c.Request, msg, data)
}

func Ok(c *HttpRequestContext, data any) {
	OkRaw(c.Writer, c.Request, data)
}

func Fail(c *HttpRequestContext, msg string) {
	FailRaw(c.Writer, c.Request, msg)
}

func ErrWithMsg(c *HttpRequestContext, err error, msg string) {
	ErrWithMsgRaw(c.Writer, c.Request, err, msg)
}

func Err(c *HttpRequestContext, err error) {
	ErrRaw(c.Writer, c.Request, err)
}
