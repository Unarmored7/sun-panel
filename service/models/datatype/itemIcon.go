package datatype

type ItemIconIconInfo struct {
	ItemType  int    `json:"itemType"`
	Src       string `json:"src"`
	CachedSrc string `json:"cachedSrc,omitempty"`
	Text      string `json:"text"`
	// BgColor  string `json:"bgColor"`
	BackgroundColor string `json:"backgroundColor"`
}
