package siakad

import "context"

type MahasiswaRepository interface {
    Tambah(ctx context.Context, m Mahasiswa) error
    Cari(ctx context.Context, nim string) (Mahasiswa, error)
    Semua(ctx context.Context) ([]Mahasiswa, error)
    Update(ctx context.Context, m Mahasiswa) error
    Hapus(ctx context.Context, nim string) error
}

type NilaiRepository interface {
    Tambah(ctx context.Context, n Nilai) error
    PerMahasiswa(ctx context.Context, nim string) ([]Nilai, error)
}
