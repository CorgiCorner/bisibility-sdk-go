package bisibility

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testMigrationToken = "mig_secret_do_not_log"

func TestGetCloudImportCompatibility(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.Method, http.MethodGet)
		assertEqual(t, r.URL.EscapedPath(), "/api/v1/cloud/import/compatibility")
		// Compatibility is an unauthenticated preflight: no credentials are sent
		// even though the client is configured with an API key.
		assertEqual(t, r.Header.Get("Authorization"), "")
		writeJSON(t, w, http.StatusOK, map[string]any{
			"app_version":               "1.4.0",
			"latest_migration":          "2026-07-01",
			"schema_versions_supported": []any{1, 2, 3},
		})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	response, err := client.GetCloudImportCompatibility(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, response.AppVersion, "1.4.0")
	assertEqual(t, *response.LatestMigration, "2026-07-01")
	assertEqual(t, len(response.SchemaVersionsSupported), 3)
	assertEqual(t, response.SchemaVersionsSupported[2], 3)
}

func TestImportCloudExport(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured := captureRequest(t, r)
		assertEqual(t, captured.Method, http.MethodPost)
		assertEqual(t, captured.Path, "/api/v1/cloud/import")
		assertEqual(t, captured.Header.Get("Authorization"), "Bearer "+testMigrationToken)
		assertEqual(t, captured.Header.Get(contentTypeHeader), "application/json")
		assertJSONEqual(t, captured.Body, `{"version":3,"keywords":[{"keyword":"rank tracker"}]}`)
		writeJSON(t, w, http.StatusCreated, map[string]any{
			"counts": map[string]any{"keywords": 1},
			"job_id": "job_1",
			"state":  "done",
		})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	response, err := client.ImportCloudExport(context.Background(), testMigrationToken, CloudImportPackage{
		Version:  3,
		Keywords: []CloudImportKeyword{{Keyword: "rank tracker"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, response.State, "done")
	assertEqual(t, response.JobID, "job_1")
	assertEqual(t, response.Counts["keywords"], 1)
}

func TestImportCloudExportRequiresToken(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.com/api/v1")
	_, err := client.ImportCloudExport(context.Background(), "  ", CloudImportPackage{Version: 3})
	if err == nil {
		t.Fatal("expected error for empty migration token")
	}
}

func TestCreateCloudImportSession(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured := captureRequest(t, r)
		assertEqual(t, captured.Method, http.MethodPost)
		assertEqual(t, captured.Path, "/api/v1/cloud/import/sessions")
		assertEqual(t, captured.Header.Get("Authorization"), "Bearer "+testMigrationToken)
		assertJSONEqual(t, captured.Body, `{"version":3,"chunk_count":2,"totals":{"keywords":10,"rank_checks":100}}`)
		writeJSON(t, w, http.StatusCreated, map[string]any{
			"session_id": "sess_1",
			"state":      "receiving",
			"chunk_limits": map[string]any{
				"max_body_bytes":   1048576,
				"max_history_rows": 5000,
				"max_keywords":     500,
			},
		})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	response, err := client.CreateCloudImportSession(context.Background(), testMigrationToken, CloudImportSessionCreate{
		ChunkCount: 2,
		Version:    3,
		Totals:     &CloudImportSessionTotals{Keywords: 10, RankChecks: 100},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, response.SessionID, "sess_1")
	assertEqual(t, response.State, "receiving")
	assertEqual(t, response.ChunkLimits.MaxKeywords, 500)
	assertEqual(t, response.ChunkLimits.MaxBodyBytes, 1048576)
}

func TestUploadCloudImportChunk(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured := captureRequest(t, r)
		assertEqual(t, captured.Method, http.MethodPut)
		assertEqual(t, captured.Path, "/api/v1/cloud/import/sessions/sess%201/chunks/0")
		assertEqual(t, captured.Header.Get("Authorization"), "Bearer "+testMigrationToken)
		assertEqual(t, captured.Header.Get(contentTypeHeader), "application/json")
		assertJSONEqual(t, captured.Body, `{"checksum":"sha256:`+
			`0000000000000000000000000000000000000000000000000000000000000000",`+
			`"kind":"keywords","keywords":[{"keyword":"rank tracker"}]}`)
		writeJSON(t, w, http.StatusOK, map[string]any{
			"state":           "receiving",
			"chunks_received": 1,
			"chunk_count":     2,
		})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	response, err := client.UploadCloudImportChunk(context.Background(), testMigrationToken, "sess 1", 0, CloudImportUploadChunk{
		Checksum: "sha256:0000000000000000000000000000000000000000000000000000000000000000",
		Kind:     "keywords",
		Keywords: []CloudImportKeyword{{Keyword: "rank tracker"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, response.State, "receiving")
	assertEqual(t, response.ChunksReceived, 1)
	assertEqual(t, response.ChunkCount, 2)
}

func TestUploadCloudImportChunkRawGzip(t *testing.T) {
	t.Parallel()

	const payload = `{"checksum":"sha256:aa","kind":"sections","sections":{}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured := captureRequest(t, r)
		assertEqual(t, captured.Method, http.MethodPut)
		assertEqual(t, captured.Path, "/api/v1/cloud/import/sessions/sess_1/chunks/3")
		assertEqual(t, captured.Header.Get("Authorization"), "Bearer "+testMigrationToken)
		assertEqual(t, captured.Header.Get(contentTypeHeader), "application/json")
		assertEqual(t, captured.Header.Get("Content-Encoding"), "gzip")
		assertEqual(t, captured.Body, payload)
		writeJSON(t, w, http.StatusOK, map[string]any{
			"state":           "receiving",
			"chunks_received": 4,
			"chunk_count":     4,
		})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	response, err := client.UploadCloudImportChunkRaw(context.Background(), testMigrationToken, "sess_1", 3, bytes.NewBufferString(payload), true)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, response.ChunksReceived, 4)
}

func TestUploadCloudImportChunkValidatesIndex(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.com/api/v1")
	_, err := client.UploadCloudImportChunk(context.Background(), testMigrationToken, "sess_1", -1, CloudImportUploadChunk{Kind: "keywords"})
	if err == nil {
		t.Fatal("expected error for negative chunk index")
	}
}

func TestFinalizeCloudImportSession(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured := captureRequest(t, r)
		assertEqual(t, captured.Method, http.MethodPost)
		assertEqual(t, captured.Path, "/api/v1/cloud/import/sessions/sess_1/finalize")
		assertEqual(t, captured.Header.Get("Authorization"), "Bearer "+testMigrationToken)
		writeJSON(t, w, http.StatusOK, map[string]any{
			"counts": map[string]any{"keywords": 10, "rank_checks": 100},
			"job_id": "job_9",
			"state":  "done",
		})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	response, err := client.FinalizeCloudImportSession(context.Background(), testMigrationToken, "sess_1")
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, response.State, "done")
	assertEqual(t, response.JobID, "job_9")
	assertEqual(t, response.Counts["rank_checks"], 100)
}

func TestFinalizeCloudImportSessionRequiresSession(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.com/api/v1")
	_, err := client.FinalizeCloudImportSession(context.Background(), testMigrationToken, "")
	if err == nil {
		t.Fatal("expected error for empty session id")
	}
}
