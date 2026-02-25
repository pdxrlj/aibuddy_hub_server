// generate generates query code for GORM models.
package main

import (
	"aibuddy/internal/model"
	"aibuddy/pkg/config"
)

func main() {
	config.Setup()
	model.Conn().GenerateQuery()
}
