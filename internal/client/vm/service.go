package vm

import (
	"context"
	"fmt"

	"terraform-provider-hashicups/internal/client"
)

const vmBasePath = "/vms"

// Service provides typed VM operations on top of the generic HTTP client.
type Service struct {
	c client.HTTPClient
}

func NewService(c client.HTTPClient) *Service {
	return &Service{c: c}
}

func (s *Service) Create(ctx context.Context, payload VMPayload) (*VM, error) {
	var vm VM
	if err := s.c.Post(ctx, vmBasePath, payload, &vm); err != nil {
		return nil, err
	}
	return &vm, nil
}

func (s *Service) Get(ctx context.Context, id string) (*VM, error) {
	var vm VM
	if err := s.c.Get(ctx, fmt.Sprintf("%s/%s", vmBasePath, id), &vm); err != nil {
		return nil, err
	}
	return &vm, nil
}

func (s *Service) Update(ctx context.Context, id string, payload VMPayload) (*VM, error) {
	if err := s.c.Put(ctx, fmt.Sprintf("%s/%s", vmBasePath, id), payload, nil); err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.c.Delete(ctx, fmt.Sprintf("%s/%s", vmBasePath, id))
}
