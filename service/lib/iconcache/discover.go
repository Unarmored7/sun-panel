package iconcache

import (
	"bytes"
	"context"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"net/url"
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
			title, _ := doc.Find(`meta[property="og:title"]`).First().Attr("content")
			if strings.TrimSpace(title) == "" {
				title = doc.Find("title").First().Text()
			}
			if strings.TrimSpace(title) != "" {
				info.Title = strings.Join(strings.Fields(title), " ")
			}
			description, _ := doc.Find(`meta[name="description"]`).First().Attr("content")
			if description == "" {
				description, _ = doc.Find(`meta[property="og:description"]`).First().Attr("content")
			}
			info.Description = strings.Join(strings.Fields(description), " ")
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
