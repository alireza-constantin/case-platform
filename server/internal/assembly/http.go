package assembly

import (
	"github.com/alireza-constantin/case-platform/server/internal/cases/phone"
	"github.com/alireza-constantin/case-platform/server/internal/cases/terminal"
	"net/http"

	"github.com/alireza-constantin/case-platform/server/internal/kernel"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Handler(db *pgxpool.Pool, origin string, secureCookie bool) http.Handler {
	return kernel.NewHandler(db, origin, secureCookie, []kernel.Case{phone.Module{}, terminal.Module{}})
}
