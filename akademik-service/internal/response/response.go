package response

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

type envelope struct {
    Sukses bool   `json:"sukses"`
    Data   any    `json:"data,omitempty"`
    Error  string `json:"error,omitempty"`
}

func Success(c *gin.Context, status int, data any) {
    c.JSON(status, envelope{Sukses: true, Data: data})
}

func Error(c *gin.Context, status int, err error) {
    msg := http.StatusText(status)
    if err != nil {
        msg = err.Error()
    }
    c.AbortWithStatusJSON(status, envelope{Sukses: false, Error: msg})
}
