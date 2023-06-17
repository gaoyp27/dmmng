package main

import (
	"fmt"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"

	"dmmng/services/alidnsmng"
	"dmmng/services/alissl"
)

func main() {
	// 创建 Gin 实例
	r := gin.Default()

	// 加载 HTML 模板
	r.HTMLRender = createTemplate()

	// 添加路由
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})
	r.POST("/buy-ssl", func(c *gin.Context) {

		c.HTML(00, "index.html", gin.H{})
		// 从请求中获取域名列表
		domainList := strings.Split(c.PostForm("domains"), ",")

		for _, domain := range domainList {
			aliyunClient := alidnsmng.GetAccessKeyFromDomainName(domain)
			cm := alissl.NewCertManager(domain, aliyunClient)
			err := cm.BuySslCert(domain)
			if err != nil {
				// 处理错误
				c.String(http.StatusInternalServerError, "Failed to purchase SSL certificate for domain %s: %s\n", domain, err)
				return
			}
		}
		// 处理证书文件
		c.String(http.StatusOK, "SSL certificates have been purchased successfully ")
	})

	r.POST("/download-ssl", func(c *gin.Context) {
		domainList := strings.Split(c.PostForm("domains"), ",")
		for _, domain := range domainList {
			aliyunClient := alidnsmng.GetAccessKeyFromDomainName(domain)
			cm := alissl.NewCertManager(domain, aliyunClient)

			certFile, keyFile, err := cm.GetSslCert(domain)
			if err != nil {
				// 处理错误
				c.String(http.StatusInternalServerError, "Failed to download SSL certificate for domain %s: %s\n", domain, err)
				continue
				// 返回成功信息
			}
			certData, err := ioutil.ReadFile(certFile)
			if err != nil {
				c.String(500, "Failed to read certificate file %s: %s", certFile, err.Error())
				return
			}
			keyData, err := ioutil.ReadFile(keyFile)
			if err != nil {
				c.String(500, "Failed to read private key file %s: %s", keyFile, err.Error())
				return
			}

			// 设置响应头，指定文件名和文件类型
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", domain))
			c.Header("Content-Type", "application/zip")
			// 将文件内容写入响应体
			// 这里可以使用任何您熟悉的方式来将文件内容写入响应体
			// 例如，您可以使用 io.Copy 函数将文件内容复制到响应体中
			_, err = c.Writer.Write(append(certData, keyData...))
			if err != nil {
				c.String(500, "Failed to write file to response: %s", err.Error())
				return
			}
		}
		c.HTML(200, "index.html", gin.H{"message": "SSL certificate  downloaded"})
	})

	// 启动 Gin
	r.Run()
}

func createTemplate() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("index.html", "templates/index.html")
	return r
}
