package controllers

import (
	"context"
	"fmt"
	"html/template"
	"ldt/go-redis/config"
	"ldt/go-redis/models"
	"ldt/go-redis/services"
	"ldt/go-redis/utils"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthController struct {
	authService services.AuthService
	userService services.UserService
	ctx         context.Context
	collection  *mongo.Collection
	temp        *template.Template
}

func NewAuthController(authService services.AuthService,
	userService services.UserService,
	ctx context.Context,
	collection *mongo.Collection,
	temp *template.Template) AuthController {
	return AuthController{authService,
		userService, ctx, collection, temp}
}

func (ac *AuthController) SignUpUser(ctx *gin.Context) {
	user := &models.SignUpInput{}

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	if user.Password != user.PasswordConfirm {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Passwords do not match"})
		return
	}

	newUser, err := ac.authService.SignUpUser(user)

	if err != nil {
		if strings.Contains(err.Error(), "email already exist") {
			ctx.JSON(http.StatusConflict, gin.H{"status": "error", "message": err.Error()})
			return
		}
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		return
	}

	config, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("Could not load config", err)
	}

	// Generate Verification Code
	code := utils.RandStringRunes(20)

	verificationCode := utils.Encode(code)

	// Update User in Database
	ac.userService.UpdateUserById(newUser.ID.Hex(), "verificationCode", verificationCode)

	var firstName = newUser.Name

	if strings.Contains(firstName, " ") {
		firstName = strings.Split(firstName, " ")[1]
	}

	// ? Send Email
	emailData := utils.EmailData{
		URL:       config.Origin + "/verifyemail/" + code,
		FirstName: firstName,
		Subject:   "Your account verification code",
	}

	err = utils.SendEmail(newUser, &emailData, ac.temp, "verificationCode.html")
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "success", "message": "There was an error sending email"})
		return
	}

	message := "We sent an email with a verification code to " + user.Email
	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "message": message})

}

func (ac *AuthController) RefreshAccessToken(ctx *gin.Context) {

	cookie, err := ctx.Cookie("refresh_token")

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "fail", "message": "could not refresh access token"})
		return
	}

	config, _ := config.LoadConfig(".")

	sub, err := services.JwtObj.ValidateToken(cookie)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	user, err := ac.userService.FindUserById(fmt.Sprint(sub))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "fail", "message": "the user belonging to this token no logger exists"})
		return
	}

	access_token, err := services.JwtObj.CreateToken(user.ID.Hex())
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	ctx.SetCookie("access_token", access_token, config.AccessTokenMaxAge*60, "/", "localhost", false, true)
	ctx.SetCookie("logged_in", "true", config.AccessTokenMaxAge*60, "/", "localhost", false, false)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "access_token": access_token})
}

func (ac *AuthController) LogoutUser(ctx *gin.Context) {
	ctx.SetCookie("access_token", "", -1, "/", "localhost", false, true)
	ctx.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)
	ctx.SetCookie("logged_in", "", -1, "/", "localhost", false, true)

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (ac *AuthController) SignInUser(ctx *gin.Context) {
	var credentials *models.SignInInput

	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	user, err := ac.userService.FindUserByEmail(credentials.Email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid email or password"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	if err := utils.Pw.VerifyPassword(user.Password, credentials.Password); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid email or Password"})
		return
	}

	config, _ := config.LoadConfig(".")

	// Generate Tokens
	access_token, err := services.JwtObj.CreateToken(user.ID.Hex())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	refresh_token, err := services.JwtObj.CreateRefreshToken(user.ID.Hex())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	ctx.SetCookie("access_token", access_token, config.AccessTokenMaxAge*60, "/", "localhost", false, true)
	ctx.SetCookie("refresh_token", refresh_token, config.RefreshTokenMaxAge*60, "/", "localhost", false, true)
	ctx.SetCookie("logged_in", "true", config.AccessTokenMaxAge*60, "/", "localhost", false, false)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "access_token": access_token})
}

func (ac *AuthController) ForgotPassword(ctx *gin.Context) {
	var userCredential *models.ForgotPasswordInput

	if err := ctx.ShouldBindJSON(&userCredential); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	message := "You will receive a reset email if user with that email exist"

	user, err := ac.userService.FindUserByEmail(userCredential.Email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, gin.H{"status": "fail", "message": message})
			return
		}
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		return
	}

	if !user.Verified {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Account not verified"})
		return
	}

	config, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("Could not load config", err)
	}

	// Generate Verification Code
	resetToken := utils.RandStringRunes(20)

	passwordResetToken := utils.Encode(resetToken)

	// Update User in Database
	query := bson.D{{Key: "email", Value: strings.ToLower(userCredential.Email)}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "passwordResetToken", Value: passwordResetToken}, {Key: "passwordResetAt", Value: time.Now().Add(time.Minute * 15)}}}}
	result, err := ac.collection.UpdateOne(ac.ctx, query, update)

	if result.MatchedCount == 0 {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "success", "message": "There was an error sending email"})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"status": "success", "message": err.Error()})
		return
	}
	var firstName = user.Name

	if strings.Contains(firstName, " ") {
		firstName = strings.Split(firstName, " ")[1]
	}

	// ? Send Email
	emailData := utils.EmailData{
		URL:       config.Origin + "/resetpassword/" + resetToken,
		FirstName: firstName,
		Subject:   "Your password reset token (valid for 10min)",
	}

	err = utils.SendEmail(user, &emailData, ac.temp, "resetPassword.html")
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "success", "message": "There was an error sending email"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": message})
}

func (ac *AuthController) ResetPassword(ctx *gin.Context) {
	resetToken := ctx.Params.ByName("resetToken")
	var userCredential *models.ResetPasswordInput

	if err := ctx.ShouldBindJSON(&userCredential); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	if userCredential.Password != userCredential.PasswordConfirm {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Passwords do not match"})
		return
	}

	hashedPassword, _ := utils.Pw.HashPassword(userCredential.Password)

	passwordResetToken := utils.Encode(resetToken)

	// Update User in Database
	query := bson.D{{Key: "passwordResetToken", Value: passwordResetToken}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "password", Value: hashedPassword}}}, {Key: "$unset", Value: bson.D{{Key: "passwordResetToken", Value: ""}, {Key: "passwordResetAt", Value: ""}}}}
	result, err := ac.collection.UpdateOne(ac.ctx, query, update)

	if result.MatchedCount == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "success", "message": "Token is invalid or has expired"})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"status": "success", "message": err.Error()})
		return
	}

	ctx.SetCookie("access_token", "", -1, "/", "localhost", false, true)
	ctx.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)
	ctx.SetCookie("logged_in", "", -1, "/", "localhost", false, true)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Password data updated successfully"})
}

func (ac *AuthController) VerifyEmail(ctx *gin.Context) {

	code := ctx.Params.ByName("verificationCode")
	verificationCode := utils.Encode(code)

	query := bson.D{{Key: "verificationCode", Value: verificationCode}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "verified", Value: true}}}, {Key: "$unset", Value: bson.D{{Key: "verificationCode", Value: ""}}}}
	result, err := ac.collection.UpdateOne(ac.ctx, query, update)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"status": "success", "message": err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		ctx.JSON(http.StatusForbidden, gin.H{"status": "success", "message": "Could not verify email address"})
		return
	}

	log.Println(result)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Email verified successfully"})

}
