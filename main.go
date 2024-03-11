package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"electratype/jailbird/models"
)

var DB *gorm.DB

func MigrateDatabase() {

	//DB.Exec("CREATE OR REPLACE FUNCTION get_film_count(len_from int, len_to int)")

	DB.AutoMigrate(&models.Organization{}, &models.Project{}, &models.User{}, &models.OrganizationUser{}, &models.ProgressItem{}, &models.ApiKey{}, &models.ProjectUser{})
}

func ListOrganizations(c *gin.Context) {
	var organizations []models.Organization
	DB.Find(&organizations)

	c.JSON(http.StatusOK, &organizations)
}

func AddOrganization(c *gin.Context) {

	var plainOrg models.PlainOrganization
	var org models.Organization

	if err := c.ShouldBindJSON(&plainOrg); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "failure", "error": err.Error()})
	}

	log.Printf("%+v\n", &plainOrg)

	org.Slug = plainOrg.Slug
	org.Name = plainOrg.Name
	org.Logo = plainOrg.Logo

	result := DB.Create(&org)

	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		c.Error(result.Error)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "failure", "error": result.Error.Error()})
	}

	c.JSON(204, "")
}

func ListProjects(c *gin.Context) {

	var project []models.Project
	DB.Find(&project)

	c.JSON(http.StatusOK, &project)
}

func DeleteProject(c *gin.Context) {
	id := c.Param("projectId")

	DB.Where("slug = ?", id).Delete(&models.Project{})

	c.JSON(204, "")
}

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
			organizations := v1.Group("/organizations")
			{
				organizations.GET("", ListOrganizations)
				organizations.POST("", AddOrganization)

				projects := organizations.Group("/projects")
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
	}

	var err error
	DB, err = gorm.Open(postgres.Open(DSN), &gorm.Config{TranslateError: true})
	if err != nil {
		panic("failed to connect database")
	}

	MigrateDatabase()

	router.Run("localhost:5454")

	log.Println("done!")

}
