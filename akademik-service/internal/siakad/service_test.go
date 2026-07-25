package siakad

import (
    "context"
    "errors"
    "math"
    "testing"
)

func TestHitungIPK(t *testing.T) {
    tests := []struct {
        name   string
        daftar []Nilai
        want   float64
    }{
        {
            name:   "daftar nilai kosong",
            daftar: []Nilai{},
            want:   0.0,
        },
        {
            name: "1 mata kuliah (3 SKS, mutu 4.0)",
            daftar: []Nilai{
                {SKS: 3, Mutu: 4.0},
            },
            want: 4.0,
        },
        {
            name: "2 mata kuliah beda SKS",
            daftar: []Nilai{
                {SKS: 3, Mutu: 4.0},
                {SKS: 2, Mutu: 3.0},
            },
            want: 3.6,
        },
        {
            name: "4 mata kuliah variatif",
            daftar: []Nilai{
                {SKS: 3, Mutu: 4.0},
                {SKS: 3, Mutu: 3.5},
                {SKS: 2, Mutu: 3.0},
                {SKS: 2, Mutu: 2.5},
            },
            want: 3.35,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := HitungIPK(tt.daftar)
            if math.Abs(got-tt.want) >= 1e-9 {
                t.Errorf("HitungIPK() = %v, want %v", got, tt.want)
            }
        })
    }
}

type mockMahasiswaRepo struct {
    data map[string]Mahasiswa
}

func newMockMahasiswaRepo() *mockMahasiswaRepo {
    return &mockMahasiswaRepo{data: make(map[string]Mahasiswa)}
}

func (m *mockMahasiswaRepo) Tambah(ctx context.Context, mhs Mahasiswa) error {
    if _, ok := m.data[mhs.NIM]; ok {
        return ErrMahasiswaSudahAda
    }
    m.data[mhs.NIM] = mhs
    return nil
}

func (m *mockMahasiswaRepo) Cari(ctx context.Context, nim string) (Mahasiswa, error) {
    mhs, ok := m.data[nim]
    if !ok {
        return Mahasiswa{}, ErrMahasiswaTidakAda
    }
    return mhs, nil
}

func (m *mockMahasiswaRepo) Semua(ctx context.Context) ([]Mahasiswa, error) {
    var result []Mahasiswa
    for _, v := range m.data {
        result = append(result, v)
    }
    return result, nil
}

func (m *mockMahasiswaRepo) Update(ctx context.Context, mhs Mahasiswa) error {
    if _, ok := m.data[mhs.NIM]; !ok {
        return ErrMahasiswaTidakAda
    }
    m.data[mhs.NIM] = mhs
    return nil
}

func (m *mockMahasiswaRepo) Hapus(ctx context.Context, nim string) error {
    if _, ok := m.data[nim]; !ok {
        return ErrMahasiswaTidakAda
    }
    delete(m.data, nim)
    return nil
}

type mockNilaiRepo struct {
    data []Nilai
}

func newMockNilaiRepo() *mockNilaiRepo {
    return &mockNilaiRepo{}
}

func (m *mockNilaiRepo) Tambah(ctx context.Context, n Nilai) error {
    m.data = append(m.data, n)
    return nil
}

func (m *mockNilaiRepo) PerMahasiswa(ctx context.Context, nim string) ([]Nilai, error) {
    var result []Nilai
    for _, n := range m.data {
        if n.NIM == nim {
            result = append(result, n)
        }
    }
    return result, nil
}

type mockService struct {
    mhsRepo *mockMahasiswaRepo
    nRepo   *mockNilaiRepo
}

func TestInputNilaiMahasiswaTidakAda(t *testing.T) {
    mhsRepo := newMockMahasiswaRepo()
    nRepo := newMockNilaiRepo()
    svc := NewService(nil, mhsRepo, nRepo)

    err := svc.InputNilai(context.Background(), "12345678", InputNilaiDTO{
        KodeMK: "MK001",
        NamaMK: "Algoritma",
        SKS:    3,
        Mutu:   4.0,
    })

    if !errors.Is(err, ErrMahasiswaTidakAda) {
        t.Errorf("expected ErrMahasiswaTidakAda, got %v", err)
    }
}

func TestInputNilaiMutuTidakValid(t *testing.T) {
    mhsRepo := newMockMahasiswaRepo()
    nRepo := newMockNilaiRepo()
    svc := NewService(nil, mhsRepo, nRepo)

    err := svc.InputNilai(context.Background(), "12345678", InputNilaiDTO{
        KodeMK: "MK001",
        NamaMK: "Algoritma",
        SKS:    3,
        Mutu:   5.0,
    })

    if !errors.Is(err, ErrNilaiTidakValid) {
        t.Errorf("expected ErrNilaiTidakValid, got %v", err)
    }
}

func TestInputNilaiSKSTidakValid(t *testing.T) {
    mhsRepo := newMockMahasiswaRepo()
    nRepo := newMockNilaiRepo()
    svc := NewService(nil, mhsRepo, nRepo)

    err := svc.InputNilai(context.Background(), "12345678", InputNilaiDTO{
        KodeMK: "MK001",
        NamaMK: "Algoritma",
        SKS:    0,
        Mutu:   4.0,
    })

    if !errors.Is(err, ErrSKSTidakValid) {
        t.Errorf("expected ErrSKSTidakValid, got %v", err)
    }
}

func TestTambahMahasiswaNIMTidakValid(t *testing.T) {
    mhsRepo := newMockMahasiswaRepo()
    nRepo := newMockNilaiRepo()
    svc := NewService(nil, mhsRepo, nRepo)

    err := svc.TambahMahasiswa(context.Background(), InputMahasiswaDTO{
        NIM:     "1234",
        Nama:    "Test",
        Jurusan: "TI",
        Status:  "Aktif",
    })

    if !errors.Is(err, ErrNIMTidakValid) {
        t.Errorf("expected ErrNIMTidakValid, got %v", err)
    }
}
