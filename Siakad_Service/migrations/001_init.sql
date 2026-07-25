CREATE TABLE IF NOT EXISTS mahasiswa (
    nim     VARCHAR(8) PRIMARY KEY,
    nama    VARCHAR NOT NULL,
    jurusan VARCHAR NOT NULL,
    status  VARCHAR NOT NULL,
    CONSTRAINT chk_nim CHECK (nim ~ '^[0-9]{8}$'),
    CONSTRAINT chk_status CHECK (status IN ('Aktif', 'Cuti', 'Lulus'))
);

CREATE TABLE IF NOT EXISTS nilai (
    id      BIGSERIAL PRIMARY KEY,
    nim     VARCHAR(8) NOT NULL REFERENCES mahasiswa(nim) ON DELETE CASCADE,
    kode_mk VARCHAR NOT NULL,
    nama_mk VARCHAR NOT NULL,
    sks     INTEGER NOT NULL,
    mutu    NUMERIC(3,2) NOT NULL,
    CONSTRAINT chk_sks CHECK (sks > 0),
    CONSTRAINT chk_mutu CHECK (mutu >= 0.0 AND mutu <= 4.0)
);

