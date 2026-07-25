package rekap

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"rekap-service/internal/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("/rekap/jurusan", h.PerJurusan)
	r.GET("/rekap/top-ipk", h.TopIPK)
	r.GET("/rekap/mahasiswa/:nim", h.RingkasanMahasiswa)
}

func handleError(c *gin.Context, err error) {
	status := MapErrorToHTTP(err)

	if errors.Is(err, ErrAkademikTimeout) {
		err = ErrAkademikTimeout
	}
	if errors.Is(err, ErrAkademikTidakTersedia) {
		err = ErrAkademikTidakTersedia
	}

	response.Error(c, status, err)
}

func (h *Handler) PerJurusan(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	data, err := h.svc.PerJurusan(ctx)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, data)
}

func (h *Handler) TopIPK(c *gin.Context) {
	nStr := c.DefaultQuery("n", "3")
	n, err := strconv.Atoi(nStr)
	if err != nil || n <= 0 {
		response.Error(c, http.StatusBadRequest, errors.New("parameter n harus bilangan bulat positif"))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	data, err := h.svc.TopIPK(ctx, n)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, data)
}

func (h *Handler) RingkasanMahasiswa(c *gin.Context) {
	nim := c.Param("nim")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	data, err := h.svc.RingkasanMahasiswa(ctx, nim)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, data)
}
