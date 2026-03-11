package observe

import (
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/gorm"
)

func WithGormTracing(db *gorm.DB, dbName string) error {
	return db.Use(otelgorm.NewPlugin(
		otelgorm.WithDBName(dbName),
	))
}
