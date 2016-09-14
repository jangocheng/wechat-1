package context

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/silenceper/wechat/util"
)

const (
	//AccessTokenURL 获取access_token的接口
	AccessTokenURL = "https://api.weixin.qq.com/cgi-bin/token"
)

//ResAccessToken struct
type ResAccessToken struct {
	ErrorCode int32  `json:"errcode"`
	ErrorMsg  string `json:"errmsg"`

	AccessToken string `json:"access_token"`
	ExpiresIn   int32  `json:"expires_in"`
}

//SetAccessTokenLock 设置读写锁（一个appID一个读写锁）
func (ctx *Context) SetAccessTokenLock(l *sync.RWMutex) {
	ctx.accessTokenLock = l
}

//GetAccessToken 获取access_token
func (ctx *Context) GetAccessToken() (accessToken string, err error) {
	ctx.accessTokenLock.Lock()
	defer ctx.accessTokenLock.Unlock()

	accessTokenCacheKey := fmt.Sprintf("access_token_%s", ctx.AppID)
	val := ctx.Cache.Get(accessTokenCacheKey)
	if val != nil {
		accessToken = val.(string)
		return
	}

	//从微信服务器获取
	var resAccessToken ResAccessToken
	resAccessToken, err = ctx.GetAccessTokenFromServer()
	if err != nil {
		return
	}

	accessToken = resAccessToken.AccessToken
	return
}

//GetAccessTokenFromServer 强制从微信服务器获取token
func (ctx *Context) GetAccessTokenFromServer() (resAccessToken ResAccessToken, err error) {
	url := fmt.Sprintf("%s?grant_type=client_credential&appid=%s&secret=%s", AccessTokenURL, ctx.AppID, ctx.AppSecret)
	var body []byte
	body, err = util.HTTPGet(url)
	err = json.Unmarshal(body, &resAccessToken)
	if err != nil {
		return
	}
	if resAccessToken.ErrorMsg != "" {
		err = fmt.Errorf("get access_token error : errcode=%v , errormsg=%v", resAccessToken.ErrorCode, resAccessToken.ErrorMsg)
		return
	}

	accessTokenCacheKey := fmt.Sprintf("access_token_%s", ctx.AppID)
	err = ctx.Cache.Set(accessTokenCacheKey, resAccessToken.AccessToken, time.Duration(resAccessToken.ExpiresIn)-1500/time.Second)
	return
}