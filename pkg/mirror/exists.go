package mirror

import (
	"context"

	"github.com/mistergrinvalds/lazyoci/pkg/ociutil"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// Exists checks whether an OCI artifact (chart or image) already exists in a
// remote registry by attempting to resolve its manifest.  Returns true when
// the manifest is found, false on any error (including auth failures and
// network problems).
func Exists(ctx context.Context, ref string, insecure bool, credFn auth.CredentialFunc) bool {
	parsed, err := ociutil.ParseReference(ref)
	if err != nil {
		return false
	}

	repo, err := ociutil.NewRemoteRepository(parsed, insecure, credFn)
	if err != nil {
		return false
	}

	_, err = repo.Resolve(ctx, parsed.Tag)
	return err == nil
}
