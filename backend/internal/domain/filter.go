package domain

// Filter narrows a search. The boolean fields are "gap" filters: when true they
// keep only companies that LACK the corresponding contact info. Categories, when
// non-empty, keeps only companies whose Category is in the list.
type Filter struct {
	NoWebsite  bool     `json:"noWebsite"`
	NoSocials  bool     `json:"noSocials"`
	NoPhone    bool     `json:"noPhone"`
	Categories []string `json:"categories"`
}

// Match reports whether a company passes the filter.
//
// Absence semantics: NoWebsite keeps companies with no website at all; the same
// for socials and phone. An empty Categories list matches every category.
func (f Filter) Match(c Company) bool {
	if f.NoWebsite && c.Website != "" {
		return false
	}
	if f.NoSocials && c.HasSocial() {
		return false
	}
	if f.NoPhone && c.Phone != "" {
		return false
	}
	if len(f.Categories) > 0 && !contains(f.Categories, c.Category) {
		return false
	}
	return true
}

func contains(xs []string, target string) bool {
	for _, x := range xs {
		if x == target {
			return true
		}
	}
	return false
}
