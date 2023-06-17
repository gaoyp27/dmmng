package tcssl

import (
	"domainmng/alidnsmng"
	"domainmng/domainlistmng"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
)

// SecretId 获取 tc accesskey 和 secret
const SecretId = "tencentsecretid"
const SecretKey = "tencentsecretkey"

type certManager struct {
	domainName   string
	aliyunClient *alidns.Client
	accessKeyId  string
	accessSecret string
	cred         *common.Credential
	client       *ssl.Client
	CertId       string
}

func NewCertManager(domainName string, aliyunClient *alidns.Client) *certManager {
	return &certManager{
		domainName:   domainName,
		aliyunClient: aliyunClient,
		accessKeyId:  SecretId,
		accessSecret: SecretKey,
	}
}

func (c *certManager) getCerd() *ssl.Client {
	credential := common.NewCredential(
		c.accessKeyId,
		c.accessSecret,
	)
	c.cred = credential
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ssl.tencentcloudapi.com"
	client, _ := ssl.NewClient(c.cred, "", cpf)
	c.client = client
	return client

}

//生成证书请求，并获取生成的证书 certid
func (c *certManager) getCertId() {
	c.getCerd()
	request := ssl.NewCreateCertificateRequest()

	request.ProductId = common.Int64Ptr(27)
	request.DomainNum = common.Int64Ptr(1)
	request.TimeSpan = common.Int64Ptr(1)

	response, err := c.client.CreateCertificate(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Println("print from getcertid!66line")
		fmt.Printf("An API error has returned: %s", err)
		return
	}
	if err != nil {
		panic(err)
	}
	CertIdSlice := response.Response.CertificateIds
	if len(CertIdSlice) == 0 {
		fmt.Println("CertIdSlice is null!")
		return
	}
	CertId := CertIdSlice[0]
	fmt.Println(*CertId)
	c.CertId = *CertId

	//fmt.Printf("getcertId func  %v", CertId)
}

func (c *certManager) apiRequest() {}

func (c *certManager) getCertList() string {

	request := ssl.NewDescribeCertificatesRequest()

	response, err := c.client.DescribeCertificates(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Println("print from getcertlist!line91")
		fmt.Printf("An API error has returned: %s", err)
		return ""
	}
	if err != nil {
		panic(err)
	}
	content := response.Response.Certificates
	//content := response.ToJsonString()
	b, _ := json.Marshal(content)
	return string(b)

	//fmt.Printf("%s", response.ToJsonString())
}
func (c *certManager) putCertInfo() {

	request := ssl.NewSubmitCertificateInformationRequest()
	request.CertificateId = common.StringPtr(c.CertId)

	request.CsrType = common.StringPtr("online")
	request.CertificateDomain = common.StringPtr("*." + c.domainName)
	request.OrganizationName = common.StringPtr("杭州兑吧网络科技有限公司")
	request.OrganizationDivision = common.StringPtr("ops")
	request.OrganizationAddress = common.StringPtr("数娱大厦 5 楼")
	request.OrganizationCountry = common.StringPtr("CN")
	request.OrganizationCity = common.StringPtr("HangZhou")
	request.OrganizationRegion = common.StringPtr("ZheJiang")
	request.PhoneAreaCode = common.StringPtr("0571")
	request.PhoneNumber = common.StringPtr("17794553302")
	request.VerifyType = common.StringPtr("DNS")
	request.AdminFirstName = common.StringPtr("先生")
	request.AdminLastName = common.StringPtr("高")
	request.AdminPhoneNum = common.StringPtr("17794553302")
	request.AdminEmail = common.StringPtr("tech_reg@duiba.com.cn")
	request.AdminPosition = common.StringPtr("ops")
	request.ContactFirstName = common.StringPtr("先生")
	request.ContactLastName = common.StringPtr("高")
	request.ContactEmail = common.StringPtr("tech_reg@duiba.com.cn")
	request.ContactNumber = common.StringPtr("17794553302")
	request.ContactPosition = common.StringPtr("ops")

	response, err := c.client.SubmitCertificateInformation(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Println("print from putcertinfo!line 133")
		fmt.Printf("An API error has returned: %s", err)
		return
	}
	if err != nil {
		panic(err)
	}
	//content := response.ToJsonString()
	fmt.Printf("%s", response.ToJsonString())
}

func (c *certManager) putCertOrder() {

	request := ssl.NewCommitCertificateInformationRequest()

	request.CertificateId = common.StringPtr(c.CertId)

	response, err := c.client.CommitCertificateInformation(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Println("print from putcertorder!line 152")
		fmt.Printf("An API error has returned: %s", err)
		return
	}
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s", response.ToJsonString())
}

func (c *certManager) getCertFile() string {

	path := "/Users/snail/Downloads/cert/certs/tmp/"
	request := ssl.NewDownloadCertificateRequest()
	request.CertificateId = common.StringPtr(c.CertId)
	response, err := c.client.DownloadCertificate(request)

	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Println("print from getcertfile! line 170")
		fmt.Printf("An API error has returned: %s", err)
		return ""
	}
	if err != nil {
		panic(err)
	}
	//fmt.Printf("%s", response.ToJsonString())
	content := *response.Response.Content

	//coder := base64.NewEncoding(content)
	//file, _ := coder.DecodeString(content)
	file, _ := base64.StdEncoding.DecodeString(content)
	_, err = os.Stat(path)
	fmt.Printf("%v", err)
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(path, 0777)
		}
	}
	filePath := path + c.domainName + ".zip"
	err = ioutil.WriteFile(filePath, []byte(file), 0644)
	fmt.Println(err)
	if err != nil {
		fmt.Printf("write file false,err:%v\n", err)
		return ""
	}
	fmt.Printf("证书下载成功，地址为%s", filePath)
	return filePath
}

func (c *certManager) getCertDnsRecord() map[string]string {
	credential := common.NewCredential(
		SecretId,
		SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ssl.tencentcloudapi.com"

	client1, _ := ssl.NewClient(credential, "", cpf)
re:
	for {
		request := ssl.NewDescribeCertificateDetailRequest()
		request.CertificateId = common.StringPtr(c.CertId)
		time.Sleep(time.Duration(15) * time.Second)
		response, err := client1.DescribeCertificateDetail(request)

		if _, ok := err.(*errors.TencentCloudSDKError); ok {
			fmt.Println("print from getCertDnsRecord!line205")
			fmt.Printf("An API error has returned: %s", err)
			return nil
		}
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", response.ToJsonString())
		content := response.Response
		//fmt.Printf("line215 content.DvAuthDetail.DvAuths:%s\n", content.DvAuthDetail.DvAuths)
		if content.DvAuthDetail.DvAuths == nil {
			continue re
		}
		result := content.DvAuthDetail.DvAuths[0]
		recordDict := map[string]string{"verifyType": *content.VerifyType,

			"domainName": *result.DvAuthDomain,
			"subName":    *result.DvAuthSubDomain,
			"value":      *result.DvAuthValue,
			"type":       *result.DvAuthVerifyType,
		}
		fmt.Printf("%s", recordDict)
		return recordDict
	}

}

//触发证书校验
func (c *certManager) getDnsAutoAuth() {

	request := ssl.NewCompleteCertificateRequest()
	request.CertificateId = common.StringPtr(c.CertId)
	response, err := c.client.CompleteCertificate(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Println("print from getdnsauroauth!line232")
		fmt.Printf("An API error has returned: %s", err)
		return
	}
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s", response.ToJsonString())
}

//调用 alidnsmng 添加 dns 解析
func (c *certManager) addDnsRecord() {
	fmt.Println(c)
	result := c.getCertDnsRecord()
	fmt.Println(result)
	domainName := result["domainName"]
	subName := result["subName"]
	value := result["value"]
	Type := result["type"]

	fmt.Println(len(domainName))
	fmt.Println(len(subName))
	fmt.Println(len(value))
	fmt.Println(len(Type))

	if len(domainName) != 0 {
		if len(subName) != 0 {
			if len(value) != 0 {
				fmt.Println(subName, domainName, Type, value)
				aliyunClient := alidnsmng.GetAccessKeyFromDomainName(domainName)
				obj := alidnsmng.NewDNSDomainRecordManager(domainName, "cn-hangzhou", aliyunClient, "A", "@", "120.26.53.4")
				obj.Get_or_add_dns_record(subName, value, Type)
			}
		}
	}
}

func (c *certManager) getCertificateIdByname() (certificates []*ssl.Certificates) {
	c.getCerd()
	request := ssl.NewDescribeCertificatesRequest()
	request.SearchKey = common.StringPtr(c.domainName)

	response, err := c.client.DescribeCertificates(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return nil
	}
	if err != nil {
		panic(err)
	}
	//fmt.Printf("%s", response.ToJsonString())
	certificates = response.Response.Certificates
	//fmt.Println(certificates)
	return certificates
}

func (c *certManager) GetNewerCertificateIdByname() (certId string) {
	certificates := c.getCertificateIdByname()
	var timeLayoutStr = "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Asia/Shanghai")
	//n := len(certificates)
	for _, cert := range certificates {
		if *cert.StatusName == "已通过" {
			nowTime := time.Now()
			st, _ := time.ParseInLocation(timeLayoutStr, *cert.CertBeginTime, loc)
			//fmt.Println(now.Sub(st))
			afterD, _ := time.ParseDuration("2160h")
			st = nowTime.Add(afterD)
			//t := now.Sub(st) < 86400*90
			//t := nowTime.Before(st)
			//fmt.Println(t)
			//if now.Sub(st) < 86400*90 {
			if nowTime.Before(st) {
				certId := *cert.CertificateId
				return certId
			}
		}
	}
	return ""
}

func execManager(domainName string) {
	aliyunClient := alidnsmng.GetAccessKeyFromDomainName(domainName)
	if aliyunClient != nil {
		certObj := NewCertManager(domainName, aliyunClient)
		certObj.getCerd()
		//certObj.CertId = "xAip7gJq"
		certObj.getCertId()
		time.Sleep(time.Duration(15) * time.Second)
		certObj.putCertInfo()
		time.Sleep(time.Duration(2) * time.Second)
		certObj.putCertOrder()
		time.Sleep(time.Duration(60) * time.Second)
		certObj.addDnsRecord()
		time.Sleep(time.Duration(120) * time.Second)
		certObj.getDnsAutoAuth()
		//certObj.getCerd()
		time.Sleep(time.Duration(300) * time.Second)
		certObj.getCertFile()
	} else {
		fmt.Println("域名不在当前 accesskey 下！")
	}
}

func Start() {
	domainlist, err := domainlistmng.GetDomainList()
	fmt.Printf("此次操作的域名列表为：%s\n", domainlist)
	if err != nil {
		fmt.Println("获取文件列表出错！")
		return
	}
	// 	ExecManager(domain)
	for i := 0; i < len(domainlist); i++ {
		execManager(domainlist[i])
	}
}
