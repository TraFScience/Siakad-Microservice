package rekap

type Mahasiswa struct {
	NIM     string `json:"nim"`
	Nama    string `json:"nama"`
	Jurusan string `json:"jurusan"`
	Status  string `json:"status"`
}

type Ringkasan struct {
	NIM      string  `json:"nim"`
	Nama     string  `json:"nama"`
	Jurusan  string  `json:"jurusan"`
	Status   string  `json:"status"`
	TotalSKS int     `json:"total_sks"`
	IPK      float64 `json:"ipk"`
}
