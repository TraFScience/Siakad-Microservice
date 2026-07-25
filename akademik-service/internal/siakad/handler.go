package siakad

import (
    "errors"
    "net/http"
    "strconv"

    "akademik-service/internal/response"

    "github.com/gin-gonic/gin"
)

type Handler struct {
    svc Service
}

func NewHandler(svc Service) *Handler {
    return &Handler{svc: svc}
}

func (h *Handler) Register(r *gin.RouterGroup) {
    r.POST("/mahasiswa", h.TambahMahasiswa)
    r.GET("/mahasiswa", h.DaftarMahasiswa)
    r.GET("/mahasiswa/:nim", h.DetailMahasiswa)
    r.PUT("/mahasiswa/:nim", h.UpdateMahasiswa)
    r.DELETE("/mahasiswa/:nim", h.HapusMahasiswa)
    r.POST("/mahasiswa/:nim/nilai", h.InputNilai)
	r.GET("/mahasiswa/:nim/transkrip", h.Transkrip)
	r.GET("/mahasiswa/:nim/ringkasan", h.Ringkasan)
}

func handleError(c *gin.Context, err error) {
    status := MapErrorToHTTP(err)
    response.Error(c, status, err)
}

func (h *Handler) TambahMahasiswa(c *gin.Context) {
    var input InputMahasiswaDTO
    if err := c.ShouldBindJSON(&input); err != nil {
        response.Error(c, http.StatusBadRequest, err)
        return
    }
    if err := h.svc.TambahMahasiswa(c.Request.Context(), input); err != nil {
        handleError(c, err)
        return
    }
    response.Success(c, http.StatusCreated, input)
}

func (h *Handler) DaftarMahasiswa(c *gin.Context) {
    jurusan := c.Query("jurusan")
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0"))
    offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

    data, err := h.svc.DaftarMahasiswa(c.Request.Context(), jurusan, limit, offset)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, err)
        return
    }
    response.Success(c, http.StatusOK, data)
}

func (h *Handler) DetailMahasiswa(c *gin.Context) {
    nim := c.Param("nim")
    data, err := h.svc.DetailMahasiswa(c.Request.Context(), nim)
    if err != nil {
        handleError(c, err)
        return
    }
    response.Success(c, http.StatusOK, data)
}

func (h *Handler) UpdateMahasiswa(c *gin.Context) {
    nim := c.Param("nim")
    var input InputMahasiswaDTO
    if err := c.ShouldBindJSON(&input); err != nil {
        response.Error(c, http.StatusBadRequest, err)
        return
    }
    if err := h.svc.UpdateMahasiswa(c.Request.Context(), nim, input); err != nil {
        handleError(c, err)
        return
    }
    response.Success(c, http.StatusOK, gin.H{"nim": nim, "pesan": "data berhasil diperbarui"})
}

func (h *Handler) HapusMahasiswa(c *gin.Context) {
    nim := c.Param("nim")
    if err := h.svc.HapusMahasiswa(c.Request.Context(), nim); err != nil {
        handleError(c, err)
        return
    }
    response.Success(c, http.StatusOK, gin.H{"nim": nim, "pesan": "mahasiswa berhasil dihapus"})
}

func (h *Handler) InputNilai(c *gin.Context) {
    nim := c.Param("nim")
    var input InputNilaiDTO
    if err := c.ShouldBindJSON(&input); err != nil {
        response.Error(c, http.StatusBadRequest, err)
        return
    }
    if err := h.svc.InputNilai(c.Request.Context(), nim, input); err != nil {
        handleError(c, err)
        return
    }
    response.Success(c, http.StatusCreated, input)
}

func (h *Handler) Transkrip(c *gin.Context) {
    nim := c.Param("nim")
    data, err := h.svc.Transkrip(c.Request.Context(), nim)
    if err != nil {
        if errors.Is(err, ErrMahasiswaTidakAda) {
            response.Error(c, http.StatusNotFound, err)
            return
        }
        response.Error(c, http.StatusInternalServerError, err)
        return
    }
    response.Success(c, http.StatusOK, data)
}

func (h *Handler) Ringkasan(c *gin.Context) {
    nim := c.Param("nim")
    data, err := h.svc.Ringkasan(c.Request.Context(), nim)
    if err != nil {
        handleError(c, err)
        return
    }
    response.Success(c, http.StatusOK, data)
}
