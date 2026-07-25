package siakad

import (
    "context"
    "database/sql"
    "fmt"
    "regexp"
    "sort"
)

type Service interface {
    TambahMahasiswa(ctx context.Context, input InputMahasiswaDTO) error
    DaftarMahasiswa(ctx context.Context, jurusan string, limit, offset int) ([]Mahasiswa, error)
    DetailMahasiswa(ctx context.Context, nim string) (MahasiswaDetailDTO, error)
    UpdateMahasiswa(ctx context.Context, nim string, input InputMahasiswaDTO) error
    HapusMahasiswa(ctx context.Context, nim string) error
    InputNilai(ctx context.Context, nim string, input InputNilaiDTO) error
    Transkrip(ctx context.Context, nim string) (TranskripDTO, error)
    PerJurusan(ctx context.Context) (map[string][]Mahasiswa, error)
    TopIPK(ctx context.Context, n int) ([]MahasiswaDetailDTO, error)
}

type siakadService struct {
    mhsRepo MahasiswaRepository
    nRepo   NilaiRepository
    db      *sql.DB
}

func NewService(db *sql.DB, mhsRepo MahasiswaRepository, nRepo NilaiRepository) Service {
    return &siakadService{
        mhsRepo: mhsRepo,
        nRepo:   nRepo,
        db:      db,
    }
}

var nimRegex = regexp.MustCompile(`^[0-9]{8}$`)

func HitungIPK(daftar []Nilai) float64 {
    var totalMutu, totalSKS float64
    for _, n := range daftar {
        totalMutu += float64(n.SKS) * n.Mutu
        totalSKS += float64(n.SKS)
    }
    if totalSKS == 0 {
        return 0.0
    }
    return totalMutu / totalSKS
}

func (s *siakadService) TambahMahasiswa(ctx context.Context, input InputMahasiswaDTO) error {
    if !nimRegex.MatchString(input.NIM) {
        return ErrNIMTidakValid
    }
    m := Mahasiswa{
        NIM:     input.NIM,
        Nama:    input.Nama,
        Jurusan: input.Jurusan,
        Status:  input.Status,
    }
    return s.mhsRepo.Tambah(ctx, m)
}

func (s *siakadService) DaftarMahasiswa(ctx context.Context, jurusan string, limit, offset int) ([]Mahasiswa, error) {
    all, err := s.mhsRepo.Semua(ctx)
    if err != nil {
        return nil, err
    }

    if jurusan != "" {
        var filtered []Mahasiswa
        for _, m := range all {
            if m.Jurusan == jurusan {
                filtered = append(filtered, m)
            }
        }
        all = filtered
    }

    if offset < 0 {
        offset = 0
    }
    if offset >= len(all) {
        return []Mahasiswa{}, nil
    }

    end := offset + limit
    if limit <= 0 || end > len(all) {
        end = len(all)
    }

    return all[offset:end], nil
}

func (s *siakadService) DetailMahasiswa(ctx context.Context, nim string) (MahasiswaDetailDTO, error) {
    m, err := s.mhsRepo.Cari(ctx, nim)
    if err != nil {
        return MahasiswaDetailDTO{}, err
    }

    nilai, _ := s.nRepo.PerMahasiswa(ctx, nim)
    ipk := HitungIPK(nilai)

    cumlaude := ""
    if ipk >= 3.50 {
        cumlaude = "Cumlaude"
    }

    return MahasiswaDetailDTO{
        NIM:            m.NIM,
        Nama:           m.Nama,
        Jurusan:        m.Jurusan,
        Status:         m.Status,
        IPK:            ipk,
        StatusCumlaude: cumlaude,
    }, nil
}

func (s *siakadService) UpdateMahasiswa(ctx context.Context, nim string, input InputMahasiswaDTO) error {
    _, err := s.mhsRepo.Cari(ctx, nim)
    if err != nil {
        return err
    }
    m := Mahasiswa{
        NIM:     nim,
        Nama:    input.Nama,
        Jurusan: input.Jurusan,
        Status:  input.Status,
    }
    return s.mhsRepo.Update(ctx, m)
}

func (s *siakadService) HapusMahasiswa(ctx context.Context, nim string) error {
    return s.mhsRepo.Hapus(ctx, nim)
}

func (s *siakadService) InputNilai(ctx context.Context, nim string, input InputNilaiDTO) error {
    if input.Mutu < 0.0 || input.Mutu > 4.0 {
        return ErrNilaiTidakValid
    }
    if input.SKS <= 0 {
        return ErrSKSTidakValid
    }

    _, err := s.mhsRepo.Cari(ctx, nim)
    if err != nil {
        return err
    }

    n := Nilai{
        NIM:    nim,
        KodeMK: input.KodeMK,
        NamaMK: input.NamaMK,
        SKS:    input.SKS,
        Mutu:   input.Mutu,
    }
    return s.nRepo.Tambah(ctx, n)
}

func (s *siakadService) Transkrip(ctx context.Context, nim string) (TranskripDTO, error) {
    m, err := s.mhsRepo.Cari(ctx, nim)
    if err != nil {
        return TranskripDTO{}, err
    }

    daftar, err := s.nRepo.PerMahasiswa(ctx, nim)
    if err != nil {
        return TranskripDTO{}, err
    }

    ipk := HitungIPK(daftar)
    totalSKS := 0
    for _, n := range daftar {
        totalSKS += n.SKS
    }

    cumlaude := ""
    if ipk >= 3.50 {
        cumlaude = "Cumlaude"
    }

    return TranskripDTO{
        Mahasiswa: MahasiswaDetailDTO{
            NIM:            m.NIM,
            Nama:           m.Nama,
            Jurusan:        m.Jurusan,
            Status:         m.Status,
            IPK:            ipk,
            StatusCumlaude: cumlaude,
        },
        DaftarNilai:    daftar,
        TotalSKS:       totalSKS,
        IPK:            ipk,
        StatusCumlaude: cumlaude,
    }, nil
}

func (s *siakadService) PerJurusan(ctx context.Context) (map[string][]Mahasiswa, error) {
    semua, err := s.mhsRepo.Semua(ctx)
    if err != nil {
        return nil, err
    }

    result := make(map[string][]Mahasiswa)
    for _, m := range semua {
        result[m.Jurusan] = append(result[m.Jurusan], m)
    }
    return result, nil
}

func (s *siakadService) TopIPK(ctx context.Context, n int) ([]MahasiswaDetailDTO, error) {
    if n <= 0 {
        n = 3
    }

    semua, err := s.mhsRepo.Semua(ctx)
    if err != nil {
        return nil, fmt.Errorf("gagal mengambil daftar mahasiswa: %w", err)
    }

    var result []MahasiswaDetailDTO
    for _, m := range semua {
        nilai, _ := s.nRepo.PerMahasiswa(ctx, m.NIM)
        ipk := HitungIPK(nilai)
        cumlaude := ""
        if ipk >= 3.50 {
            cumlaude = "Cumlaude"
        }
        result = append(result, MahasiswaDetailDTO{
            NIM:            m.NIM,
            Nama:           m.Nama,
            Jurusan:        m.Jurusan,
            Status:         m.Status,
            IPK:            ipk,
            StatusCumlaude: cumlaude,
        })
    }

    sort.Slice(result, func(i, j int) bool {
        return result[i].IPK > result[j].IPK
    })

    if n > len(result) {
        n = len(result)
    }
    return result[:n], nil
}
