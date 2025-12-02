package osrm

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/ride4Low/contracts/types"
	"github.com/ride4Low/trip-service/internal/domain"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
	}
}

func (c *Client) GetRoute(ctx context.Context, pickup, dropoff types.Coordinate) (*domain.OsrmApiResponse, error) {
	url := fmt.Sprintf(
		"%s/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		c.baseURL,
		pickup.Longitude, pickup.Latitude,
		dropoff.Longitude, dropoff.Latitude,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OSRM request failed with status code: %d", resp.StatusCode)
	}

	var osrmResponse domain.OsrmApiResponse
	if err := sonic.Unmarshal(body, &osrmResponse); err != nil {
		return nil, err
	}

	return &osrmResponse, nil
}
