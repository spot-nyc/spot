package auth

import (
	"fmt"
	"strings"

	"golang.org/x/oauth2"
)

// The following constants use placeholder values in M1b. M1c replaces them
// with real production values once the Supabase OAuth client is registered.

// SupabaseProjectRef is the subdomain of the Supabase project's auth host,
// e.g. if the project URL is https://abcd1234.supabase.co this is "abcd1234".
var SupabaseProjectRef = "iotmomhesfkqaktldtxx"

// ClientID is the OAuth 2.1 public client identifier assigned by Supabase when
// the Spot CLI was registered as an OAuth client.
var ClientID = "60e5a5cb-00a0-42bd-b7b2-6ff112656285"

// DefaultScopes is the default space-separated scope string requested at login.
//
// We request "phone profile" (not "openid ..."), because the Supabase project
// currently uses HS256 signing. The openid scope would fail to produce an ID
// token under HS256 — access tokens work regardless. The Spot API's JWT
// middleware validates the access token on the way in; we use its /users/me
// endpoint to surface user info.
const DefaultScopes = "phone profile"

// LoopbackPorts is the list of pre-registered loopback ports the CLI tries in
// order when starting its local callback server. Supabase requires exact
// redirect URI matches (no wildcards, no RFC 8252 loopback exception), so all
// of these must be registered as redirect URIs on the OAuth client.
//
// Registered URIs (20 total — "127.0.0.1" and "localhost" variants for each port):
//
//	http://127.0.0.1:52853/callback .. http://127.0.0.1:52862/callback
//	http://localhost:52853/callback .. http://localhost:52862/callback
var LoopbackPorts = []int{52853, 52854, 52855, 52856, 52857, 52858, 52859, 52860, 52861, 52862}

// AuthorizeURL builds the authorize endpoint URL.
func AuthorizeURL() string {
	return fmt.Sprintf("https://%s.supabase.co/auth/v1/oauth/authorize", SupabaseProjectRef)
}

// TokenURL builds the token endpoint URL.
func TokenURL() string {
	return fmt.Sprintf("https://%s.supabase.co/auth/v1/oauth/token", SupabaseProjectRef)
}

// DefaultOAuthConfig returns an oauth2.Config for the Spot OAuth server,
// wired to production endpoints and the default scopes.
func DefaultOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID: ClientID,
		Scopes:   strings.Fields(DefaultScopes),
		Endpoint: oauth2.Endpoint{
			AuthURL:  AuthorizeURL(),
			TokenURL: TokenURL(),
		},
	}
}
