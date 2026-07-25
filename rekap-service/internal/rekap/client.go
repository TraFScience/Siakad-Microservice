package rekap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AkademikClient interface {
	DaftarMahasiswa(ctx context.Context) ([]Mahasiswa, error)
	Ringkasan(ctx context.Context, nim string) (Ringkasan, error)
}

type akademikResponse struct {
	Sukses bool            `json:"sukses"`
	Data   json.RawMessage `json:"data"`
	Error  string          `json:"error"`
}

type HTTPAkademikClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPAkademikClient(baseURL string) *HTTPAkademikClient {
	return &HTTPAkademikClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

func (c *HTTPAkademikClient) DaftarMahasiswa(ctx context.Context) ([]Mahasiswa, error) {
	url := c.baseURL + "/api/v1/mahasiswa"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request: %w", err)
	}

	if requestID, ok := ctx.Value("request_id").(string); ok {
		req.Header.Set("X-Request-ID", requestID)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAkademikTidakTersedia, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAkademikTidakTersedia, err)
	}

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("%w: status %d", ErrAkademikTidakTersedia, resp.StatusCode)
	}

	var ar akademikResponse
	if err := json.Unmarshal(body, &ar); err != nil {
		return nil, fmt.Errorf("gagal membaca respon: %w", err)
	}

	if !ar.Sukses {
		return nil, fmt.Errorf("%w: %s", ErrAkademikTidakTersedia, ar.Error)
	}

	var mahasiswa []Mahasiswa
	if err := json.Unmarshal(ar.Data, &mahasiswa); err != nil {
		return nil, fmt.Errorf("gagal membaca data mahasiswa: %w", err)
	}

	return mahasiswa, nil
}

func (c *HTTPAkademikClient) Ringkasan(ctx context.Context, nim string) (Ringkasan, error) {
	url := fmt.Sprintf("%s/api/v1/mahasiswa/%s/ringkasan", c.baseURL, nim)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Ringkasan{}, fmt.Errorf("gagal membuat request: %w", err)
	}

	if requestID, ok := ctx.Value("request_id").(string); ok {
		req.Header.Set("X-Request-ID", requestID)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return Ringkasan{}, fmt.Errorf("%w: %v", ErrAkademikTidakTersedia, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Ringkasan{}, fmt.Errorf("%w: %v", ErrAkademikTidakTersedia, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return Ringkasan{}, ErrMahasiswaTidakAda
	}

	if resp.StatusCode >= 500 {
		return Ringkasan{}, fmt.Errorf("%w: status %d", ErrAkademikTidakTersedia, resp.StatusCode)
	}

	var ar akademikResponse
	if err := json.Unmarshal(body, &ar); err != nil {
		return Ringkasan{}, fmt.Errorf("gagal membaca respon: %w", err)
	}

	if !ar.Sukses {
		return Ringkasan{}, fmt.Errorf("%w: %s", ErrAkademikTidakTersedia, ar.Error)
	}

	var r Ringkasan
	if err := json.Unmarshal(ar.Data, &r); err != nil {
		return Ringkasan{}, fmt.Errorf("gagal membaca data ringkasan: %w", err)
	}

	return r, nil
}
