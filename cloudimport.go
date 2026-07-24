package bisibility

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const cloudImportPath = "/cloud/import"

// applyMigrationToken authenticates a cloud-import request with a migration
// token. Cloud-import writes use a migration token (Authorization: Bearer
// mig_...) rather than the client's API key, so key-based auth is disabled and
// the token is attached as the request bearer credential.
func applyMigrationToken(config *requestConfig, migrationToken string) {
	config.auth = false
	config.migrationToken = migrationToken
}

// GetCloudImportCompatibility checks cloud-import schema compatibility. It is
// an unauthenticated preflight that reports which package schema versions the
// server accepts before a migration begins.
func (c *Client) GetCloudImportCompatibility(ctx context.Context, options ...RequestOption) (*CloudImportCompatibility, error) {
	config := newRequestConfig(options...)
	config.auth = false
	return requestJSON[CloudImportCompatibility](c, ctx, http.MethodGet, cloudImportPath+"/compatibility", config)
}

// ImportCloudExport imports an export package in a single request, authenticated
// with a migration token minted by MintMigrationToken. The API responds 201
// with the completed import counts.
func (c *Client) ImportCloudExport(ctx context.Context, migrationToken string, pkg CloudImportPackage, options ...RequestOption) (*CloudImportFinalizeResponse, error) {
	if strings.TrimSpace(migrationToken) == "" {
		return nil, &ConfigurationError{Message: "ImportCloudExport requires a migration token."}
	}

	config := newRequestConfig(options...)
	applyMigrationToken(&config, migrationToken)
	config.body = pkg
	return requestJSON[CloudImportFinalizeResponse](c, ctx, http.MethodPost, cloudImportPath, config)
}

// CreateCloudImportSession opens a chunked cloud-import session, authenticated
// with a migration token. Upload each chunk with UploadCloudImportChunk, then
// complete the import with FinalizeCloudImportSession.
func (c *Client) CreateCloudImportSession(ctx context.Context, migrationToken string, input CloudImportSessionCreate, options ...RequestOption) (*CloudImportSessionCreateResponse, error) {
	if strings.TrimSpace(migrationToken) == "" {
		return nil, &ConfigurationError{Message: "CreateCloudImportSession requires a migration token."}
	}

	config := newRequestConfig(options...)
	applyMigrationToken(&config, migrationToken)
	config.body = input
	return requestJSON[CloudImportSessionCreateResponse](c, ctx, http.MethodPost, cloudImportPath+"/sessions", config)
}

// UploadCloudImportChunk uploads one JSON chunk to a chunked import session,
// authenticated with a migration token. index is the zero-based chunk position.
// The chunk is sent as an application/json body.
func (c *Client) UploadCloudImportChunk(ctx context.Context, migrationToken, sessionID string, index int, chunk CloudImportUploadChunk, options ...RequestOption) (*CloudImportChunkResponse, error) {
	if strings.TrimSpace(migrationToken) == "" {
		return nil, &ConfigurationError{Message: "UploadCloudImportChunk requires a migration token."}
	}
	if strings.TrimSpace(sessionID) == "" {
		return nil, &ConfigurationError{Message: "UploadCloudImportChunk requires a session id."}
	}
	if index < 0 {
		return nil, &ConfigurationError{Message: "UploadCloudImportChunk index cannot be negative."}
	}

	config := newRequestConfig(options...)
	applyMigrationToken(&config, migrationToken)
	config.body = chunk
	return requestJSON[CloudImportChunkResponse](c, ctx, http.MethodPut, cloudImportChunkPath(sessionID, index), config)
}

// UploadCloudImportChunkRaw uploads one pre-serialized JSON chunk body to a
// chunked import session. Use it to stream a chunk from an io.Reader or to send
// a gzip-compressed body: pass the raw application/json bytes as body, and set
// gzip to true when body is gzip compressed so the server receives the
// Content-Encoding: gzip header. The body must still decode to a
// CloudImportUploadChunk.
func (c *Client) UploadCloudImportChunkRaw(ctx context.Context, migrationToken, sessionID string, index int, body io.Reader, gzip bool, options ...RequestOption) (*CloudImportChunkResponse, error) {
	if strings.TrimSpace(migrationToken) == "" {
		return nil, &ConfigurationError{Message: "UploadCloudImportChunkRaw requires a migration token."}
	}
	if strings.TrimSpace(sessionID) == "" {
		return nil, &ConfigurationError{Message: "UploadCloudImportChunkRaw requires a session id."}
	}
	if index < 0 {
		return nil, &ConfigurationError{Message: "UploadCloudImportChunkRaw index cannot be negative."}
	}
	if body == nil {
		return nil, &ConfigurationError{Message: "UploadCloudImportChunkRaw requires a body."}
	}

	config := newRequestConfig(options...)
	applyMigrationToken(&config, migrationToken)
	config.rawBody = body
	config.headers.Set(contentTypeHeader, "application/json")
	if gzip {
		config.headers.Set("Content-Encoding", "gzip")
	}
	return requestJSON[CloudImportChunkResponse](c, ctx, http.MethodPut, cloudImportChunkPath(sessionID, index), config)
}

// FinalizeCloudImportSession finalizes a chunked import session after every
// chunk has been uploaded, authenticated with a migration token. The API
// responds 200 with the completed import counts.
func (c *Client) FinalizeCloudImportSession(ctx context.Context, migrationToken, sessionID string, options ...RequestOption) (*CloudImportFinalizeResponse, error) {
	if strings.TrimSpace(migrationToken) == "" {
		return nil, &ConfigurationError{Message: "FinalizeCloudImportSession requires a migration token."}
	}
	if strings.TrimSpace(sessionID) == "" {
		return nil, &ConfigurationError{Message: "FinalizeCloudImportSession requires a session id."}
	}

	config := newRequestConfig(options...)
	applyMigrationToken(&config, migrationToken)
	path := cloudImportPath + "/sessions/" + url.PathEscape(sessionID) + "/finalize"
	return requestJSON[CloudImportFinalizeResponse](c, ctx, http.MethodPost, path, config)
}

func cloudImportChunkPath(sessionID string, index int) string {
	return cloudImportPath + "/sessions/" + url.PathEscape(sessionID) + "/chunks/" + strconv.Itoa(index)
}
