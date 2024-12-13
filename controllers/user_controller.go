package controllers

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

type User struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

type UserController struct{}

func NewUserController() *UserController {
    return &UserController{}
}

func (uc *UserController) CreateUser(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "User created successfully",
        "user":    user,
    })
}

func (uc *UserController) GetUser(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "message": "Get user endpoint",
    })
}