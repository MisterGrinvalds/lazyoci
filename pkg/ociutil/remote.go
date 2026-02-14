package ociutil

import (
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

// NewRemoteRepository creates an oras remote.Repository for the given reference.
// If credFn is non-nil it is used for authentication; otherwise anonymous auth is used.
func NewRemoteRepository(ref *Reference, insecure bool, credFn auth.CredentialFunc) (*remote.Repository, error) {
	// Build the full repository reference
	repoRef := ref.Registry + "/" + ref.Repository

	repo, err := remote.NewRepository(repoRef)
	if err != nil {
		return nil, err
	}

	// Handle docker.io remapping
	if ref.Registry == "docker.io" {
		repo.Reference.Registry = "registry-1.docker.io"
	}

	// Enable plain HTTP for insecure registries
	repo.PlainHTTP = insecure

	// Use provided credentials or fall back to anonymous auth
	if credFn == nil {
		credFn = auth.StaticCredential(repo.Reference.Registry, auth.Credential{})
	}

	repo.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Credential: credFn,
	}

	return repo, nil
}
