package v1

import (
	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
	"taylors/global"
	"taylors/global/response"
	resp "taylors/model/response"
	"taylors/utils"
)

// @Tags base
// @Summary 生成验证码
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /base/captcha [post]
func Captcha(c *gin.Context) {
	captchaId := captcha.NewLen(global.GVA_CONFIG.Captcha.KeyLong)
	response.OkDetailed(resp.SysCaptchaResponse{
		CaptchaId: captchaId,
		PicPath:   "/base/captcha/" + captchaId + ".png",
	}, "验证码获取成功", c)
}

// @Tags base
// @Summary 生成验证码图片路径
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /base/captcha/:captchaId [get]
func CaptchaImg(c *gin.Context) {
	utils.GinCaptchaServeHTTP(c.Writer, c.Request)
}
