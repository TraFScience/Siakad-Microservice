package rekap

import (
	"context"
	"errors"
	"testing"
)

type mockAkademikClient struct {
	mahasiswa []Mahasiswa
	ringkasan map[string]Ringkasan
	err       error
}

func (m *mockAkademikClient) DaftarMahasiswa(ctx context.Context) ([]Mahasiswa, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.mahasiswa, nil
}

func (m *mockAkademikClient) Ringkasan(ctx context.Context, nim string) (Ringkasan, error) {
	if m.err != nil {
		return Ringkasan{}, m.err
	}
	r, ok := m.ringkasan[nim]
	if !ok {
		return Ringkasan{}, ErrMahasiswaTidakAda
	}
	return r, nil
}

func TestTopIPK(t *testing.T) {
	t.Run("top 2 dari 4 mahasiswa", func(t *testing.T) {
		mock := &mockAkademikClient{
			mahasiswa: []Mahasiswa{
				{NIM: "23010001", Nama: "A", Jurusan: "TI"},
				{NIM: "23010002", Nama: "B", Jurusan: "SI"},
				{NIM: "23010003", Nama: "C", Jurusan: "TI"},
				{NIM: "23010004", Nama: "D", Jurusan: "SI"},
			},
			ringkasan: map[string]Ringkasan{
				"23010001": {NIM: "23010001", Nama: "A", Jurusan: "TI", IPK: 3.5},
				"23010002": {NIM: "23010002", Nama: "B", Jurusan: "SI", IPK: 3.8},
				"23010003": {NIM: "23010003", Nama: "C", Jurusan: "TI", IPK: 3.2},
				"23010004": {NIM: "23010004", Nama: "D", Jurusan: "SI", IPK: 4.0},
			},
		}

		svc := NewService(mock)
		result, err := svc.TopIPK(context.Background(), 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2, got %d", len(result))
		}
		if result[0].NIM != "23010004" || result[0].IPK != 4.0 {
			t.Errorf("expected 23010004 IPK 4.0 first, got %s IPK %f", result[0].NIM, result[0].IPK)
		}
		if result[1].NIM != "23010002" || result[1].IPK != 3.8 {
			t.Errorf("expected 23010002 IPK 3.8 second, got %s IPK %f", result[1].NIM, result[1].IPK)
		}
	})

	t.Run("top 10 dari 4 mahasiswa", func(t *testing.T) {
		mock := &mockAkademikClient{
			mahasiswa: []Mahasiswa{
				{NIM: "23010001", Nama: "A", Jurusan: "TI"},
				{NIM: "23010002", Nama: "B", Jurusan: "SI"},
				{NIM: "23010003", Nama: "C", Jurusan: "TI"},
				{NIM: "23010004", Nama: "D", Jurusan: "SI"},
			},
			ringkasan: map[string]Ringkasan{
				"23010001": {NIM: "23010001", Nama: "A", Jurusan: "TI", IPK: 3.5},
				"23010002": {NIM: "23010002", Nama: "B", Jurusan: "SI", IPK: 3.8},
				"23010003": {NIM: "23010003", Nama: "C", Jurusan: "TI", IPK: 3.2},
				"23010004": {NIM: "23010004", Nama: "D", Jurusan: "SI", IPK: 4.0},
			},
		}

		svc := NewService(mock)
		result, err := svc.TopIPK(context.Background(), 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 4 {
			t.Fatalf("expected 4, got %d", len(result))
		}
	})
}

func TestPerJurusan(t *testing.T) {
	t.Run("3 mahasiswa di 2 jurusan", func(t *testing.T) {
		mock := &mockAkademikClient{
			mahasiswa: []Mahasiswa{
				{NIM: "23010001", Nama: "A", Jurusan: "TI"},
				{NIM: "23010002", Nama: "B", Jurusan: "SI"},
				{NIM: "23010003", Nama: "C", Jurusan: "TI"},
			},
			ringkasan: map[string]Ringkasan{
				"23010001": {NIM: "23010001", Nama: "A", Jurusan: "TI", IPK: 3.5},
				"23010002": {NIM: "23010002", Nama: "B", Jurusan: "SI", IPK: 3.8},
				"23010003": {NIM: "23010003", Nama: "C", Jurusan: "TI", IPK: 3.2},
			},
		}

		svc := NewService(mock)
		result, err := svc.PerJurusan(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2 jurusan, got %d", len(result))
		}
		if len(result["TI"]) != 2 {
			t.Errorf("expected 2 mahasiswa di TI, got %d", len(result["TI"]))
		}
		if len(result["SI"]) != 1 {
			t.Errorf("expected 1 mahasiswa di SI, got %d", len(result["SI"]))
		}
	})

	t.Run("client mengembalikan error", func(t *testing.T) {
		mock := &mockAkademikClient{
			err: ErrAkademikTidakTersedia,
		}

		svc := NewService(mock)
		_, err := svc.PerJurusan(context.Background())
		if !errors.Is(err, ErrAkademikTidakTersedia) {
			t.Errorf("expected ErrAkademikTidakTersedia, got %v", err)
		}
	})
}

func TestRingkasanMahasiswaNIMTidakValid(t *testing.T) {
	mock := &mockAkademikClient{}
	svc := NewService(mock)

	_, err := svc.RingkasanMahasiswa(context.Background(), "1234")
	if !errors.Is(err, ErrNIMTidakValid) {
		t.Errorf("expected ErrNIMTidakValid, got %v", err)
	}
}
