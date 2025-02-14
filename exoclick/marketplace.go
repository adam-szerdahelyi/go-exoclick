package exoclick

import (
	"context"
	"net/http"
)

type MarketplaceService service

type Marketplace struct {
	SiteHostname       string `json:"site_hostname"`
	SiteOwner          string `json:"site_owner"`
	Description        string `json:"description"`
	URL                string `json:"url"`
	MainCategoryID     int    `json:"maincat"`
	CategoryName       string `json:"category_name"`
	Size               string `json:"size"`
	PublisherAdTypeID  int    `json:"idpublisher_ad_type"`
	ZoneID             int    `json:"idzone"`
	SiteID             int    `json:"idsite"`
	SiteType           int    `json:"site_type"`
	CertifiedLevel     int    `json:"certified_level"`
	DailyImpressions   int    `json:"daily_impressions"`
	DailyClicks        int    `json:"daily_clicks"`
	TrafficType        int    `json:"traffic_type"`
	AdvertiserAdTypeID int    `json:"idadvertiser_ad_type"`
	ImgURL             string `json:"img_url"`
	Blacklisted        int    `json:"blacklisted"`
	SiteBlacklisted    int    `json:"site_blacklisted"`
	BlacklistType      int    `json:"blacklist_type"`
	Alexa              int    `json:"alexa"`
	Idname             int    `json:"idname"`
}

func (c Marketplace) String() string {
	return Stringify(c)
}

type MarketplaceListOptions struct {
	OrderBy string `url:"orderBy,omitempty"`

	ListOptions
}

func (c *MarketplaceService) List(ctx context.Context, opts *MarketplaceListOptions) ([]*Marketplace, *http.Response, error) {
	u := "marketplace"

	if opts.OrderBy == "" {
		opts.OrderBy = "d:daily_impressions"
	}

	u, err := addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	marketplaceResponse := struct {
		Result []*Marketplace `json:"result,omitempty"`
	}{}

	resp, err := c.client.Do(ctx, req, &marketplaceResponse)
	if err != nil {
		return nil, resp, err
	}

	return marketplaceResponse.Result, resp, nil
}
