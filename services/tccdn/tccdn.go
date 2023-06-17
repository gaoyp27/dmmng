package tccdn

import (
	"domainmng/alidnsmng"
	"domainmng/domainlistmng"
	"domainmng/tcssl"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"time"
)

// SecretId 获取 tc accesskey 和 secret
const SecretId = "tencentsecretid"
const SecretKey = "tencentsecretkey"

type cdnManager struct {
	domainName   string
	aliyunClient *alidns.Client
	accessKeyId  string
	accessSecret string
	cred         *common.Credential
	client       *cdn.Client
}

func newcdnManager(domainName string, aliyunClient *alidns.Client) *cdnManager {
	return &cdnManager{
		domainName:   domainName,
		aliyunClient: aliyunClient,
		accessKeyId:  SecretId,
		accessSecret: SecretKey,
	}
}

func (d *cdnManager) getCerd() *cdn.Client {
	credential := common.NewCredential(
		d.accessKeyId,
		d.accessSecret,
	)
	d.cred = credential
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cdn.tencentcloudapi.com"
	client, _ := cdn.NewClient(d.cred, "", cpf)
	d.client = client
	return client

}

func (d *cdnManager) getDnsAutoAuth() {

	request := cdn.NewDescribeDomainsConfigRequest()

	response, err := d.client.DescribeDomainsConfig(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Println("from getdnsauth line58")
		fmt.Printf("An API error has returned: %s", err)
		return
	}
	if err != nil {
		panic(err)
	}
	fmt.Printf("from getDnsAutoAuth line65 %s\n", response.ToJsonString())
}

func (d *cdnManager) getCnameFromCdn() string {
	request := cdn.NewDescribeDomainsConfigRequest()

	response, err := d.client.DescribeDomainsConfig(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Println("from getcnamfromcdn line 73")
		fmt.Printf("An API error has returned: %s", err)
		return ""
	}
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s", response.ToJsonString())
	cname := *response.Response.Domains[0].Cname
	fmt.Printf("cdn cname 是：%s\n", cname)
	return cname
}

func (d *cdnManager) createInstanceFromCopy() {
	request := cdn.NewDuplicateDomainConfigRequest()
	request.Domain = common.StringPtr("yun." + d.domainName)

	request.ReferenceDomain = common.StringPtr("yun.tuiaaaa.cn")

	response, err := d.client.DuplicateDomainConfig(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return
	}
	if err != nil {
		panic(err)
	}
	fmt.Printf("for createInstanceFromCopy line100%s\n", response.ToJsonString())
}

func (d *cdnManager) updateCdnWithHttps(certId string) {

	request := cdn.NewUpdateDomainConfigRequest()
	request.Domain = common.StringPtr("yun." + d.domainName)

	request.Https = &cdn.Https{
		Switch: common.StringPtr("on"),
		Http2:  common.StringPtr("on"),
		CertInfo: &cdn.ServerCert{
			CertId: common.StringPtr(certId),
		},
	}

	_, err := d.client.UpdateDomainConfig(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return
	}
	if err != nil {
		panic(err)
	}
	fmt.Printf("for updateCdnWithHttps line125 ,cdn配置更新成功\n")
}

// 添加 CDN 域名到 tc

func execManager(domainName string) {
	aliyunClient := alidnsmng.GetAccessKeyFromDomainName(domainName)
	if aliyunClient != nil {
		certObj := tcssl.NewCertManager(domainName, aliyunClient)
		certId := certObj.GetNewerCertificateIdByname()
		if certId != "" {
			cdnObj := newcdnManager(domainName, aliyunClient)
			cdnObj.getCerd()
			cdnObj.createInstanceFromCopy()
			time.Sleep(time.Duration(200) * time.Second)
			cdnObj.updateCdnWithHttps(certId)
			cdnCname := cdnObj.getCnameFromCdn()
			//fmt.Println(cdnCname)
			// aliyun 上添加解析
			dnsObj := alidnsmng.NewDNSDomainRecordManager(domainName, "cn-hangzhou", aliyunClient, "CNAME", "", "")
			dnsObj.Get_or_add_dns_record("yun", cdnCname, "CNAME")
		}

	} else {
		fmt.Println("域名不存在当前accesskey下")
	}

}

func Start() {

	domainlist, err := domainlistmng.GetDomainList()
	if err != nil {
		fmt.Println("获取文件列表出错！")
		return
	}
	for i := 0; i < len(domainlist); i++ {
		execManager(domainlist[i])
	}
}
