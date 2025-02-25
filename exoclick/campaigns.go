package exoclick

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type CampaignsService service

type Campaign struct {
	Campaign           *CampaignData       `json:"campaign,omitempty"`
	CampaignZones      *[]CampaignZones    `json:"zones,omitempty"`
	CampaignCategories *CampaignCategories `json:"categories,omitempty"`
	Variations         *[]Variation        `json:"variations,omitempty"`
	ZoneTargeting      *ZoneTargeting      `json:"zone_targeting,omitempty"`
}

func (c Campaign) String() string {
	return Stringify(c)
}

type CampaignData struct {
	ID           *int          `json:"id,omitempty"`
	Name         *string       `json:"name,omitempty"`
	CampaignType *CampaignType `json:"campaign_type,omitempty"`
	Status       *int          `json:"status,omitempty"`
	PricingModel *int          `json:"pricing_model,omitempty"`
	Price        *float64      `json:"price,omitempty"`
	DateCreated  *CustomDate   `json:"date_created,omitempty"`
}

type CampaignType struct {
	ID   int    `json:"id" db:"campaign_type"`
	Name string `json:"name"`
}

type CampaignZones struct {
	CampaignID      *int     `json:"idcampaign"`
	ZoneID          *int     `json:"idzone"`
	Price           *float64 `json:"price"`
	SubIDTargetType *int     `json:"sub_id_target_type"`
	SiteID          *int     `json:"idsite"`
	SubIDs          *[]int   `json:"sub_ids"`
}

type CampaignCategories struct {
	Targeted []Category `json:"targeted"`
	Blocked  []Category `json:"blocked"`
}

type ZoneTargeting struct {
	Type int `json:"type" db:"zone_targeting_type,omitempty"`
}

type CampaignListOptions struct {
	Status       int    `url:"status,omitempty"`
	CustomSearch string `url:"custom_search,omitempty"`
	OrderBy      string `url:"orderBy,omitempty"`

	ListOptions
}

func (c *CampaignsService) List(ctx context.Context, opts *CampaignListOptions) ([]*CampaignData, *http.Response, error) {
	u := "campaigns"

	if opts.OrderBy == "" {
		opts.OrderBy = "d:id"
	}

	u, err := addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	campaignsResponse := struct {
		Result []*CampaignData `json:"result,omitempty"`
	}{}

	resp, err := c.client.Do(ctx, req, &campaignsResponse)
	if err != nil {
		return nil, resp, err
	}

	return campaignsResponse.Result, resp, nil
}

func (c *CampaignsService) Get(ctx context.Context, id int, isDetailed bool) (*Campaign, *http.Response, error) {
	u := fmt.Sprintf("campaigns/%d", id)
	u, err := addOptions(u, struct{ detailed bool }{detailed: isDetailed})
	if err != nil {
		return nil, nil, err
	}

	req, err := c.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	campaignResponse := struct {
		Result Campaign `json:"result,omitempty"`
	}{}

	resp, err := c.client.Do(ctx, req, &campaignResponse)
	if err != nil {
		return nil, resp, err
	}

	return &campaignResponse.Result, resp, nil
}

type TargetingType uint8

const (
	Target TargetingType = 1 + iota
	Block
)

type TargetingOptions struct {
	Type TargetingType
}

func (c *CampaignsService) ToggleCategories(ctx context.Context, campaignID int, categories []int, opts TargetingOptions) (*http.Response, error) {
	u := fmt.Sprintf("campaigns/%d/targeted/categories", campaignID)

	if len(categories) == 0 {
		return nil, errors.New("categories array cannot be empty")
	}

	var method string

	switch opts.Type {
	case Target:
		method = http.MethodPost
	case Block:
		method = http.MethodDelete
	default:
		return nil, fmt.Errorf("unsupported targeting type %d", opts.Type)
	}

	req, err := c.client.NewRequest(method, u, categories)
	if err != nil {
		return nil, err
	}

	return c.client.Do(ctx, req, nil)
}

func (c *CampaignsService) TargetCategories(ctx context.Context, campaignID int, categories []int) (*http.Response, error) {
	return c.ToggleCategories(ctx, campaignID, categories, TargetingOptions{Target})
}

func (c *CampaignsService) BlockCategories(ctx context.Context, campaignID int, categories []int) (*http.Response, error) {
	return c.ToggleCategories(ctx, campaignID, categories, TargetingOptions{Block})
}
