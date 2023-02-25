package db

import (
	"github.com/google/uuid"
)

type Source struct {
	ID         string
	Type       string
	URL        string
	Credential string
}

func NewSource(creds string) []Source {
	return []Source{
		{
			ID:         uuid.NewString(),
			Type:       "shopify",
			URL:        "https://shineos-test.myshopify.com/admin/api/2022-10/",
			Credential: creds,
		},
		{
			ID:         uuid.NewString(),
			Type:       "shopify",
			URL:        "https://shineos-test.myshopify.com/admin/api/2022-10/",
			Credential: creds,
		},
	}
}
