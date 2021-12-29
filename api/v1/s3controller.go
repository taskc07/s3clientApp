package v1

import (
	"github.com/khan1507017/s3clientApp/helper"
	"github.com/labstack/echo/v4"
)

type S3controllerInf interface {
	CreateFiles(e echo.Context) error
	CreateDirs(e echo.Context) error
}

type S3controllerInstance struct {
}

func S3Controller() S3controllerInf {
	return new(S3controllerInstance)
}
func (s S3controllerInstance) CreateFiles(e echo.Context) error {
	return helper.CreateFiles().Execute(e)
}

func (s S3controllerInstance) CreateDirs(e echo.Context) error {
	return helper.CreateDirs().Execute(e)
}
