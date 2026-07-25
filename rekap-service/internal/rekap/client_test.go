package rekap

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientDaftarMahasiswa(t *testing.T) {
	t.Run("respon valid", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(akademikResponse{
				Sukses: true,
				Data: json.RawMessage(`[{"nim":"23010001","nama":"Bunga","jurusan":"Teknik Informatika","status":"Aktif"}]`),
			})
		}))
		defer srv.Close()

		client := NewHTTPAkademikClient(srv.URL)
		mhs, err := client.DaftarMahasiswa(context.Background())

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(mhs) != 1 {
			t.Fatalf("expected 1 mahasiswa, got %d", len(mhs))
		}
		if mhs[0].NIM != "23010001" {
			t.Errorf("expected NIM 23010001, got %s", mhs[0].NIM)
		}
	})

	t.Run("server error 500", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		client := NewHTTPAkademikClient(srv.URL)
		_, err := client.DaftarMahasiswa(context.Background())

		if !errors.Is(err, ErrAkademikTidakTersedia) {
			t.Errorf("expected ErrAkademikTidakTersedia, got %v", err)
		}
	})

	t.Run("latensi timeout", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(3 * time.Second)
		}))
		defer srv.Close()

		client := NewHTTPAkademikClient(srv.URL)
		_, err := client.DaftarMahasiswa(context.Background())

		if err == nil {
			t.Fatal("expected timeout error")
		}
		if !strings.Contains(err.Error(), "context deadline exceeded") && !strings.Contains(err.Error(), "Timeout") {
			t.Logf("got error: %v", err)
		}
	})
}

func TestClientRingkasan(t *testing.T) {
	t.Run("respon valid", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(akademikResponse{
				Sukses: true,
				Data: json.RawMessage(`{"nim":"23010001","nama":"Bunga","jurusan":"Teknik Informatika","status":"Aktif","total_sks":5,"ipk":3.60}`),
			})
		}))
		defer srv.Close()

		client := NewHTTPAkademikClient(srv.URL)
		r, err := client.Ringkasan(context.Background(), "23010001")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if r.NIM != "23010001" {
			t.Errorf("expected NIM 23010001, got %s", r.NIM)
		}
		if r.IPK != 3.60 {
			t.Errorf("expected IPK 3.60, got %f", r.IPK)
		}
	})

	t.Run("mahasiswa tidak ada", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		client := NewHTTPAkademikClient(srv.URL)
		_, err := client.Ringkasan(context.Background(), "99999999")

		if !errors.Is(err, ErrMahasiswaTidakAda) {
			t.Errorf("expected ErrMahasiswaTidakAda, got %v", err)
		}
	})

	t.Run("body json rusak", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("{bukan json"))
		}))
		defer srv.Close()

		client := NewHTTPAkademikClient(srv.URL)
		_, err := client.Ringkasan(context.Background(), "23010001")

		if err == nil {
			t.Fatal("expected error for bad JSON")
		}
	})
}
