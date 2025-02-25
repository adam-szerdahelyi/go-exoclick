package exoclick

type Variation struct {
	ID               int     `json:"idvariation"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	Active           int     `json:"active"`
	Status           int     `json:"status"`
	Url              string  `json:"url"`
	ImgUrl           string  `json:"imgurl"`
	UrlDescription   string  `json:"durl"`
	OfferID          *int    `json:"offer_id"`
	OfferName        string  `json:"offer_name"`
	FileID           int     `json:"idvariations_file"`
	UrlID            int     `json:"idvariations_url"`
	HtmlID           *int    `json:"idvariations_html"`
	IframeUrlID      *int    `json:"idvariations_iframe_url"`
	IsExplicit       bool    `json:"is_explicit"`
	TestVariationURL string  `json:"test_variation_url"`
	Share            int     `json:"share"`
	FileType         string  `json:"file_type"`
	CalculatedStatus string  `json:"calculated_status"`
	Duration         float64 `json:"duration"`
}

func (v Variation) String() string {
	return Stringify(v)
}
