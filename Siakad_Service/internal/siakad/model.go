package siakad

type Mahasiswa struct {
	NIM 	string `json:"nim"`
	Nama 	string `json:"nama"`
	Jurusan string `json:"jurusan"`
	Status 	string `json:"status"`
}

type Nilai struct {
    ID     int64   `json:"id"`
    NIM    string  `json:"nim"`
    KodeMK string  `json:"kode_mk"`
    NamaMK string  `json:"nama_mk"`
    SKS    int     `json:"sks"`
    Mutu   float64 `json:"mutu"`
} 

type MahasiswaDetailDTO struct {
    NIM            string  `json:"nim"`
    Nama           string  `json:"nama"`
    Jurusan        string  `json:"jurusan"`
    Status         string  `json:"status"`
    IPK            float64 `json:"ipk"`
    StatusCumlaude string  `json:"status_cumlaude"`
}

type TranskripDTO struct {
    Mahasiswa      MahasiswaDetailDTO `json:"mahasiswa"`
    DaftarNilai    []Nilai            `json:"daftar_nilai"`
    TotalSKS       int                `json:"total_sks"`
    IPK            float64            `json:"ipk"`
    StatusCumlaude string             `json:"status_cumlaude"`
}

type InputMahasiswaDTO struct {
    NIM     string `json:"nim" binding:"required"`
    Nama    string `json:"nama" binding:"required"`
    Jurusan string `json:"jurusan" binding:"required"`
    Status  string `json:"status" binding:"required"`
}

type InputNilaiDTO struct {
    KodeMK string  `json:"kode_mk" binding:"required"`
    NamaMK string  `json:"nama_mk" binding:"required"`
    SKS    int     `json:"sks" binding:"required"`
    Mutu   float64 `json:"mutu" binding:"required"`
}
