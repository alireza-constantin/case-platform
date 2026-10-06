package assembly

import (
	"net/http"

	"github.com/alireza-constantin/case-platform/server/internal/kernel"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Handler(db *pgxpool.Pool, origin string, secureCookie bool) http.Handler {
	// Register first-party implementations here when their slices are delivered.
	return kernel.NewHandler(db, origin, secureCookie, []kernel.CaseMetadata{})
}
