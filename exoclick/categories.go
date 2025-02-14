package exoclick

import (
	"context"
	"net/http"
)

type CategoryService service

type Category struct {
	ID         *int    `json:"id,omitempty"`
	Name       *string `json:"name,omitempty"`
	LongName   *string `json:"long_name,omitempty"`
	Parent     *int    `json:"parent,omitempty"`
	Selectable *int    `json:"selectable,omitempty"`
	Enabled    *int    `json:"enabled,omitempty"`
	Deleted    *int    `json:"deleted,omitempty"`
}

func (c Category) String() string {
	return Stringify(c)
}

type CategoryListOptions struct {
	OrderBy string `url:"orderBy,omitempty"`

	ListOptions
}

func (c *CategoryService) List(ctx context.Context, opts *CategoryListOptions) ([]*Category, *http.Response, error) {
	u := "collections/categories"

	if opts.OrderBy == "" {
		opts.OrderBy = "a:id"
	}

	u, err := addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	categoriesResponse := struct {
		Result []*Category `json:"result,omitempty"`
	}{}

	resp, err := c.client.Do(ctx, req, &categoriesResponse)
	if resp.StatusCode != http.StatusNotFound && err != nil {
		return nil, resp, err
	}

	return categoriesResponse.Result, resp, nil
}
