package rekap

import (
	"context"
	"regexp"
	"sort"
)

type Service interface {
	PerJurusan(ctx context.Context) (map[string][]Ringkasan, error)
	TopIPK(ctx context.Context, n int) ([]Ringkasan, error)
	RingkasanMahasiswa(ctx context.Context, nim string) (Ringkasan, error)
}

type rekapService struct {
	client AkademikClient
}

func NewService(client AkademikClient) Service {
	return &rekapService{client: client}
}

var nimRegex = regexp.MustCompile(`^[0-9]{8}$`)

func (s *rekapService) PerJurusan(ctx context.Context) (map[string][]Ringkasan, error) {
	mahasiswa, err := s.client.DaftarMahasiswa(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]Ringkasan)
	for _, m := range mahasiswa {
		r, err := s.client.Ringkasan(ctx, m.NIM)
		if err != nil {
			return nil, err
		}
		result[m.Jurusan] = append(result[m.Jurusan], r)
	}
	return result, nil
}

func (s *rekapService) TopIPK(ctx context.Context, n int) ([]Ringkasan, error) {
	mahasiswa, err := s.client.DaftarMahasiswa(ctx)
	if err != nil {
		return nil, err
	}

	var result []Ringkasan
	for _, m := range mahasiswa {
		r, err := s.client.Ringkasan(ctx, m.NIM)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].IPK > result[j].IPK
	})

	if n > len(result) {
		n = len(result)
	}
	return result[:n], nil
}

func (s *rekapService) RingkasanMahasiswa(ctx context.Context, nim string) (Ringkasan, error) {
	if !nimRegex.MatchString(nim) {
		return Ringkasan{}, ErrNIMTidakValid
	}
	return s.client.Ringkasan(ctx, nim)
}
