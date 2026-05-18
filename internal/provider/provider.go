package provider

import (
	"context"
	"vinylquoter/internal/catalog"
)

type RecognitionRequest struct {
	ImagePath       string
	MediaCondition  string
	SleeveCondition string
}

type Recognizer interface {
	Identify(ctx context.Context, request RecognitionRequest) (catalog.Identification, error)
}
