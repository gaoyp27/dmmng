package alicdn

import (
	"dmmng/services/alidnsmng"
	"dmmng/services/alissl"
	"dmmng/services/domainlistmng"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"io/ioutil"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	cdn "github.com/aliyun/alibaba-cloud-sdk-go/services/cdn"
)

// SecretId 获取 tc accesskey 和 secret
const accessKeyId = aliyunkey
const accessSecret = aliyunsecret

type cdnManager struct {
	domainName   string
	aliyunClient *alidns.Client
	accessKeyId  string
	accessSecret string
	cred         *credentials.AccessKeyCredential
	client       *cdn.Client
}

func newcdnManager(domainName string, aliyunClient *alidns.Client) *cdnManager {
	return &cdnManager{
		domainName:   domainName,
		aliyunClient: aliyunClient,
		accessKeyId:  accessKeyId,
		accessSecret: accessSecret,
	}
}

func (d *cdnManager) getCerd() *cdn.Client {

	config := sdk.NewConfig()

	credential := credentials.NewAccessKeyCredential(accessKeyId, accessSecret)
	d.cred = credential
	/* use STS Token
	 */
	client, err := cdn.NewClientWithOptions("cn-hangzhou", config, d.cred)
	if err != nil {
		panic(err)
	}

	d.client = client
	return client
}

func (d *cdnManager) addCdnDomain() {
	d.getCerd()
	request := cdn.CreateAddCdnDomainRequest()
	request.Scheme = "https"
	request.CdnType = "web"
	request.DomainName = "yun." + d.domainName
	request.Sources = "duiba.oss-cn-hangzhou.aliyuncs.com"
	request.Scope = "domestic"
	request.Sources = "[{\"content\":\"duiba.oss-cn-hangzhou.aliyuncs.com\",\"type\":\"oss\",\"priority\":\"20\",\"port\":80,\"weight\":\"15\"}]"

	_, err := d.client.AddCdnDomain(request)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	fmt.Printf("%s CDN添加成功！", d.domainName)

}

//func (d *cdnManager) getDnsAutoAuth() {}

func (d *cdnManager) getCnameFromCdn() string {
	d.getCerd()

	request := cdn.CreateDescribeCdnDomainDetailRequest()
	request.Scheme = "https"
	request.DomainName = "yun." + d.domainName

	response, err := d.client.DescribeCdnDomainDetail(request)
	if err != nil {
		fmt.Print(err.Error())
	}
	//fmt.Printf("response is %#v\n", response)
	cname := response.GetDomainDetailModel.Cname
	fmt.Printf("cdn cname 是：%s\n", cname)
	return cname
}

// 从模板复制 cdn 配置
//func (d *cdnManager) createInstanceFromCopy() {}

// 配置 cdn,需要处理 域名列表得到cdndomainlist
func (d *cdnManager) batchSetCdnDomainConfig() {
	//func (d *cdnManager) batchSetCdnDomainConfig(cdndomainlist string) {
	d.getCerd()
	path := "/Users/snail/devwork/gop/test/src/aliyungop/domainmng/alicdn/functions.json"
	// 使用 ioutil.ReadFile 函数一次性读取所有字符串
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("从functions.json 读取文件出错！！！")
		//panic(err)
	}
	functions := string(bytes)

	request := cdn.CreateBatchSetCdnDomainConfigRequest()
	request.Scheme = "https"
	//request.DomainNames = cdndomainlist
	request.DomainNames = "yun." + d.domainName
	request.Functions = functions

	_, err = d.client.BatchSetCdnDomainConfig(request)
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Printf("cdn 域名s%更新完成！", d.domainName)
	//fmt.Printf("response is %#v\n", response)
}

//

func (d *cdnManager) updateCdnWithHttps(certName string) {
	d.getCerd()

	request := cdn.CreateSetDomainServerCertificateRequest()

	request.Scheme = "https"
	request.ServerCertificateStatus = "on"
	request.CertType = "cas"
	request.CertName = certName
	request.DomainName = "yun." + d.domainName

	_, err := d.client.SetDomainServerCertificate(request)
	if err != nil {
		fmt.Print(err.Error())
	}
	//fmt.Printf("response is %#v\n", response)
	fmt.Printf("cdn 域名s%配置证书完成！", d.domainName)
}

// 添加 CDN 域名到 tc

func execManager(domainName string) {
	aliyunClient := alidnsmng.GetAccessKeyFromDomainName(domainName)
	if aliyunClient != nil {
		cdnObj := newcdnManager(domainName, aliyunClient)
		cdnObj.getCerd()
		cdnObj.addCdnDomain()
		certObj := alissl.NewCertManager(domainName, aliyunClient)
		_, certName := certObj.GetNewerCertificateIdByname()
		if certName != "" {
			cdnObj.updateCdnWithHttps(certName)
			//time.Sleep(time.Duration(200) * time.Second)
			cdnObj.batchSetCdnDomainConfig()
			cdnCname := cdnObj.getCnameFromCdn()
			//fmt.Println(cdnCname)
			// aliyun 上添加解析
			dnsObj := alidnsmng.NewDNSDomainRecordManager(domainName, "cn-hangzhou", aliyunClient, "CNAME", "", "")
			dnsObj.Get_or_add_dns_record("yun", cdnCname, "CNAME")
		}

	} else {
		fmt.Printf("%s域名不存在当前accesskey下", domainName)
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
