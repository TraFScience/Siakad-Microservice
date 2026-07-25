package siakad

import (
    "errors"
    "net/http"
)

var (
	ErrNIMTidakValid     = errors.New("NIM bukan 8 digit angka")
	ErrMahasiswaSudahAda = errors.New("NIM sudah terdaftar")
	ErrMahasiswaTidakAda = errors.New("mahasiswa tidak ditemukan")
	ErrNilaiTidakValid   = errors.New("mutu di luar rentang 0.0-4.0")
	ErrSKSTidakValid     = errors.New("SKS harus lebih besar dari 0")
)

func MapErrorToHTTP(err error) int {
    switch {
    case errors.Is(err, ErrNIMTidakValid),
        errors.Is(err, ErrNilaiTidakValid),
        errors.Is(err, ErrSKSTidakValid):
        return http.StatusBadRequest
    case errors.Is(err, ErrMahasiswaTidakAda):
        return http.StatusNotFound
    case errors.Is(err, ErrMahasiswaSudahAda):
        return http.StatusConflict
    default:
        return http.StatusInternalServerError
    }
}