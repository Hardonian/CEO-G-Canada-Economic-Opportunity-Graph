package ckan

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/safefetch"
)

const (
	DefaultPageSize    = 100
	MaxPageSize        = 1000
	DefaultMaxPages    = 100
	DefaultMaxDatasets = 10_000
)

// Fetcher is the GET-only transport required by catalog discovery.
type Fetcher interface {
	Get(ctx context.Context, rawURL string, checkpoint safefetch.Checkpoint) (*safefetch.Response, error)
}

// SearchOptions maps directly to safe, read-only CKAN package_search query
// parameters.
type SearchOptions struct {
	Query       string
	FilterQuery string
	Sort        string
	Start       int
	Rows        int
}

// RecentlyChangedOptions bounds the CKAN activity feed.
type RecentlyChangedOptions struct {
	Offset int
	Limit  int
}

// ClientOption customizes bounded catalog enumeration.
type ClientOption func(*Client) error

// WithPageSize changes the default package_search page size.
func WithPageSize(rows int) ClientOption {
	return func(client *Client) error {
		if rows < 1 || rows > MaxPageSize {
			return fmt.Errorf("CKAN page size must be between 1 and %d", MaxPageSize)
		}
		client.pageSize = rows
		return nil
	}
}

// WithDiscoveryLimits bounds recursive pagination independently of the remote
// catalog's advertised count.
func WithDiscoveryLimits(maxPages, maxDatasets int) ClientOption {
	return func(client *Client) error {
		if maxPages < 1 || maxDatasets < 1 {
			return fmt.Errorf("CKAN discovery limits must be positive")
		}
		client.maxPages = maxPages
		client.maxDatasets = maxDatasets
		return nil
	}
}

// Client discovers CKAN datasets and resources without activating or fetching
// the downstream resources it finds.
type Client struct {
	catalogURL  string
	actionBase  string
	fetcher     Fetcher
	pageSize    int
	maxPages    int
	maxDatasets int
}

// NewClient accepts a CKAN catalog root, /api/3 path, or /api/3/action path.
// A nil fetcher uses the shared public-network safe fetcher.
func NewClient(catalogURL string, fetcher Fetcher, options ...ClientOption) (*Client, error) {
	normalized, actionBase, err := normalizeCatalogURL(catalogURL)
	if err != nil {
		return nil, err
	}
	if fetcher == nil {
		fetcher = safefetch.NewClient(safefetch.ClientOptions{})
	}
	client := &Client{
		catalogURL:  normalized,
		actionBase:  actionBase,
		fetcher:     fetcher,
		pageSize:    DefaultPageSize,
		maxPages:    DefaultMaxPages,
		maxDatasets: DefaultMaxDatasets,
	}
	for _, option := range options {
		if option != nil {
			if err := option(client); err != nil {
				return nil, err
			}
		}
	}
	return client, nil
}

func (client *Client) CatalogURL() string { return client.catalogURL }

// PackageSearchURL constructs a GET URL without interpolating query text into
// the path or request body.
func (client *Client) PackageSearchURL(options SearchOptions) (string, error) {
	rows, err := client.searchRows(options.Rows)
	if err != nil {
		return "", err
	}
	if options.Start < 0 {
		return "", fmt.Errorf("CKAN package_search start must be non-negative")
	}
	if len(options.Query) > 2048 || len(options.FilterQuery) > 4096 || len(options.Sort) > 512 {
		return "", fmt.Errorf("CKAN package_search query exceeds local limits")
	}
	action, err := url.Parse(client.actionBase + "/package_search")
	if err != nil {
		return "", err
	}
	query := action.Query()
	query.Set("start", strconv.Itoa(options.Start))
	query.Set("rows", strconv.Itoa(rows))
	if options.Query != "" {
		query.Set("q", options.Query)
	}
	if options.FilterQuery != "" {
		query.Set("fq", options.FilterQuery)
	}
	if options.Sort != "" {
		query.Set("sort", options.Sort)
	}
	action.RawQuery = query.Encode()
	return action.String(), nil
}

// RecentlyChangedURL constructs a GET URL for CKAN's activity stream.
func (client *Client) RecentlyChangedURL(options RecentlyChangedOptions) (string, error) {
	if options.Offset < 0 {
		return "", fmt.Errorf("CKAN recently-changed offset must be non-negative")
	}
	limit := options.Limit
	if limit == 0 {
		limit = client.pageSize
	}
	if limit < 1 || limit > MaxPageSize {
		return "", fmt.Errorf("CKAN recently-changed limit must be between 1 and %d", MaxPageSize)
	}
	action, err := url.Parse(client.actionBase + "/recently_changed_packages_activity_list")
	if err != nil {
		return "", err
	}
	query := action.Query()
	query.Set("offset", strconv.Itoa(options.Offset))
	query.Set("limit", strconv.Itoa(limit))
	action.RawQuery = query.Encode()
	return action.String(), nil
}

// SearchPage fetches and parses one conditional package_search page.
func (client *Client) SearchPage(ctx context.Context, options SearchOptions, checkpoint safefetch.Checkpoint) (*SearchPage, error) {
	rawURL, err := client.PackageSearchURL(options)
	if err != nil {
		return nil, err
	}
	response, err := client.fetcher.Get(ctx, rawURL, checkpoint)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, fmt.Errorf("CKAN fetcher returned a nil response")
	}
	if response.NotModified {
		return &SearchPage{Start: options.Start, Checkpoint: response.Checkpoint(), NotModified: true}, nil
	}
	page, err := ParsePackageSearch(client.catalogURL, response.Body, options.Start)
	if err != nil {
		return nil, err
	}
	page.Checkpoint = response.Checkpoint()
	return page, nil
}

// Discover follows package_search pagination under explicit page and dataset
// ceilings. It errors on repeated rows or a stalled catalog instead of looping.
func (client *Client) Discover(ctx context.Context, options SearchOptions) (*DiscoveryResult, error) {
	rows, err := client.searchRows(options.Rows)
	if err != nil {
		return nil, err
	}
	if options.Start < 0 {
		return nil, fmt.Errorf("CKAN package_search start must be non-negative")
	}
	result := &DiscoveryResult{Checkpoints: make(map[int]safefetch.Checkpoint)}
	seen := make(map[string]struct{})
	start := options.Start
	for pageNumber := 0; pageNumber < client.maxPages; pageNumber++ {
		pageOptions := options
		pageOptions.Rows = rows
		pageOptions.Start = start
		page, err := client.SearchPage(ctx, pageOptions, safefetch.Checkpoint{})
		if err != nil {
			return nil, err
		}
		if page.NotModified {
			return nil, fmt.Errorf("unexpected not-modified response without a discovery checkpoint")
		}
		result.Pages++
		result.TotalAvailable = page.Total
		result.Checkpoints[start] = page.Checkpoint
		for _, dataset := range page.Datasets {
			if _, duplicate := seen[dataset.ID]; duplicate {
				return nil, fmt.Errorf("CKAN pagination repeated dataset %q", dataset.RemoteID)
			}
			seen[dataset.ID] = struct{}{}
			result.Datasets = append(result.Datasets, dataset)
			if len(result.Datasets) > client.maxDatasets {
				return nil, fmt.Errorf("CKAN discovery exceeded %d-dataset limit", client.maxDatasets)
			}
		}
		if len(page.Datasets) == 0 {
			if start < page.Total {
				return nil, fmt.Errorf("CKAN pagination stalled at start=%d of %d", start, page.Total)
			}
			return result, nil
		}
		start += len(page.Datasets)
		if start >= page.Total {
			return result, nil
		}
	}
	return nil, fmt.Errorf("CKAN discovery exceeded %d-page limit", client.maxPages)
}

// RecentlyChanged conditionally retrieves one activity page.
func (client *Client) RecentlyChanged(ctx context.Context, options RecentlyChangedOptions, checkpoint safefetch.Checkpoint) (*ChangePage, error) {
	rawURL, err := client.RecentlyChangedURL(options)
	if err != nil {
		return nil, err
	}
	response, err := client.fetcher.Get(ctx, rawURL, checkpoint)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, fmt.Errorf("CKAN fetcher returned a nil response")
	}
	limit := options.Limit
	if limit == 0 {
		limit = client.pageSize
	}
	page := &ChangePage{Offset: options.Offset, Limit: limit, Checkpoint: response.Checkpoint(), NotModified: response.NotModified}
	if response.NotModified {
		return page, nil
	}
	changes, err := ParseRecentlyChanged(client.catalogURL, response.Body)
	if err != nil {
		return nil, err
	}
	page.Changes = changes
	return page, nil
}

func (client *Client) searchRows(requested int) (int, error) {
	if requested == 0 {
		return client.pageSize, nil
	}
	if requested < 1 || requested > MaxPageSize {
		return 0, fmt.Errorf("CKAN package_search rows must be between 1 and %d", MaxPageSize)
	}
	return requested, nil
}

func normalizeCatalogURL(rawURL string) (catalogURL, actionBase string, err error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", "", fmt.Errorf("parse CKAN catalog URL: %w", err)
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", "", fmt.Errorf("CKAN catalog URL must use HTTP(S)")
	}
	if parsed.Hostname() == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", "", fmt.Errorf("CKAN catalog URL must have a host and no user information, query, or fragment")
	}
	parsed.Host = strings.ToLower(parsed.Host)
	cleanPath := strings.TrimSuffix(parsed.EscapedPath(), "/")
	for _, marker := range []string{"/api/3/action/package_search", "/api/3/action/recently_changed_packages_activity_list", "/api/3/action", "/api/3"} {
		if strings.HasSuffix(strings.ToLower(cleanPath), marker) {
			cleanPath = cleanPath[:len(cleanPath)-len(marker)]
			break
		}
	}
	if cleanPath == "." || cleanPath == "/" {
		cleanPath = ""
	} else {
		cleanPath = path.Clean("/" + strings.TrimPrefix(cleanPath, "/"))
	}
	parsed.RawPath = ""
	parsed.Path = cleanPath
	catalogURL = strings.TrimSuffix(parsed.String(), "/")
	actionBase = catalogURL + "/api/3/action"
	return catalogURL, actionBase, nil
}
