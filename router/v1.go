package router

import (
	v1 "github.com/khan1507017/s3clientApp/api/v1"
	"github.com/labstack/echo/v4"
)

func v1Apis(group *echo.Group) {

	//s3clientApis
	s3 := group.Group("/s3")
	s3.POST("/files", v1.S3Controller().CreateFiles)
	s3.POST("/dirs", v1.S3Controller().CreateDirs)

	//OtherDb apis -> future extend
}
