package rekap

import (
	"errors"
	"net/http"
)

var (
	ErrNIMTidakValid         = errors.New("NIM bukan 8 digit angka")
	ErrMahasiswaTidakAda     = errors.New("mahasiswa tidak ditemukan")
	ErrAkademikTimeout       = errors.New("waktu panggilan ke akademik service habis")
	ErrAkademikTidakTersedia = errors.New("akademik service tidak tersedia")
)

func MapErrorToHTTP(err error) int {
	switch {
	case errors.Is(err, ErrNIMTidakValid):
		return http.StatusBadRequest
	case errors.Is(err, ErrMahasiswaTidakAda):
		return http.StatusNotFound
	case errors.Is(err, ErrAkademikTimeout):
		return http.StatusGatewayTimeout
	case errors.Is(err, ErrAkademikTidakTersedia):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
