package domain

// Language represents a supported language.
type Language struct {
	Code  string            `json:"code"`
	Names map[string]string `json:"names"`
	Flag  string            `json:"flag"`
}

// InterestCategory represents a group of interests.
type InterestCategory struct {
	ID    string            `json:"id"`
	Slug  string            `json:"slug"`
	Names map[string]string `json:"names"`
	Order int               `json:"order"`
}

// Interest represents a specific topic.
type Interest struct {
	ID         string            `json:"id"`
	CategoryID string            `json:"categoryId"`
	Slug       string            `json:"slug"`
	Names      map[string]string `json:"names"`
}
