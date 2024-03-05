package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "electratype/jailbird/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var DB *gorm.DB

type ApiKey struct {
	ID        uint
	Name      string
	Value     uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time
}

// swagger:model
type PlainProject struct {

	// The id/slug for the project
	// Required: true
	// Min length: 1
	Slug string `gorm:"unique" json:"id" binding:"required" validate:"min=1,regexp=^[a-zA-Z0-9-]*$"`

	// Name of the project
	// Required: false
	Name *string `json:"name"`

	// Description of the project
	// Required: false
	Description *string `json:"description"`
}

type Project struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	PlainProject
}

type ProgressItem struct {
	gorm.Model
	EID         uuid.UUID
	CompletedAt time.Time
	State       string
}

func MigrateDatabase() {
	DB.AutoMigrate(&Project{})
}

// @Summary Return all projects
// @Schemes
// @Description Returns array of all projects
// @Tags project
// @Accept json
// @Produce json
// @Success 200 {array} main.Project
// @Router /projects [get]
func ListProjects(c *gin.Context) {

	var project []Project
	DB.Find(&project)

	c.JSON(http.StatusOK, &project)
}

// @Summary Delete project
// @Tags project
// @Param id path string true "Project ID"
// @Success 200
// @Router /projects/{id} [delete]
func DeleteProject(c *gin.Context) {
	id := c.Param("id")

	DB.Where("slug = ?", id).Delete(&Project{})

	c.JSON(204, "")
}

// @Summary Create new project
// @Tags project
// @Param body body PlainProject true "Project definition"
// @Accept json
// @Produce json
// @Success 200
// @Router /projects [post]
func AddProject(c *gin.Context) {

	var plainProject PlainProject
	var project Project

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
				projects.DELETE(":id", DeleteProject)
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
