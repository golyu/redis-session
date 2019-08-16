package main

import (
	"github.com/gin-gonic/gin"
	redisSession "github.com/golyu/redis-session"
	"github.com/golyu/redis-session/data"
	"log"
	"net/http"
)

var store *redisSession.RedisStore

func main() {
	initStore()
	route := gin.Default()
	route.POST("/login", Login)
	auth := route.Group("", CheckLoginMiddleware)
	auth.GET("/info", GetInfo)
	if err := route.Run(":6666"); err != nil {
		log.Panicf("start http server failure:err%v", err)
	}
}
func initStore() {
	_ = data.InitRedis()
	store = redisSession.NewRedisStore(data.RedisConn, []byte("hash-key")).SetExpireSecond(data.SessionExpireTime).
		SetSerializer(&SessionInfoJsonSerialize{})
}

func Login(ctx *gin.Context) {
	reqData := struct {
		Account  string
		Password string
	}{}
	err := ctx.ShouldBind(&reqData)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"msg": "request data err"})
		return
	}
	if reqData.Account != "golyu" || reqData.Password != "123456" {
		ctx.JSON(http.StatusOK, gin.H{"msg": "account or password incorrect"})
		return
	}
	info := &SessionInfo{
		Account: "account",
		Name:    "golyu",
		Age:     30,
	}
	err = SaveSession(info, ctx)
	if err != nil {
		log.Printf("%v:\n", err)
		ctx.JSON(http.StatusOK, gin.H{"msg": "Operation session failed"})
		return
	}
}
func GetInfo(ctx *gin.Context) {
	info := GetSession(ctx)
	ctx.JSON(http.StatusOK, info)
}

//CheckLoginMiddleware
func CheckLoginMiddleware(ctx *gin.Context) {
	session, err := store.Get(ctx.Request, SessionKey)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	value, ok := session.Values[ValueKey]
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	sessionInfo, ok := value.(*SessionInfo)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.Set(ValueKey, sessionInfo)
	err = session.Save(ctx.Request, ctx.Writer)
	if err != nil {
		log.Printf("%v\n", err)
	}
	ctx.Next()
}
