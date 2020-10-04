package eventbus

import (
	"github.com/gin-gonic/gin"
	"os"
)

const (
	POST = "post"
)

func MountPublicRoute(path string, method string, fn gin.HandlerFunc) *gin.Engine {
	engine := buildEngine()
	group := engine.Group("/")
	setMethodHandlerForGroup(method, path, fn, group)
	return engine
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func buildEngine() *gin.Engine {
	if os.Getenv("SECRET") == "" {
		panic("Environmental variable 'SECRET' is not set -- aborting startup")
	}

	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	engine.Use(CORSMiddleware())
	engine.Use(authorizedHandler())
	return engine
}

func setMethodHandlerForGroup(method string, path string, fn gin.HandlerFunc, group *gin.RouterGroup) {
	switch method {
	case POST:
		{
			group.POST(path, fn)
		}
	}
}

func authorizedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if os.Getenv("SECRET") != token {
			respondWithError(401, "API: Authorization token invalid", c)
			return
		}
		c.Next()
	}
}

func respondWithError(code int, message string, c *gin.Context) {
	resp := map[string]string{"error": message}
	c.JSON(code, resp)
	c.Abort()
}