package domain

import "testing"

func TestFilterMatch(t *testing.T) {
	withSite := Company{Category: "shop", Website: "https://x.test"}
	noSite := Company{Category: "shop"}
	withInsta := Company{Category: "amenity", Instagram: "@x"}
	withPhone := Company{Category: "office", Phone: "+1"}

	tests := []struct {
		name string
		f    Filter
		c    Company
		want bool
	}{
		{"no filters keeps everything", Filter{}, withSite, true},
		{"noWebsite rejects company with website", Filter{NoWebsite: true}, withSite, false},
		{"noWebsite keeps company without website", Filter{NoWebsite: true}, noSite, true},
		{"noSocials rejects company with instagram", Filter{NoSocials: true}, withInsta, false},
		{"noSocials keeps company without socials", Filter{NoSocials: true}, noSite, true},
		{"noPhone rejects company with phone", Filter{NoPhone: true}, withPhone, false},
		{"noPhone keeps company without phone", Filter{NoPhone: true}, noSite, true},
		{"category whitelist match", Filter{Categories: []string{"shop"}}, noSite, true},
		{"category whitelist miss", Filter{Categories: []string{"office"}}, noSite, false},
		{"empty category list matches any", Filter{Categories: nil}, withInsta, true},
		{"combined filters all pass", Filter{NoWebsite: true, NoPhone: true, Categories: []string{"shop"}}, noSite, true},
		{"combined filters one fails", Filter{NoWebsite: true, NoPhone: true}, withPhone, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.Match(tt.c); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompanyHasSocial(t *testing.T) {
	if (Company{}).HasSocial() {
		t.Error("empty company should have no social")
	}
	if !(Company{Telegram: "@x"}).HasSocial() {
		t.Error("company with telegram should have social")
	}
}
