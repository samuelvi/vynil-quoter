package provider

import (
	"context"
	"vinylquoter/internal/catalog"
)

type Recognizer interface {
	Identify(ctx context.Context, imagePath string) (catalog.Identification, error)
}
