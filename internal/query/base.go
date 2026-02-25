// Package query provides generated query code for database models.
package query

import "aibuddy/internal/model"

func init() {
	SetDefault(model.Conn().GetDB())
}
