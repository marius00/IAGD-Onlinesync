package eventbus

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/storage"
	"os"
)

const (
	GET = "get"
	POST = "post"
	PUT = "put"
	DELETE = "delete"
)

func MountPublicRoute(path string, method string, fn gin.HandlerFunc) *gin.Engine {
	engine := buildEngine()
	AddPublicRoute(engine, path, method, fn)
	return engine
}

func AddProtectedRoute(engine *gin.Engine, path string, method string, fn gin.HandlerFunc) *gin.Engine {
	authDb := &storage.AuthDb{}
	throttleDb := &storage.ThrottleDb{}
	group := engine.Group("/")
	group.Use(authorizedHandler(authDb, throttleDb))
	setMethodHandlerForGroup(method, path, fn, group)
	return engine
}

func MountProtectedRoute(path string, method string, fn gin.HandlerFunc) *gin.Engine {
	engine := buildEngine()
	AddProtectedRoute(engine, path, method, fn)
	return engine
}

func Build() *gin.Engine {
	engine := buildEngine()
	return engine
}

func AddPublicRoute(engine *gin.Engine, path string, method string, fn gin.HandlerFunc) *gin.Engine {
	group := engine.Group("/")
	setMethodHandlerForGroup(method, path, fn, group)
	return engine
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func buildEngine() *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	engine.Use(CORSMiddleware())
	return engine
}

func setMethodHandlerForGroup(method string, path string, fn gin.HandlerFunc, group *gin.RouterGroup) {
	switch method {
	case POST:
		{
			group.POST(path, fn)
		}
	case DELETE:
		{
			group.DELETE(path, fn)
		}
	case GET:
		{
			group.GET(path, fn)
		}
	}
}
