package paprika

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// roundTripper is a wrapper around http.RoundTripper
// that adds the specified headers to each request
type roundTripper struct {
	headers   map[string]string
	transport http.RoundTripper
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range r.headers {
		// Only set if not already present
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}
	return r.transport.RoundTrip(req)
}

func userAgent(version string) string {
	return fmt.Sprintf("paprika-3-mcp/%s (golang; %s)", version, runtime.Version())
}

func NewClient(username, password, version string, logger *slog.Logger) (*Client, error) {
	// Create the http client & login to retrieve an authentication token
	t := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			d := &net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}
			return d.DialContext(ctx, network, addr)
		},
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	client := &http.Client{
		Transport: t,
		Timeout:   10 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	token, err := login(ctx, *client, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	client.Transport = &roundTripper{
		transport: t,
		headers: map[string]string{
			"Accept":        "*/*",
			"Authorization": fmt.Sprintf("Bearer %s", token),
			"Connection":    "keep-alive",
			"User-Agent":    userAgent(version),
		},
	}

	l := logger
	if l == nil {
		l = slog.Default()
	}

	return &Client{
		client:   client,
		logger:   l,
		username: username,
		password: password,
	}, nil
}

type Client struct {
	client   *http.Client
	logger   *slog.Logger
	username string
	password string
}

type loginResponse struct {
	Result struct {
		Token string `json:"token"`
	} `json:"result"`
}

type errorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// login authenticates with the Paprika API and returns an authentication token
// The token is used for all subsequent requests to the API. As far as I can tell, this is a JWT with no expiration.
func login(ctx context.Context, client http.Client, username, password string) (string, error) {
	body := fmt.Sprintf("email=%s&password=%s", username, password)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://paprikaapp.com/api/v1/account/login", bytes.NewBufferString(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to login: %s", resp.Status)
	}

	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var loginResp loginResponse
	if err := json.Unmarshal(rawBytes, &loginResp); err != nil {
		return "", err
	}

	if loginResp.Result.Token == "" {
		return "", fmt.Errorf("failed to get token: %s", string(rawBytes))
	}

	return loginResp.Result.Token, nil
}

type RecipeList struct {
	Result []struct {
		UID  string `json:"uid"`
		Hash string `json:"hash"`
	} `json:"result"`
}

// ListRecipes retrieves a list of recipes from the Paprika API - the response objects
// only contain the UID and hash of each recipe, not the full recipe object
func (c *Client) ListRecipes(ctx context.Context) (*RecipeList, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://paprikaapp.com/api/v2/sync/recipes", nil)
	if err != nil {
		c.logger.Error("failed to create request", "error", err)
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("failed to get recipes", "error", err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		c.logger.Error("failed to get recipes", "status", resp.Status)
		return nil, fmt.Errorf("failed to get recipes: %s", resp.Status)
	}

	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("failed to read response body", "error", err)
		return nil, err
	}
	var recipeList RecipeList
	if err := json.Unmarshal(rawBytes, &recipeList); err != nil {
		c.logger.Error("failed to unmarshal response", "error", err)
		return nil, err
	}

	c.logger.Info("found recipes", "count", len(recipeList.Result))
	return &recipeList, nil
}

type Recipe struct {
	UID             string   `json:"uid"`
	Name            string   `json:"name"`
	Ingredients     string   `json:"ingredients"`
	Directions      string   `json:"directions"`
	Description     string   `json:"description"`
	Notes           string   `json:"notes"`
	NutritionalInfo string   `json:"nutritional_info"`
	Servings        string   `json:"servings"`
	Difficulty      string   `json:"difficulty"`
	PrepTime        string   `json:"prep_time"`
	CookTime        string   `json:"cook_time"`
	TotalTime       string   `json:"total_time"`
	Source          string   `json:"source"`
	SourceURL       string   `json:"source_url"`
	ImageURL        string   `json:"image_url"`
	Photo           string   `json:"photo"`
	PhotoHash       string   `json:"photo_hash"`
	PhotoLarge      string   `json:"photo_large"`
	Scale           string   `json:"scale"`
	Hash            string   `json:"hash"`
	Categories      []string `json:"categories"`
	Rating          int      `json:"rating"`
	InTrash         bool     `json:"in_trash"`
	IsPinned        bool     `json:"is_pinned"`
	OnFavorites     bool     `json:"on_favorites"`
	OnGroceryList   bool     `json:"on_grocery_list"`
	Created         string   `json:"created"`
	PhotoURL        string   `json:"photo_url"`
}

func (r *Recipe) ResourceDescription() string {
	if len(r.Description) == 0 {
		return fmt.Sprintf("A recipe for %s", r.Name)
	}

	return fmt.Sprintf("A recipe for %s: %s", r.Name, r.Description)
}

func (r *Recipe) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", r.Name))

	if r.Description != "" {
		sb.WriteString(fmt.Sprintf("_%s_\n\n", r.Description))
	}

	if r.Servings != "" || r.PrepTime != "" || r.CookTime != "" || r.Difficulty != "" {
		sb.WriteString("## Details\n")
		if r.Servings != "" {
			sb.WriteString(fmt.Sprintf("- **Servings:** %s\n", r.Servings))
		}
		if r.PrepTime != "" {
			sb.WriteString(fmt.Sprintf("- **Prep Time:** %s\n", r.PrepTime))
		}
		if r.CookTime != "" {
			sb.WriteString(fmt.Sprintf("- **Cook Time:** %s\n", r.CookTime))
		}
		if r.Difficulty != "" {
			sb.WriteString(fmt.Sprintf("- **Difficulty:** %s\n", r.Difficulty))
		}
		sb.WriteString("\n")
	}

	if r.Ingredients != "" {
		sb.WriteString("## Ingredients\n")
		for _, line := range strings.Split(strings.TrimSpace(r.Ingredients), "\n") {
			if line != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", line))
			}
		}
		sb.WriteString("\n")
	}

	if r.Directions != "" {
		sb.WriteString("## Directions\n")
		lines := strings.Split(strings.TrimSpace(r.Directions), "\n")
		for i, line := range lines {
			if line != "" {
				sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, line))
			}
		}
		sb.WriteString("\n")
	}

	if r.Notes != "" {
		sb.WriteString("## Notes\n")
		sb.WriteString(r.Notes + "\n\n")
	}

	return sb.String()
}

func (r *Recipe) generateUUID() {
	// Generate a new UUID for the recipe
	if r.UID == "" {
		r.UID = strings.ToUpper(uuid.New().String())
		return
	}

	r.UID = strings.ToUpper(r.UID)
}

func (r *Recipe) updateCreated() {
	layout := "2006-01-02 15:04:05"
	r.Created = time.Now().Format(layout)
}

func (r *Recipe) asMap() (map[string]interface{}, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	var fields map[string]interface{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}

	return fields, nil
}

func (r *Recipe) updateHash() error {
	fields, err := r.asMap()
	if err != nil {
		return err
	}

	// Remove the "hash" field
	delete(fields, "hash")

	// Sort keys manually to ensure consistent JSON output
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build a sorted map for consistent hashing
	sorted := make(map[string]interface{}, len(fields))
	for _, k := range keys {
		sorted[k] = fields[k]
	}

	// Marshal the sorted map to JSON
	jsonBytes, err := json.Marshal(sorted)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(jsonBytes)
	r.Hash = hex.EncodeToString(hash[:])
	return nil
}

func (r *Recipe) asGzip() ([]byte, error) {
	jsonBytes, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return gzipBytes(jsonBytes)
}

type GetRecipeResponse struct {
	Result Recipe `json:"result"`
}

func (c *Client) GetRecipe(ctx context.Context, uid string) (*Recipe, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://paprikaapp.com/api/v2/sync/recipe/%s/", uid), nil)
	if err != nil {
		c.logger.Error("failed to create request", "error", err)
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("failed to get recipe", "error", err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		c.logger.Error("failed to get recipe", "status", resp.Status)
		return nil, fmt.Errorf("failed to get recipe: %s", resp.Status)
	}

	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("failed to read response body", "error", err)
		return nil, err
	}

	var recipeResp GetRecipeResponse
	if err := json.Unmarshal(rawBytes, &recipeResp); err != nil {
		c.logger.Error("failed to unmarshal response", "error", err)
		return nil, err
	}

	return &recipeResp.Result, nil
}

// GetRecipeRaw fetches the raw JSON bytes for a recipe — used to discover all API fields.
func (c *Client) GetRecipeRaw(ctx context.Context, uid string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://paprikaapp.com/api/v2/sync/recipe/%s/", uid), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("failed to get recipe: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) DeleteRecipe(ctx context.Context, recipe Recipe) (*Recipe, error) {
	// Set the recipe to be in the trash
	// TODO: reverse-engineer full deletions; currently a user must go in-app to empty their trash and fully delete something
	recipe.InTrash = true
	return c.SaveRecipe(ctx, recipe)
}

// SaveRecipe saves a recipe to the Paprika API. If the recipe already exists, it will be updated.
// If the recipe does not exist, it will be created.
func (c *Client) SaveRecipe(ctx context.Context, recipe Recipe) (*Recipe, error) {
	// set the created timestamp
	recipe.updateCreated()
	// generate a new UUID if one doesn't exist
	recipe.generateUUID()
	// generate a hash of the recipe object
	if err := recipe.updateHash(); err != nil {
		return nil, err
	}

	fileData, err := recipe.asGzip()
	if err != nil {
		return nil, err
	}

	body, contentType, err := buildMultipartBody(fileData)
	if err != nil {
		c.logger.Error("failed to build multipart body", "error", err)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://paprikaapp.com/api/v2/sync/recipe/%s/", recipe.UID), body)
	if err != nil {
		c.logger.Error("failed to create request", "error", err)
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = int64(body.Len())

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("failed to create recipe", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		c.logger.Error("failed to create recipe", "status", resp.Status)
		return nil, fmt.Errorf("failed to create recipe: %s", resp.Status)
	}

	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("failed to read response body", "error", err)
		return nil, err
	}

	if err := isErrorResponse(rawBytes); err != nil {
		c.logger.Error("failed to create recipe", "error", err)
		return nil, err
	}

	defer c.notify(ctx)

	return &recipe, nil
}

type GroceryItem struct {
	UID         string  `json:"uid"`
	Name        string  `json:"name"`
	Ingredient  string  `json:"ingredient"`
	Quantity    string  `json:"quantity"`
	Recipe      *string `json:"recipe"`
	RecipeUID   *string `json:"recipe_uid"`
	ListUID     string  `json:"list_uid"`
	Aisle       string  `json:"aisle"`
	AisleUID    string  `json:"aisle_uid,omitempty"`
	Instruction string  `json:"instruction"`
	Separate    bool    `json:"separate"`
	Purchased   bool    `json:"purchased"`
	OrderFlag   int     `json:"order_flag"`
	Deleted     bool    `json:"deleted"`
}


func (c *Client) ListGroceryItems(ctx context.Context) ([]GroceryItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://paprikaapp.com/api/v2/sync/groceries/", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("paprika API error %d: %s", resp.StatusCode, b)
	}

	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Result []GroceryItem `json:"result"`
	}
	return result.Result, json.Unmarshal(rawBytes, &result)
}

// SaveGroceryItem creates a new grocery item via the V1 sync endpoint.
// It follows the same gzip+multipart+Basic Auth pattern as SaveMealPlanEntry and SavePantryItem.
func (c *Client) SaveGroceryItem(ctx context.Context, item GroceryItem) error {
	if item.UID == "" {
		item.UID = newUID()
	}
	item.Deleted = false

	data, err := json.Marshal([]GroceryItem{item})
	if err != nil {
		return err
	}
	gz, err := gzipBytes(data)
	if err != nil {
		return err
	}

	body, contentType, err := buildMultipartBody(gz)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://paprikaapp.com/api/v1/sync/groceries/", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	req.SetBasicAuth(c.username, c.password)
	req.ContentLength = int64(body.Len())

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("paprika API error %d: %s", resp.StatusCode, b)
	}

	return c.notify(ctx)
}

// UpdateGroceryItem saves aisle (and other mutable fields) back to Paprika via the V1 sync
// endpoint. The V2 single-item endpoint (/api/v2/sync/grocery/{uid}/) returns 404 for
// existing items, so we use the same V1 upsert path as SaveGroceryItem.
func (c *Client) UpdateGroceryItem(ctx context.Context, item GroceryItem) error {
	if item.UID == "" {
		item.UID = newUID()
	}

	data, err := json.Marshal([]GroceryItem{item})
	if err != nil {
		return err
	}
	gz, err := gzipBytes(data)
	if err != nil {
		return err
	}

	body, contentType, err := buildMultipartBody(gz)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://paprikaapp.com/api/v1/sync/groceries/", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	req.SetBasicAuth(c.username, c.password)
	req.ContentLength = int64(body.Len())

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("paprika API error %d: %s", resp.StatusCode, b)
	}

	return c.notify(ctx)
}

// DeleteGroceryItem removes a grocery item via the V1 sync endpoint by sending deleted=true.
// This follows the same soft-delete pattern as MealPlanEntry and PantryItem.
// NOTE: Verify this works against the real API — the Paprika V1 sync pattern is consistent
// across meals and pantry, but grocery deletion has not been tested independently.
func (c *Client) DeleteGroceryItem(ctx context.Context, item GroceryItem) error {
	item.Deleted = true

	data, err := json.Marshal([]GroceryItem{item})
	if err != nil {
		return err
	}
	gz, err := gzipBytes(data)
	if err != nil {
		return err
	}

	body, contentType, err := buildMultipartBody(gz)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://paprikaapp.com/api/v1/sync/groceries/", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	req.SetBasicAuth(c.username, c.password)
	req.ContentLength = int64(body.Len())

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("paprika API error %d: %s", resp.StatusCode, b)
	}

	return c.notify(ctx)
}

// MealType constants match Paprika's internal representation.
const (
	MealTypeBreakfast = 0
	MealTypeLunch     = 1
	MealTypeDinner    = 2
	MealTypeSnack     = 3
)

type MealPlanEntry struct {
	UID        string `json:"uid"`
	RecipeUID  string `json:"recipe_uid"`
	RecipeName string `json:"name"`
	Date       string `json:"date"`       // "YYYY-MM-DD 00:00:00" (Paprika's format)
	MealType   int    `json:"type"`       // 0-3; Paprika uses "type" not "meal_type"
	OrderFlag  int    `json:"order_flag"`
	Note       string `json:"note"`
	Deleted    bool   `json:"deleted"`
}

func (c *Client) ListMealPlanEntries(ctx context.Context, start, end time.Time) ([]MealPlanEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://paprikaapp.com/api/v2/sync/meals/", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("paprika API error %d: %s", resp.StatusCode, b)
	}

	var result struct {
		Result []MealPlanEntry `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Filter to the requested date range
	var filtered []MealPlanEntry
	for _, entry := range result.Result {
		t, err := time.Parse("2006-01-02 15:04:05", entry.Date)
		if err != nil {
			// Fallback for bare date format.
			t, err = time.Parse("2006-01-02", entry.Date)
			if err != nil {
				continue
			}
		}
		if !t.Before(start.Truncate(24*time.Hour)) && !t.After(end.Truncate(24*time.Hour)) {
			filtered = append(filtered, entry)
		}
	}
	return filtered, nil
}

// DeleteMealPlanEntry soft-deletes a meal plan entry by setting deleted=true and POSTing
// to the V1 sync endpoint. This follows the same pattern as DeleteGroceryItem.
func (c *Client) DeleteMealPlanEntry(ctx context.Context, entry MealPlanEntry) error {
	entry.Deleted = true

	data, err := json.Marshal([]MealPlanEntry{entry})
	if err != nil {
		return err
	}
	gz, err := gzipBytes(data)
	if err != nil {
		return err
	}

	body, contentType, err := buildMultipartBody(gz)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://paprikaapp.com/api/v1/sync/meals/", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	req.SetBasicAuth(c.username, c.password)
	req.ContentLength = int64(body.Len())

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("paprika API error %d: %s", resp.StatusCode, b)
	}

	return c.notify(ctx)
}

func (c *Client) SaveMealPlanEntry(ctx context.Context, entry MealPlanEntry) error {
	if entry.UID == "" {
		entry.UID = newUID()
	}
	entry.Deleted = false

	// V1 API requires a JSON array, not a single object.
	data, err := json.Marshal([]MealPlanEntry{entry})
	if err != nil {
		return err
	}
	gz, err := gzipBytes(data)
	if err != nil {
		return err
	}

	body, contentType, err := buildMultipartBody(gz)
	if err != nil {
		return err
	}

	// V1 API: plural endpoint, no UID in URL, Basic Auth.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://paprikaapp.com/api/v1/sync/meals/", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	req.SetBasicAuth(c.username, c.password)
	req.ContentLength = int64(body.Len())

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("paprika API error %d: %s", resp.StatusCode, b)
	}

	return c.notify(ctx)
}

type PantryItem struct {
	UID           string  `json:"uid"`
	Ingredient    string  `json:"ingredient"`
	Quantity      string  `json:"quantity"`
	Aisle         string  `json:"aisle"`
	InStock       bool    `json:"in_stock"`
	PurchaseDate  string  `json:"purchase_date"`
	ExpirationDate *string `json:"expiration_date"`
	HasExpiration bool    `json:"has_expiration"`
	Notes         *string `json:"notes"`
	Deleted       bool    `json:"deleted"`
}

func (c *Client) ListPantryItems(ctx context.Context) ([]PantryItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://paprikaapp.com/api/v2/sync/pantry/", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("paprika API error %d: %s", resp.StatusCode, b)
	}

	var result struct {
		Result []PantryItem `json:"result"`
	}
	return result.Result, json.NewDecoder(resp.Body).Decode(&result)
}

func (c *Client) SavePantryItem(ctx context.Context, item PantryItem) error {
	if item.UID == "" {
		item.UID = newUID()
	}
	if item.PurchaseDate == "" {
		item.PurchaseDate = time.Now().UTC().Format("2006-01-02 15:04:05")
	}
	item.Deleted = false

	data, err := json.Marshal([]PantryItem{item})
	if err != nil {
		return err
	}
	gz, err := gzipBytes(data)
	if err != nil {
		return err
	}

	body, contentType, err := buildMultipartBody(gz)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://paprikaapp.com/api/v1/sync/pantry/", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	req.SetBasicAuth(c.username, c.password)
	req.ContentLength = int64(body.Len())

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("paprika API error %d: %s", resp.StatusCode, b)
	}

	return c.notify(ctx)
}

// notify sends a POST to /v2/sync/notify, which tells all Paprika clients to sync.
// We usually defer this call after a recipe is created/updated/deleted, since we don't care whether it suceeds or not.
func (c *Client) notify(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://paprikaapp.com/api/v2/sync/notify", nil)
	if err != nil {
		c.logger.Error("failed to create request", "error", err)
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("failed to notify", "error", err)
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("failed to notify", "status", resp.Status)
		return fmt.Errorf("failed to notify: %s", resp.Status)
	}

	return nil
}

// isErrorResponse checks if the response body contains an error message
// and returns an error if it does. The Paprika API is very inconsistent with how it returns errors;
// sometimes a successful status code can be returned but an error is still returned in the body
func isErrorResponse(body []byte) error {
	var errResp errorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		// Not even valid JSON
		return err
	}

	// Check if it's likely an error response
	if errResp.Error.Message != "" || errResp.Error.Code != 0 {
		return fmt.Errorf("error: %s (code: %d)", errResp.Error.Message, errResp.Error.Code)
	}

	return nil
}

// newUID generates a new uppercase UUID for Paprika items.
func newUID() string {
	return strings.ToUpper(uuid.New().String())
}

// gzipBytes compresses data using gzip and returns the result.
func gzipBytes(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		gz.Close()
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// buildMultipartBody wraps data in a multipart form with a single "data" file field,
// matching the format expected by Paprika's sync endpoints.
// Returns the body buffer, the Content-Type header value, and any error.
func buildMultipartBody(data []byte) (*bytes.Buffer, string, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("data", "data")
	if err != nil {
		return nil, "", err
	}
	if _, err := part.Write(data); err != nil {
		return nil, "", err
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return &body, w.FormDataContentType(), nil
}
