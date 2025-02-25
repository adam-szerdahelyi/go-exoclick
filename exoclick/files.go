package exoclick

import (
	"context"
	"errors"
	"net/http"
)

type FileService service

type FileType string

const (
	FileTypeImage       FileType = "image"
	FileTypeVideo       FileType = "video"
	FileTypeVideoBanner FileType = "video_banner"
)

type File struct {
	ID                   int      `json:"id"`
	Type                 FileType `json:"type"`
	Width                int      `json:"width"`
	Height               int      `json:"height"`
	WidthPublic          int      `json:"width_public"`
	HeightPublic         int      `json:"height_public"`
	Duration             float64  `json:"duration"`
	IsAdult              int      `json:"is_adult"`
	FileHashOriginal     string   `json:"file_hash_original"`
	FileHashPublic       string   `json:"file_hash_public"`
	FileHashOptimum      *string  `json:"file_hash_optimum"`
	FileExtension        string   `json:"file_extension"`
	FileExtensionOptimum *string  `json:"file_extension_optimum"`
	FileName             string   `json:"file_name"`
	FileSizeOriginal     int      `json:"file_size_original"`
	FileSizePublic       int      `json:"file_size_public"`
	FileSizeOptimum      int      `json:"file_size_optimum"`
	URL                  string   `json:"url"`
	URLOptimum           *string  `json:"url_optimum"`
	IsArchived           *int     `json:"is_archived"`
	Status               int      `json:"status"`
}

func (f File) String() string {
	return Stringify(f)
}

type FileListOptions struct {
	Type         FileType `url:"type,omitempty"`
	ShowArchived bool     `url:"show_archived,omitempty"`
	IsArchived   bool     `url:"is_archived,omitempty"`

	OrderBy string `url:"orderBy,omitempty"`

	ListOptions
}

func (f *FileService) List(ctx context.Context, opts *FileListOptions) ([]*File, *http.Response, error) {
	if opts.Type == "" {
		return nil, nil, errors.New("type must be set")
	}

	u := "library/file"

	if opts.OrderBy == "" {
		opts.OrderBy = "a:id"
	}

	u, err := addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := f.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	filesResponse := struct {
		Result []*File `json:"result,omitempty"`
	}{}

	resp, err := f.client.Do(ctx, req, &filesResponse)
	if err != nil {
		return nil, resp, err
	}

	return filesResponse.Result, resp, nil
}
