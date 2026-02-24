package topology

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"terraform-provider-hashicups/internal/client"
)

const topoBasePath = "/topology"

// Service provides typed operations for topologies.
type Service struct {
	c client.HTTPClient
}

func NewService(c client.HTTPClient) *Service {
	return &Service{c: c}
}

// List returns topologies matching filters.
func (s *Service) List(ctx context.Context, f Filters) (*ListResponse, error) {
	q := url.Values{}
	if f.ID != "" {
		q.Set("id", f.ID)
	}
	if f.Name != "" {
		q.Set("name", f.Name)
	}
	if f.Ecosystem != "" {
		q.Set("ecosystem", f.Ecosystem)
	}
	if f.Mock != nil {
		q.Set("mock", strconv.FormatBool(*f.Mock))
	}
	if f.State != "" {
		q.Set("state", f.State)
	}
	if f.Owner != "" {
		q.Set("owner", f.Owner)
	}
	if f.Env != "" {
		q.Set("env", f.Env)
	}
	if f.CreatedAt != "" {
		q.Set("created_at", f.CreatedAt)
	}
	if f.UpdatedAt != "" {
		q.Set("updated_at", f.UpdatedAt)
	}
	if f.Page > 0 {
		q.Set("page", strconv.Itoa(f.Page))
	}
	if f.Size > 0 {
		q.Set("size", strconv.Itoa(f.Size))
	}

	path := topoBasePath
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}

	var resp ListResponse
	if err := s.c.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("listing topologies: %w", err)
	}
	return &resp, nil
}

// Get returns a topology by ID, optionally including nodes.
func (s *Service) Get(ctx context.Context, id string, includeNodes bool) (*Topology, error) {
	path := fmt.Sprintf("%s/%s", topoBasePath, id)
	if includeNodes {
		path += "?nodes=true"
	}
	var topo Topology
	if err := s.c.Get(ctx, path, &topo); err != nil {
		return nil, fmt.Errorf("getting topology %s: %w", id, err)
	}
	return &topo, nil
}
