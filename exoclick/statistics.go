package exoclick

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
)

type StatisticsService service

type CustomFloat float64

func (cf *CustomFloat) UnmarshalCSV(csv string) (err error) {
	s := strings.Replace(csv, ",", "", -1)
	t, err := strconv.ParseFloat(s, 64)

	*cf = CustomFloat(t)

	return err
}

type Statistic struct {
	Date             *time.Time  `csv:"date,omitempty"`
	CampaignID       *int        `csv:"campaign_id,omitempty"`
	VariationID      *int        `csv:"variation_id,omitempty"`
	SiteID           *int        `csv:"site_id,omitempty"`
	SiteName         *string     `csv:"site_name,omitempty"`
	ZoneID           *int        `csv:"zone_id,omitempty"`
	ZoneName         *string     `csv:"zone_name,omitempty"`
	CategoryID       *int        `csv:"category_id,omitempty"`
	Clicks           int         `csv:"clicks"`
	Impressions      int         `csv:"impressions"`
	VideoImpressions int         `csv:"video_impressions"`
	VideoViews       int         `csv:"video_views"`
	G1               int         `csv:"g1"`
	G5               int         `csv:"g5"`
	Cost             CustomFloat `csv:"cost"`
}

func (s Statistic) String() string {
	return Stringify(s)
}

var DefaultTimezone = TimeZone{time.UTC}

type StatisticsField string

const (
	Date             StatisticsField = "date"
	Hour             StatisticsField = "hour"
	CampaignID       StatisticsField = "campaign_id"
	VariationID      StatisticsField = "variation_id"
	SiteID           StatisticsField = "site_id"
	SiteName         StatisticsField = "site_name"
	ZoneID           StatisticsField = "zone_id"
	ZoneName         StatisticsField = "zone_name"
	CategoryID       StatisticsField = "category_id"
	Clicks           StatisticsField = "clicks"
	Impressions      StatisticsField = "impressions"
	VideoImpressions StatisticsField = "video_impressions"
	VideoViews       StatisticsField = "video_views"
	G1               StatisticsField = "g1"
	G5               StatisticsField = "g5"
	Cost             StatisticsField = "cost"
)

type StatisticsOptions struct {
	Timezone        *TimeZone           `json:"timezone,omitempty"`
	Filter          StatisticsFilters   `json:"filter,omitempty"`
	GroupBy         []StatisticsField   `json:"group_by,omitempty"`
	OrderBy         []StatisticsOrderBy `json:"order_by,omitempty"`
	OutputCsvFields []StatisticsField   `json:"output_csv_fields,omitempty"`
	Detailed        bool                `json:"detailed,omitempty"`
	ListOptions
}

type StatisticsFilters struct {
	DateFrom       CustomDate `json:"date_from,omitempty"`
	DateTo         CustomDate `json:"date_to,omitempty"`
	Hour           []int      `json:"hour,omitempty"`
	CampaignID     int        `json:"campaign_id,omitempty"`
	VariationID    int        `json:"variation_id,omitempty"`
	SiteID         int        `json:"site_id,omitempty"`
	ZoneID         int        `json:"zone_id,omitempty"`
	CategoryID     int        `json:"category_id,omitempty"`
	ExcludeDeleted int        `json:"exclude_deleted,omitempty"`
}

var ValidGroupByFields = []StatisticsField{
	CampaignID,
	CategoryID,
	Date,
	Hour,
	SiteID,
	ZoneID,
	VariationID,
}

var FieldsRequireDetailed = []StatisticsField{
	SiteName,
	ZoneName,
}

type OrderType string

const (
	Asc  OrderType = "asc"
	Desc OrderType = "desc"
)

type StatisticsOrderBy struct {
	Field StatisticsField `json:"field,omitempty"`
	Order OrderType       `json:"order,omitempty"`
}

// func (s *StatisticsService) GetStatistics(ctx context.Context, opts *StatisticsOptions) ([]*Statistic, *http.Response, error) {
// 	u := "statistics/a/global"

// 	if err := validateStatisticsOptions(opts); err != nil {
// 		return nil, nil, err
// 	}

// 	req, err := s.client.NewRequest(http.MethodPost, u, opts)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	statisticsResponse := struct {
// 		Result []*Statistic `json:"result,omitempty"`
// 	}{}

// 	resp, err := s.client.Do(ctx, req, &statisticsResponse)
// 	if err != nil {
// 		return nil, resp, err
// 	}

// 	return statisticsResponse.Result, resp, nil
// }

func (s *StatisticsService) GetStatisticsCSV(ctx context.Context, opts *StatisticsOptions) ([]*Statistic, *http.Response, error) {
	u := "statistics/a/global"

	if err := validateStatisticsOptions(opts); err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(http.MethodPost, u, opts)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Accept", "text/csv")

	resp, err := s.client.BareDo(ctx, req)
	if err != nil {
		return nil, resp, err
	}

	defer resp.Body.Close()

	var statistics []*Statistic

	if err := gocsv.Unmarshal(resp.Body, &statistics); err != nil {
		return nil, resp, err
	}

	return statistics, resp, nil
}

func validateStatisticsOptions(opts *StatisticsOptions) error {
	if opts.Filter.DateFrom.Time.After(opts.Filter.DateTo.Time) {
		return errors.New("date from must be before or equal to date to")
	}

	if len(opts.Filter.Hour) > 0 && opts.Timezone == nil {
		return errors.New("timezone must be set if filter hour is set")
	}

	for _, hour := range opts.Filter.Hour {
		if hour < 1 || hour > 23 {
			return fmt.Errorf("invalid hour: hour must be between 1 and 23, but got %d", hour)
		}
	}

	if len(opts.GroupBy) > 4 {
		return fmt.Errorf("invalid group by: maximum of 4 fields allowed, but got %d", len(opts.GroupBy))
	}

	if len(opts.OrderBy) > 2 {
		return fmt.Errorf("invalid order by: maximum of 2 fields allowed, but got %d", len(opts.OrderBy))
	}

	if len(opts.OutputCsvFields) == 0 {
		return fmt.Errorf("invalid output csv fields: minimum of 1 fields required, but got %d", len(opts.OutputCsvFields))
	}

	if !opts.Detailed {
		for _, outputField := range opts.OutputCsvFields {
			if slices.Contains(FieldsRequireDetailed, outputField) {
				return fmt.Errorf("\"%s\" field requires detailed enabled", outputField)
			}
		}
	}

	for _, groupBy := range opts.GroupBy {
		if !slices.Contains(ValidGroupByFields, groupBy) {
			return fmt.Errorf("invalid group by field: \"%s\"", groupBy)
		}

		if !slices.Contains(opts.OutputCsvFields, groupBy) {
			return fmt.Errorf("invalid output csv fields: must contain all group by fields, but \"%s\" is missing", groupBy)
		}
	}

	for _, orderBy := range opts.OrderBy {
		if orderBy.Field == "" {
			return errors.New("missing order by Field")
		}

		if orderBy.Order == "" {
			return errors.New("missing order by Order")
		}

		if !slices.Contains(opts.OutputCsvFields, orderBy.Field) {
			return fmt.Errorf("invalid output csv fields: must contain all order by fields, but \"%s\" is missing", orderBy.Field)
		}
	}

	return nil
}
