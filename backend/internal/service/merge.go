package service

import (
	"fmt"
	"strings"

	"github.com/parse-companies/backend/internal/domain"
)

// dedupKey identifies a company for cross-source deduplication. Named companies
// merge across sources when they share a name and a rounded location; unnamed
// ones fall back to their source identity so they never collapse together.
func dedupKey(c domain.Company) string {
	name := strings.ToLower(strings.TrimSpace(c.Name))
	if name == "" {
		return c.OSMType + "/" + c.OSMID
	}
	return fmt.Sprintf("%s|%.3f,%.3f", name, c.Lat, c.Lon)
}

// mergeInto fills empty fields of dst from src (dst wins when both are set).
func mergeInto(dst, src domain.Company) domain.Company {
	dst.Name = firstNonEmpty(dst.Name, src.Name)
	dst.Category = firstNonEmpty(dst.Category, src.Category)
	dst.Subcat = firstNonEmpty(dst.Subcat, src.Subcat)
	dst.Website = firstNonEmpty(dst.Website, src.Website)
	dst.Phone = firstNonEmpty(dst.Phone, src.Phone)
	dst.Email = firstNonEmpty(dst.Email, src.Email)
	dst.Instagram = firstNonEmpty(dst.Instagram, src.Instagram)
	dst.Facebook = firstNonEmpty(dst.Facebook, src.Facebook)
	dst.VK = firstNonEmpty(dst.VK, src.VK)
	dst.Telegram = firstNonEmpty(dst.Telegram, src.Telegram)
	dst.OpeningHours = firstNonEmpty(dst.OpeningHours, src.OpeningHours)
	if dst.Addr == (domain.Address{}) {
		dst.Addr = src.Addr
	}
	return dst
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// accumulator deduplicates a growing set of companies across sources while
// preserving insertion order.
type accumulator struct {
	seen  map[string]int
	items []domain.Company
}

func newAccumulator() *accumulator {
	return &accumulator{seen: make(map[string]int)}
}

// add merges the batch into the set and returns only the companies that were new
// (not merges into existing ones) — useful for incremental streaming.
func (a *accumulator) add(batch []domain.Company) []domain.Company {
	var fresh []domain.Company
	for _, c := range batch {
		k := dedupKey(c)
		if i, ok := a.seen[k]; ok {
			a.items[i] = mergeInto(a.items[i], c)
			continue
		}
		a.seen[k] = len(a.items)
		a.items = append(a.items, c)
		fresh = append(fresh, c)
	}
	return fresh
}

func (a *accumulator) all() []domain.Company { return a.items }
