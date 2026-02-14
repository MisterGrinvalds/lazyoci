package build

import (
	"context"
	"fmt"

	"github.com/mistergrinvalds/lazyoci/pkg/ociutil"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// pushToTarget pushes an OCI layout to a single target registry with a given tag.
func (b *Builder) pushToTarget(ctx context.Context, ociLayoutPath, tag string, target Target) (*TargetResult, error) {
	ref := target.Registry + ":" + tag
	b.logf("  Pushing %s...\n", ref)

	// Open the local OCI layout as a source store
	store, err := oci.New(ociLayoutPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open OCI layout: %w", err)
	}

	// Parse the target as a reference
	parsed, err := ociutil.ParseReference(ref)
	if err != nil {
		return nil, fmt.Errorf("invalid target reference %q: %w", ref, err)
	}

	// Resolve credentials for this registry
	var credFn auth.CredentialFunc
	if b.opts.CredentialFunc != nil {
		credFn = b.opts.CredentialFunc(parsed.Registry)
	}

	// Create remote repository
	remoteRepo, err := ociutil.NewRemoteRepository(parsed, b.opts.Insecure, credFn)
	if err != nil {
		return nil, fmt.Errorf("failed to create remote repository: %w", err)
	}

	// We need to determine the source tag in the local OCI store.
	// The local store was created by the build handlers and has a single tag.
	// We'll use the oras.Copy with the source tag.
	//
	// The local OCI store stores content by tag. During build, we tag it with
	// a canonical name. We need to find what tag exists in the local store.
	// Use the index.json to find the manifest.
	index, err := ociutil.ReadOCIIndex(ociLayoutPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read OCI index: %w", err)
	}
	if len(index.Manifests) == 0 {
		return nil, fmt.Errorf("OCI layout has no manifests")
	}

	// Copy from local store to remote, using the first manifest's digest as source
	srcDigest := index.Manifests[0].Digest
	desc, err := oras.Copy(ctx, store, srcDigest, remoteRepo, tag, oras.CopyOptions{})
	if err != nil {
		return nil, fmt.Errorf("push failed: %w", err)
	}

	b.logf("  Pushed %s (digest: %s)\n", ref, desc.Digest.String())

	return &TargetResult{
		Reference: ref,
		Digest:    desc.Digest.String(),
		Pushed:    true,
	}, nil
}
