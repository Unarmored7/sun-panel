package iconcache

import (
	"bytes"
	"context"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sun-panel/lib/remoteimage"
)

type SiteInfo struct {
	IconURL         string `json:"iconUrl"`
	OriginalIconURL string `json:"originalIconUrl"`
	Title           string `json:"title"`
	Description     string `json:"description"`
}
type candidate struct {
	url   string
	score int
}

func (c *Cache) Discover(ctx context.Context, raw string) (SiteInfo, error) {
	base, err := remoteimage.ValidateURL(raw)
	if err != nil {
		return SiteInfo{}, err
	}
	info := SiteInfo{Title: strings.TrimPrefix(base.Hostname(), "www.")}
	options := []candidate{}
	data, _, final, err := remoteimage.Get(ctx, c.Client, raw, 1<<20)
	if err == nil {
		base = final
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
		if err == nil {
			title := doc.Find("title").First().Text()
			if strings.TrimSpace(title) == "" {
				title, _ = doc.Find(`meta[property="og:title"]`).First().Attr("content")
			}
			site := ""
			for _, selector := range []string{`meta[property="og:site_name"]`, `meta[name="application-name"]`, `meta[name="apple-mobile-web-app-title"]`} {
				if value, _ := doc.Find(selector).First().Attr("content"); strings.TrimSpace(value) != "" {
					site = value
					break
				}
			}
			if name := siteName(title, site, base); name != "" {
				info.Title = name
			}
			description, _ := doc.Find(`meta[name="description"]`).First().Attr("content")
			if description == "" {
				description, _ = doc.Find(`meta[property="og:description"]`).First().Attr("content")
			}
			info.Description = clip(strings.Join(strings.Fields(description), " "), 100)
			if href, exists := doc.Find("base[href]").First().Attr("href"); exists {
				if u, e := base.Parse(href); e == nil && (u.Scheme == "https" || u.Scheme == "http") {
					base = u
				}
			}
			doc.Find("link[href]").Each(func(_ int, s *goquery.Selection) {
				rel, _ := s.Attr("rel")
				rel = strings.ToLower(rel)
				if !strings.Contains(rel, "icon") {
					return
				}
				href, _ := s.Attr("href")
				u, err := base.Parse(strings.TrimSpace(href))
				if err != nil {
					return
				}
				score := 10
				if strings.Contains(rel, "apple-touch-icon") {
					score = 1000
				}
				sizes, _ := s.Attr("sizes")
				for _, size := range strings.Fields(sizes) {
					width, _ := strconv.Atoi(strings.Split(size, "x")[0])
					score += min(width, 512)
				}
				if strings.Contains(rel, "mask-icon") {
					score = -1
				}
				options = append(options, candidate{u.String(), score})
			})
		}
	}
	fallback := &url.URL{Scheme: base.Scheme, Host: base.Host, Path: "/favicon.ico"}
	options = append(options, candidate{fallback.String(), 0})
	sort.SliceStable(options, func(i, j int) bool { return options[i].score > options[j].score })
	seen := map[string]bool{}
	attempts := 0
	for _, option := range options {
		if seen[option.url] {
			continue
		}
		seen[option.url] = true
		attempts++
		if attempts > 4 {
			break
		}
		if local, err := c.Fetch(ctx, option.url); err == nil {
			info.IconURL = local
			info.OriginalIconURL = option.url
			return info, nil
		}
		if ctx.Err() != nil {
			break
		}
	}
	return info, errors.New("no usable icon found")
}

var titleSeparator = regexp.MustCompile(`\s+[|·•–—\-:]\s+|\s*[｜：]\s*`)

// A short bookmark name: the site's own name for a home page, otherwise the
// page-specific part of the title ("requests · PyPI" -> "requests").
func siteName(title, site string, page *url.URL) string {
	site = strings.Join(strings.Fields(site), " ")
	parts := []string{}
	for _, part := range titleSeparator.Split(strings.Join(strings.Fields(title), " "), -1) {
		if part = strings.TrimSpace(part); part != "" {
			parts = append(parts, part)
		}
	}
	home := strings.Trim(page.Path, "/") == "" && page.RawQuery == "" && !strings.HasPrefix(page.Fragment, "/")
	if home && site != "" {
		return clip(site, 20)
	}
	if !home {
		for _, part := range parts {
			if !strings.EqualFold(part, site) {
				return clip(part, 20)
			}
		}
	}
	if site != "" {
		return clip(site, 20)
	}
	// Without a declared name, prefer the title part naming the domain.
	label := strings.Split(strings.TrimPrefix(strings.ToLower(page.Hostname()), "www."), ".")[0]
	for _, part := range parts {
		if strings.Contains(strings.ReplaceAll(strings.ToLower(part), " ", ""), label) {
			return clip(part, 20)
		}
	}
	if len(parts) > 0 {
		return clip(parts[0], 20)
	}
	return ""
}

func clip(value string, limit int) string {
	runes := []rune(value)
	if len(runes) > limit {
		return strings.TrimSpace(string(runes[:limit]))
	}
	return value
}
