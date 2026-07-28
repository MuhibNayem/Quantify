package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type stubUploader struct {
	exists bool
	err    error
}

func (s *stubUploader) UploadFile(string, string, *bytes.Buffer, int64) (minio.UploadInfo, error) {
	return minio.UploadInfo{}, nil
}
func (s *stubUploader) UploadFileFromMultipart(string, string, *multipart.FileHeader) (minio.UploadInfo, error) {
	return minio.UploadInfo{}, nil
}
func (s *stubUploader) DownloadFile(string, string) (*minio.Object, error) { return nil, nil }
func (s *stubUploader) GetFile(string, string) (*minio.Object, error)      { return nil, nil }
func (s *stubUploader) BucketExists(context.Context, string) (bool, error) { return s.exists, s.err }

func serveReady(t *testing.T, h *ReadyHandler) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ready", h.Ready)
	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func failedDeps(t *testing.T, w *httptest.ResponseRecorder) []string {
	t.Helper()
	var body struct {
		Status string   `json:"status"`
		Failed []string `json:"failed"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	return body.Failed
}

func newReadyTestRedis(t *testing.T) redis.UniversalClient {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	return rdb
}

func newClosedDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
	return db
}

func TestReady_FailsWhenDependencyDown(t *testing.T) {
	h := &ReadyHandler{
		db:       newClosedDB(t),
		redis:    newReadyTestRedis(t),
		amqp:     nil,
		uploader: &stubUploader{exists: false},
	}

	w := serveReady(t, h)

	assert.Equal(t, 503, w.Code)
	failed := failedDeps(t, w)
	assert.Contains(t, failed, "db")
	assert.Contains(t, failed, "amqp")
	assert.Contains(t, failed, "minio")
	assert.NotContains(t, failed, "redis", "healthy redis must not be reported as failed")
}

func TestReady_ReportsRedisDown(t *testing.T) {
	rdb := newReadyTestRedis(t)
	_ = rdb.Close()

	h := &ReadyHandler{
		db:       newClosedDB(t),
		redis:    rdb,
		amqp:     nil,
		uploader: &stubUploader{exists: true},
	}

	w := serveReady(t, h)

	assert.Equal(t, 503, w.Code)
	assert.Contains(t, failedDeps(t, w), "redis")
}

func TestReady_NilDependenciesDoNotPanic(t *testing.T) {
	h := &ReadyHandler{}

	w := serveReady(t, h)

	assert.Equal(t, 503, w.Code)
	failed := failedDeps(t, w)
	assert.ElementsMatch(t, []string{"db", "redis", "amqp", "minio"}, failed)
}
