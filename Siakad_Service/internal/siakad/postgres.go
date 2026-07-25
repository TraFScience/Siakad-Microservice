package siakad

import (
    "context"
    "database/sql"
    "errors"
    "fmt"

    "github.com/jackc/pgx/v5/pgconn"
)

type mahasiswaRepo struct {
    db *sql.DB
}

type nilaiRepo struct {
    db *sql.DB
}

func NewMahasiswaRepository(db *sql.DB) MahasiswaRepository {
    return &mahasiswaRepo{db: db}
}

func NewNilaiRepository(db *sql.DB) NilaiRepository {
    return &nilaiRepo{db: db}
}

func (r *mahasiswaRepo) Tambah(ctx context.Context, m Mahasiswa) error {
    _, err := r.db.ExecContext(ctx,
        `INSERT INTO mahasiswa (nim, nama, jurusan, status) VALUES ($1, $2, $3, $4)`,
        m.NIM, m.Nama, m.Jurusan, m.Status,
    )
    if err != nil {
        var pgerr *pgconn.PgError
        if errors.As(err, &pgerr) && pgerr.Code == "23505" {
            return fmt.Errorf("%w: %v", ErrMahasiswaSudahAda, err)
        }
        return err
    }
    return nil
}

func (r *mahasiswaRepo) Cari(ctx context.Context, nim string) (Mahasiswa, error) {
    var m Mahasiswa
    err := r.db.QueryRowContext(ctx,
        `SELECT nim, nama, jurusan, status FROM mahasiswa WHERE nim = $1`, nim,
    ).Scan(&m.NIM, &m.Nama, &m.Jurusan, &m.Status)
    if errors.Is(err, sql.ErrNoRows) {
        return m, fmt.Errorf("%w: %v", ErrMahasiswaTidakAda, err)
    }
    if err != nil {
        return m, err
    }
    return m, nil
}

func (r *mahasiswaRepo) Semua(ctx context.Context) ([]Mahasiswa, error) {
    rows, err := r.db.QueryContext(ctx,
        `SELECT nim, nama, jurusan, status FROM mahasiswa ORDER BY nim`,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var result []Mahasiswa
    for rows.Next() {
        var m Mahasiswa
        if err := rows.Scan(&m.NIM, &m.Nama, &m.Jurusan, &m.Status); err != nil {
            return nil, err
        }
        result = append(result, m)
    }
    return result, rows.Err()
}

func (r *mahasiswaRepo) Update(ctx context.Context, m Mahasiswa) error {
    res, err := r.db.ExecContext(ctx,
        `UPDATE mahasiswa SET nama = $1, jurusan = $2, status = $3 WHERE nim = $4`,
        m.Nama, m.Jurusan, m.Status, m.NIM,
    )
    if err != nil {
        return err
    }
    if n, _ := res.RowsAffected(); n == 0 {
        return fmt.Errorf("%w", ErrMahasiswaTidakAda)
    }
    return nil
}

func (r *mahasiswaRepo) Hapus(ctx context.Context, nim string) error {
    res, err := r.db.ExecContext(ctx,
        `DELETE FROM mahasiswa WHERE nim = $1`, nim,
    )
    if err != nil {
        return err
    }
    if n, _ := res.RowsAffected(); n == 0 {
        return fmt.Errorf("%w", ErrMahasiswaTidakAda)
    }
    return nil
}

func (r *nilaiRepo) Tambah(ctx context.Context, n Nilai) error {
    _, err := r.db.ExecContext(ctx,
        `INSERT INTO nilai (nim, kode_mk, nama_mk, sks, mutu) VALUES ($1, $2, $3, $4, $5)`,
        n.NIM, n.KodeMK, n.NamaMK, n.SKS, n.Mutu,
    )
    return err
}

func (r *nilaiRepo) PerMahasiswa(ctx context.Context, nim string) ([]Nilai, error) {
    rows, err := r.db.QueryContext(ctx,
        `SELECT id, nim, kode_mk, nama_mk, sks, mutu FROM nilai WHERE nim = $1 ORDER BY id`, nim,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var result []Nilai
    for rows.Next() {
        var n Nilai
        if err := rows.Scan(&n.ID, &n.NIM, &n.KodeMK, &n.NamaMK, &n.SKS, &n.Mutu); err != nil {
            return nil, err
        }
        result = append(result, n)
    }
    return result, rows.Err()
}
