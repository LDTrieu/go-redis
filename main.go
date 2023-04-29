package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/go-delve/delve/pkg/config"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	server = gin.Default()
	ctx    = context.Background()

	mongoclient    *mongo.Client
	authCollection *mongo.Collection

	redisclient *redis.Client

	// userService services.UserService
	// authService services.AuthService

	postCollection *mongo.Collection
)

func init() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		cfg, err = config.LoadConfig("../../")
		if err != nil {
			log.Fatal("Could not load config", err)
		}
	}

	// Connect to MongoDB
	mongoconnOpt := options.Client().ApplyURI(config.DBUri)
	mongoclient, err := mongo.Connect(ctx, mongoconnOpt)

	if err != nil {
		panic(err)
	}

	if err := mongoclient.Ping(ctx,
		readpref.Primary()); err != nil {
		panic(err)
	}
	log.Println("MongoDB successfully connected...")

	redisclient = redis.NewClient(&redis.Options{
		Addr: cfg.RedisUri,
	})

	err = redisclient.Set("test", "Welcome to Golang with Redis and MongoDB",
		0).Err()
	if err != nil {
		panic(err)
	}
	log.Println("Redis client connected successfully...")

	// Collections
	// authCollection = mongoclient.Database("golang_mongodb").Collection("users")
	// userService = services.NewUserServiceImpl(authCollection, ctx)
	// authService = services.NewAuthService(authCollection, ctx)
	// AuthController = controllers.NewAuthController(authService, userService, ctx, authCollection, temp)
	// AuthRouteController = routes.NewAuthRouteController(AuthController)

	postCollection = mongoclient.Database("golang_mongodb").Collection("posts")
	// postService = services.NewPostService(postCollection, ctx)
	// PostController = controllers.NewPostController(postService)
	// PostRouteController = routes.NewPostControllerRoute(PostController)

}

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		cfg, err = config.LoadConfig("../../")
		if err != nil {
			cfg, err = config.LoadConfig("../../")
			if err != nil {
				log.Fatal("Could not load config", err)
			}
		}
	}
	defer mongoclient.Disconnect(ctx)

	// startGinServer(config)

	startGrpcServer(cfg)
}

func startGrpcServer(config config.Config) {
	//authServer, err := gapi.
}
