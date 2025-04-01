package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/data/errors"
	u_log "gitlab.calendaria.team/services/utils/v1/log"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"
	"gitlab.calendaria.team/services/utils/v4/config"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/mitchellh/mapstructure"
	xoauth2 "golang.org/x/oauth2"
)

// GoogleOAuthConfig contains Google OAuth configuration
type GoogleOAuthConfig struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURIs []string `json:"redirect_uris"`
	AuthURI      string   `json:"auth_uri"`
	TokenURI     string   `json:"token_uri"`
}

// GoogleOAuthToken represents the OAuth2 token response from Google
type GoogleOAuthToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope,omitempty"`
}

// ToXOAuth2Token converts the Google OAuth token to standard xoauth2.Token
func (t *GoogleOAuthToken) ToXOAuth2Token() *xoauth2.Token {
	if t == nil {
		return nil
	}

	expiry := time.Now().Add(time.Duration(t.ExpiresIn) * time.Second)
	return &xoauth2.Token{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		TokenType:    t.TokenType,
		Expiry:       expiry,
	}
}

// GoogleUserInfo represents user information from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// GoogleOAuthError represents an error response from Google OAuth
type GoogleOAuthError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// GoogleGateway implements IProviderGateway for Google OAuth2
type GoogleGateway struct {
	config      config.IConfig
	log         *log.Helper
	httpClient  *http.Client
	oauthConfig *GoogleOAuthConfig
	userInfoURL string
	scopes      []string
}

const (
	// Google OAuth2 credentials vault key
	googleCredentialsKey = "gwebcredentials"

	// Default Google OAuth endpoints
	googleTokenURI      = "https://oauth2.googleapis.com/token"
	googleRevocationURL = "https://oauth2.googleapis.com/revoke"
	googleUserInfoURI   = "https://www.googleapis.com/oauth2/v2/userinfo"
	defaultRedirectURI  = "https://localhost:8100"

	// Token buffer time defines how long before expiry we should refresh (30 minutes)
	tokenBufferTime = 30 * time.Minute

	// HTTP Client timeout
	httpClientTimeout = 10 * time.Second
)

// NewGoogleRemote creates a new Google OAuth2 gateway
// It loads Google credentials immediately to validate configuration up front
func NewGoogleRemote(config config.IConfig) (IProviderGateway, error) {
	logger := u_log.NewStdLogger()
	logHelper := log.NewHelper(log.With(logger, "ts", log.DefaultTimestamp, "component", "integration.google"))

	g := &GoogleGateway{
		config:      config,
		log:         logHelper,
		httpClient:  &http.Client{Timeout: httpClientTimeout},
		userInfoURL: googleUserInfoURI,
		scopes:      []string{"profile", "email", "https://www.googleapis.com/auth/calendar"},
	}

	// Load Google credentials upfront to validate configuration
	ctx := context.Background()

	// Get Google credentials from secure storage
	mapGoogleCredentials, err := g.config.ReadGlobalSecretsFor(ctx, googleCredentialsKey)
	if err != nil {
		g.log.Errorf("Failed to read client secret from vault: %v", err)
		return nil, fmt.Errorf("unable to read client secret: %w", err)
	}

	// Decode Google credentials
	var googleCredentialsJSON string
	err = mapstructure.Decode(mapGoogleCredentials["data"], &googleCredentialsJSON)
	if err != nil {
		g.log.Errorf("Failed to decode client secret: %v", err)
		return nil, fmt.Errorf("unable to decode client secret: %w", err)
	}

	// Validate credential format
	if len(googleCredentialsJSON) < 10 {
		return nil, errors.ErrConfigNotFound
	}

	// Parse credentials from JSON
	var webCredentials struct {
		Web GoogleOAuthConfig `json:"web"`
	}

	if err2 := json.Unmarshal([]byte(googleCredentialsJSON), &webCredentials); err2 != nil {
		// Try alternate format (installed apps)
		var installedCredentials struct {
			Installed GoogleOAuthConfig `json:"installed"`
		}
		if err3 := json.Unmarshal([]byte(googleCredentialsJSON), &installedCredentials); err3 != nil {
			g.log.Errorf("Failed to parse client secret JSON: %v", err3)
			return nil, fmt.Errorf("unable to parse client secret: %w", err3)
		}
		g.oauthConfig = &installedCredentials.Installed
	} else {
		g.oauthConfig = &webCredentials.Web
	}

	// Set default token URI if not present
	if g.oauthConfig.TokenURI == "" {
		g.oauthConfig.TokenURI = googleTokenURI
	}

	g.log.Infof("Google OAuth config loaded successfully (client_id_prefix: %s...)",
		g.oauthConfig.ClientID[:8])

	return g, nil
}

// Authenticate exchanges an authorization code for OAuth tokens and user info
func (r *GoogleGateway) Authenticate(actorID int64, authCode string) (*CredentialDto, error) {
	// Input validation
	if authCode == "" {
		return nil, errors.ErrAuthCodeInvalid
	}

	// Ensure we have a valid OAuth config
	if r.oauthConfig == nil {
		return nil, errors.ErrConfigNotFound
	}

	// Exchange auth code for token with direct HTTP request
	token, err := r.exchangeAuthCodeForToken(authCode)
	if err != nil {
		r.log.Errorf("Failed to exchange auth code for token: %v (actor_id: %d)", err, actorID)
		return nil, errors.MapToExternalError(err)
	}

	// Validate received token
	if token.AccessToken == "" {
		r.log.Errorf("Received empty access token (actor_id: %d)", actorID)
		return nil, errors.ErrInvalidToken
	}

	// Validate refresh token - critical for offline access
	if token.RefreshToken == "" {
		r.log.Warnf("No refresh token received, offline access may be limited (actor_id: %d)", actorID)
	}

	// Get user information using the access token
	userInfo, err := r.getUserInfo(token.AccessToken)
	if err != nil {
		r.log.Errorf("Failed to retrieve user info: %v (actor_id: %d)", err, actorID)
		return nil, errors.MapToExternalError(err)
	}

	// Convert to our token format
	xOAuthToken := token.ToXOAuth2Token()

	// Validate the token
	if !xOAuthToken.Valid() {
		r.log.Errorf("Invalid token received (actor_id: %d)", actorID)
		return nil, errors.ErrInvalidToken
	}

	// Create credential DTO with token and user info
	dto := &CredentialDto{
		UserID:      actorID,
		DisplayName: userInfo.Name,
		Email:       userInfo.Email,
		Provider:    u_struc.Google,
		Token:       xOAuthToken,
	}

	return dto, nil
}

// exchangeAuthCodeForToken exchanges authorization code for an OAuth token
func (r *GoogleGateway) exchangeAuthCodeForToken(
	authCode string,
) (*GoogleOAuthToken, error) {
	// Select a redirect URI - use the first one from the array
	var redirectURI string
	if len(r.oauthConfig.RedirectURIs) > 0 {
		redirectURI = r.oauthConfig.RedirectURIs[0]
	} else {
		// Default fallback - though this might not work if not registered
		redirectURI = defaultRedirectURI
	}

	// Prepare form data
	data := url.Values{}
	data.Set("code", authCode)
	data.Set("client_id", r.oauthConfig.ClientID)
	data.Set("client_secret", r.oauthConfig.ClientSecret)
	data.Set("redirect_uri", redirectURI) // FIXED: Using single redirect_uri
	data.Set("grant_type", "authorization_code")

	// Critical parameters for offline access
	data.Set("access_type", "offline")
	data.Set("prompt", "consent")

	// Add scopes
	data.Set("scope", strings.Join(r.scopes, " "))

	// Create request
	req, err := http.NewRequest(http.MethodPost, r.oauthConfig.TokenURI, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		// Log full response for debugging
		r.log.Debugf("Token exchange failed - Response: %s", string(body))

		var oauthErr GoogleOAuthError
		err = json.Unmarshal(body, &oauthErr)
		if err == nil && oauthErr.Error != "" {
			r.log.Errorf("OAuth error response: %s - %s", oauthErr.Error, oauthErr.ErrorDescription)

			if oauthErr.Error == "invalid_grant" {
				return nil, errors.ErrInvalidGrant
			}

			if oauthErr.Error == "redirect_uri_mismatch" {
				r.log.Errorf("Redirect URI mismatch - used: %s", redirectURI)
				return nil, fmt.Errorf("redirect URI mismatch, used: %s", redirectURI)
			}

			return nil, fmt.Errorf("oauth error: %s - %s", oauthErr.Error, oauthErr.ErrorDescription)
		}

		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the successful response
	var token GoogleOAuthToken
	err = json.Unmarshal(body, &token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &token, nil
}

// getUserInfo retrieves the user's profile information using an access token
func (r *GoogleGateway) getUserInfo(accessToken string) (*GoogleUserInfo, error) {
	// Create request
	req, err := http.NewRequest(http.MethodGet, r.userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user info request: %w", err)
	}

	// Set authorization header
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("user info request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user info response: %w", err)
	}

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		r.log.Errorf("User info request failed with status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("user info request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the successful response
	var userInfo GoogleUserInfo
	err = json.Unmarshal(body, &userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user info response: %w", err)
	}

	return &userInfo, nil
}

// RefreshToken refreshes an expired OAuth token
func (r *GoogleGateway) RefreshToken(
	credential *ent.UserCredentials,
) (*CredentialDto, error) {
	// Input validation
	if credential == nil || credential.Provider == nil || *credential.Provider != u_struc.Google {
		return nil, errors.ErrInvalidCredential
	}

	// Convert DB credential to DTO
	dto := CredentialToDto(credential)

	// Check if token is invalid
	if dto == nil || dto.Token == nil {
		return nil, errors.ErrInvalidToken
	}

	// Check for refresh token existence (required for token refresh)
	if dto.Token.RefreshToken == "" {
		r.log.Warnf("Missing refresh token, re-authentication required (user_id: %d, credential_id: %d)",
			credential.UserID, credential.ID)
		return nil, errors.ErrMissingRefreshToken
	}

	// RefreshToken obtains a new access token when the current one is nearing expiration.
	// It uses a 30-minute buffer time to proactively refresh tokens before they expire,
	// preventing disruptions during user sessions.
	now := time.Now()
	if !dto.Token.Expiry.IsZero() && dto.Token.Expiry.After(now.Add(tokenBufferTime)) {
		r.log.Debugf("Token still valid, no refresh needed (user_id: %d, expires_at: %s, now: %s)",
			credential.UserID, dto.Token.Expiry.Format(time.RFC3339), now.Format(time.RFC3339))
		return dto, nil
	}

	// Ensure we have a valid OAuth config
	if r.oauthConfig == nil {
		return nil, errors.ErrConfigNotFound
	}

	// Refresh the token with direct HTTP request
	newToken, err := r.refreshTokenRequest(r.oauthConfig, dto.Token.RefreshToken)
	if err != nil {
		r.log.Errorf("Token refresh failed: %v (user_id: %d, credential_id: %d)",
			err, credential.UserID, credential.ID)
		return nil, errors.MapToExternalError(err)
	}

	// Convert to our token format
	xOAuthToken := newToken.ToXOAuth2Token()

	// Validate the token
	if !xOAuthToken.Valid() {
		r.log.Errorf("Token validation failed after refresh (user_id: %d)",
			credential.UserID)
		return nil, errors.ErrInvalidToken
	}

	// Preserve the refresh token if not returned in the response
	// Google doesn't always return a new refresh token during refresh operations
	if xOAuthToken.RefreshToken == "" {
		xOAuthToken.RefreshToken = dto.Token.RefreshToken
	}

	// Update the DTO with the new token
	dto.Token = xOAuthToken

	// Log successful refresh
	r.log.Debugf("Token refreshed successfully (user_id: %d, credential_id: %d, new_expiry: %s)",
		credential.UserID, credential.ID, xOAuthToken.Expiry.Format(time.RFC3339))

	return dto, nil
}

// refreshTokenRequest sends a refresh token request to Google OAuth server
func (r *GoogleGateway) refreshTokenRequest(
	googleConfig *GoogleOAuthConfig,
	refreshToken string,
) (*GoogleOAuthToken, error) {
	// Prepare form data
	data := url.Values{}
	data.Set("client_id", googleConfig.ClientID)
	data.Set("client_secret", googleConfig.ClientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	// Create request
	req, err := http.NewRequest(http.MethodPost, googleConfig.TokenURI, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token request: %w", err)
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh token request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh token response: %w", err)
	}

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		var oauthErr GoogleOAuthError
		err = json.Unmarshal(body, &oauthErr)
		if err == nil && oauthErr.Error != "" {
			r.log.Errorf("OAuth refresh error: %s - %s", oauthErr.Error, oauthErr.ErrorDescription)

			if oauthErr.Error == "invalid_grant" {
				return nil, errors.ErrInvalidGrant
			}

			return nil, fmt.Errorf("oauth error: %s - %s", oauthErr.Error, oauthErr.ErrorDescription)
		}

		return nil, fmt.Errorf("refresh token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the successful response
	var token GoogleOAuthToken
	err = json.Unmarshal(body, &token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token response: %w", err)
	}

	return &token, nil
}

// RevokeToken revokes both access and refresh tokens for a credential
// This ensures the credential is completely invalidated with Google
func (r *GoogleGateway) RevokeToken(credential *ent.UserCredentials) error {
	if credential == nil {
		return errors.ErrInvalidCredential
	}

	var errs []error

	// Try to revoke refresh token first (this also invalidates all associated access tokens)
	if credential.RefreshToken != nil && *credential.RefreshToken != "" {
		err := r.revokeTokenRequest(*credential.RefreshToken, true)
		if err != nil {
			r.log.Warnf("Failed to revoke refresh token: %v (user_id: %d, credential_id: %d)",
				err, credential.UserID, credential.ID)
			errs = append(errs, err)

			// Continue to try revoking access token even if refresh token revocation fails
		} else {
			// If refresh token revocation was successful, we don't need to revoke the access token
			// as it's automatically invalidated

			mail := "-"
			if credential.Mail != nil {
				mail = *credential.Mail
			}
			r.log.Infof("Revoked [Refresh token including Access token] successfully "+
				"for userID (%v) with mail (%s)", credential.UserID, mail)

			return nil
		}
	}

	// If refresh token revocation failed or there was no refresh token,
	// try to revoke the access token as a fallback
	if credential.AccessToken != "" {
		err := r.revokeTokenRequest(credential.AccessToken, false)
		if err != nil {
			r.log.Warnf("Failed to revoke access token: %v (user_id: %d, credential_id: %d)",
				err, credential.UserID, credential.ID)
			errs = append(errs, err)
		}
	}

	// If we have errors from both revocation attempts, return a combined error
	if len(errs) > 0 {
		// Format combined error message
		var errMsgs []string
		for _, e := range errs {
			errMsgs = append(errMsgs, e.Error())
		}
		return fmt.Errorf("token revocation failed: %s", strings.Join(errMsgs, "; "))
	}

	mail := "-"
	if credential.Mail != nil {
		mail = *credential.Mail
	}
	r.log.Infof(
		"Revoked [only Access token] successfully for userID (%v) with mail (%s)",
		credential.UserID,
		mail,
	)

	return nil
}

// revokeTokenRequest revokes an OAuth token with Google
// It supports revoking both access tokens and refresh tokens
// When a refresh token is revoked, all associated access tokens are automatically invalidated
func (r *GoogleGateway) revokeTokenRequest(token string, isRefreshToken bool) error {
	// Validate input
	if token == "" {
		return errors.ErrInvalidToken
	}

	// Prepare form data
	data := url.Values{}
	data.Set("token", token)

	// Optionally specify token type hint to optimize revocation on Google's servers
	if isRefreshToken {
		data.Set("token_type_hint", "refresh_token")
	} else {
		data.Set("token_type_hint", "access_token")
	}

	// Create request
	req, err := http.NewRequest(http.MethodPost, googleRevocationURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create revocation request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("revocation request failed: %w", err)
	}
	defer resp.Body.Close()

	// Google returns 200 OK for successful revocation
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		r.log.Errorf("Token revocation failed with status %d: %s", resp.StatusCode, string(body))

		// Try to parse error response
		var oauthErr GoogleOAuthError
		err = json.Unmarshal(body, &oauthErr)
		if err == nil && oauthErr.Error != "" {
			r.log.Errorf("OAuth revocation error: %s - %s", oauthErr.Error, oauthErr.ErrorDescription)
			return fmt.Errorf("oauth error: %s - %s", oauthErr.Error, oauthErr.ErrorDescription)
		}

		return fmt.Errorf("token revocation failed with status %d", resp.StatusCode)
	}

	return nil
}
