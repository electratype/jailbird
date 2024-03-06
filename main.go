package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "electratype/jailbird/docs"

	"electratype/jailbird/models"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var DB *gorm.DB

func MigrateDatabase() {

	//DB.Exec("CREATE OR REPLACE FUNCTION get_film_count(len_from int, len_to int)")

	DB.AutoMigrate(&models.Project{}, &models.ProgressItem{}, &models.ApiKey{}, &models.User{})
}

// @Summary Return all projects
// @Schemes
// @Description Returns array of all projects
// @Tags project
// @Accept json
// @Produce json
// @Success 200 {array} models.Project
// @Router /projects [get]
func ListProjects(c *gin.Context) {

	var project []models.Project
	DB.Find(&project)

	c.JSON(http.StatusOK, &project)
}

// @Summary Delete project
// @Tags project
// @Param id path string true "Project ID"
// @Success 200
// @Router /projects/{id} [delete]
func DeleteProject(c *gin.Context) {
	id := c.Param("projectId")

	DB.Where("slug = ?", id).Delete(&models.Project{})

	c.JSON(204, "")
}

// @Summary Create new project
// @Tags project
// @Param body body models.PlainProject true "Project definition"
// @Accept json
// @Produce json
// @Success 200
// @Router /projects [post]
func AddProject(c *gin.Context) {

	var plainProject models.PlainProject
	var project models.Project

	if err := c.ShouldBindJSON(&plainProject); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "failure", "error": err.Error()})
	}

	log.Printf("%+v\n", &plainProject)

	project.Slug = plainProject.Slug
	project.Name = plainProject.Name
	project.Description = plainProject.Description

	result := DB.Create(&project)

	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		c.Error(result.Error)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "failure", "error": result.Error.Error()})
	}

	c.JSON(204, "")
}

func ListItems(c *gin.Context) {

}

// @title           JailBird API
// @version         1.0
// @description     This a progress management system API.

// @contact.name   API support and issue management
// @contact.url    https://electratype.com
// @contact.email  support@electratype.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:5454
// @BasePath  /api/v1
func main() {

	DSN := os.Getenv("DSN")
	if DSN == "" {
		log.Fatal("DSN not set! Terminating!")
	}
	log.Println("DSN is set to", DSN)

	router := gin.Default()

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			projects := v1.Group("/projects")
			{
				projects.GET("", ListProjects)
				projects.POST("", AddProject)
				projects.DELETE(":projectId", DeleteProject)
				items := projects.Group(":projectId/items")
				{
					items.GET("", ListItems)
				}
			}

		}
	}
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	var err error
	DB, err = gorm.Open(postgres.Open(DSN), &gorm.Config{TranslateError: true})
	if err != nil {
		panic("failed to connect database")
	}

	MigrateDatabase()

	router.Run("localhost:5454")

	log.Println("done!")

}
