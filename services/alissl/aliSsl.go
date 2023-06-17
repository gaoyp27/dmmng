package alissl

import (
	"dmmng/services/alidnsmng"
	"dmmng/services/domainlistmng"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	cas "github.com/aliyun/alibaba-cloud-sdk-go/services/cas"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

type certManager struct {
	domainName   string
	aliyunClient *alidns.Client
	cred         *credentials.AccessKeyCredential
	CertId       int64
	OrderId      int64
	certName     string
	client       *cas.Client
}

func NewCertManager(domainName string, aliyunClient *alidns.Client) *certManager {
	return &certManager{
		domainName:   domainName,
		aliyunClient: aliyunClient,
	}
}

// func (c *certManager) getCerd() (*cas.Client, error) {
func (c *certManager) getCerd() error {
	config := sdk.NewConfig()
	credential := credentials.NewAccessKeyCredential("aliyunkey", "aliyunsecret")
	client, err := cas.NewClientWithOptions("cn-hangzhou", config, credential)
	//if err != nil {
	//	panic(err)
	//}
	if err != nil {
		return err
	}
	c.client = client
	return nil
}

// 生成证书请求，并获取生成的证书 certid
// TODO certid 还是 orderid ？？？
func (c *certManager) getOrderId() {
	_ = c.getCerd()

	request := cas.CreateCreateCertificateForPackageRequestRequest()

	request.Scheme = "https"

	request.ProductCode = "vtrus-dv-w-starter"
	request.Username = "yourname"
	request.Phone = "yourphone"
	request.Email = "xx@126.com"
	request.Domain = "*." + c.domainName
	request.ValidateType = "DNS"

	response, err := c.client.CreateCertificateForPackageRequest(request)

	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Printf("response is %#v\n", response)

	OrderId := response.OrderId
	//if CertId !=0 {}
	fmt.Println(OrderId)
	c.OrderId = OrderId
}

//
//func (c *certManager) apiRequest() {}
//
//func (c *certManager) getCertList() {}
//
//func (c *certManager) putCertInfo() {}
//
//func (c *certManager) putCertOrder() {}

func (c *certManager) DownloadCertFile() string {
start:
	sslrecord := c.getCertRecord()
	path := "/Users/snail/Downloads/cert/certs/tmp/"

	if sslrecord["Type"] == "process" {
		fmt.Println("证书审核中等待 1 分钟！")
		time.Sleep(60 * time.Second)
		goto start
	}
	if sslrecord["Type"] == "domain_verify" {
		fmt.Println("等待dns 检验 1 分钟！")
		time.Sleep(60 * time.Second)
		goto start
	}
	if sslrecord["Type"] == "certificate" {

		_, err := os.Stat(path)
		fmt.Printf("%v--96line", err)
		if err != nil {
			if os.IsNotExist(err) {
				os.Mkdir(path, 0777)
			}
		}
		sslpemfile := path + c.domainName + ".pem"
		sslkeyfile := path + c.domainName + ".key"

		err = ioutil.WriteFile(sslpemfile, []byte(sslrecord["certificate"]), 0644)
		fmt.Println(err)
		if err != nil {
			fmt.Printf("write file false,err:%v\n", err)
			return ""
		}
		err = ioutil.WriteFile(sslkeyfile, []byte(sslrecord["privateKey"]), 0644)
		fmt.Println(err)
		if err != nil {
			fmt.Printf("write file false,err:%v\n", err)
			return ""
		}
		fmt.Printf("证书下载成功，地址为%s%s", sslpemfile, sslkeyfile)
		return path
	} else {
		fmt.Printf("当前状态为%s", sslrecord["Type"])
	}
	return path
}

// 获取
func (c *certManager) getCertRecord() map[string]string {
	_ = c.getCerd()

	request := cas.CreateDescribeCertificateStateRequest()

	request.Scheme = "https"

	request.OrderId = requests.NewInteger(int(c.OrderId))

	response, err := c.client.DescribeCertificateState(request)
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Printf("response is %#v\n\n", response)
	recordDict := map[string]string{
		"requestId":    response.RequestId,
		"Type":         response.Type,
		"domainName":   response.Domain,
		"recordType":   response.RecordType,
		"recordDomain": response.RecordDomain,
		"recordValue":  response.RecordValue,
		"validateType": response.ValidateType,
		"certificate":  response.Certificate,
		"privateKey":   response.PrivateKey,
	}
	fmt.Printf("%s\n", recordDict)
	return recordDict
}

// 触发证书校验,阿里云这里无单独接口
//func (c *certManager) getDnsAutoAuth() {}

// 调用 alidnsmng 添加 dns 解析
func (c *certManager) addDnsRecord() {
	fmt.Println(c)
start:
	result := c.getCertRecord()
	if result["Type"] == "payed" {
		fmt.Println("目前状态是 payed，等待订单提交\n")
		time.Sleep(3 * time.Second)
		goto start
	}
	if result["Type"] == "certificate" {
		fmt.Printf("%s DNS解析已自动添加！\n", result["domainName"])
		return
	}
	fmt.Println(result)
	sumdomainstr := result["recordDomain"]
	s := strings.SplitN(sumdomainstr, ".", 2)

	domainName := s[1]
	subName := s[0]
	value := result["recordValue"]
	Type := result["recordType"]

	fmt.Println(len(domainName))
	fmt.Println(len(subName))
	fmt.Println(len(value))
	fmt.Println(len(Type))

	if len(domainName) != 0 {
		if len(subName) != 0 {
			if len(value) != 0 {
				fmt.Println(subName, domainName, Type, value)
				aliyunClient := alidnsmng.GetAccessKeyFromDomainName(domainName)
				obj := alidnsmng.NewDNSDomainRecordManager(domainName, "cn-hangzhou", aliyunClient, "A", "@", "yourIP")
				obj.Get_or_add_dns_record(subName, value, Type)
			}
		}
	}
}

func (c *certManager) getCertificateIdByname() (certificates []cas.CertificateOrderListItem) {
	_ = c.getCerd()
	//TODO Keyword模糊查询 ,后续需要优化
	request := cas.CreateListUserCertificateOrderRequest()

	request.Scheme = "https"
	request.OrderType = "CERT"
	request.Status = "ISSUED"
	request.Keyword = c.domainName

	response, err := c.client.ListUserCertificateOrder(request)
	if err != nil {
		fmt.Print(err.Error())
	}
	certificates = response.CertificateOrderList

	return certificates
}

// 获取最新证书的 oderid
func (c *certManager) GetNewerCertificateIdByname() (int64, string) {
	certificates := c.getCertificateIdByname()
	var timeLayoutStr = "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Asia/Shanghai")

	for _, cert := range certificates {
		if cert.Status == "ISSUED" {

			nowTime := time.Now()
			st, _ := time.ParseInLocation(timeLayoutStr, cert.StartDate, loc)
			afterD, _ := time.ParseDuration("2160h")
			st = nowTime.Add(afterD)
			if nowTime.Before(st) {
				c.certName = cert.Name
				CertId := cert.CertificateId
				return CertId, c.certName
			}
		}
	}
	return 0, ""
}

// 定义采购 ssl 证书方法
func (c *certManager) BuySslCert(domainName string) error {
	err := c.getCerd()
	if err != nil {
		return err
	}

	c.getOrderId()
	c.getCertRecord()
	//time.Sleep(20 * time.Second)
	c.addDnsRecord()
	return nil
}

func (c *certManager) GetSslCert(domainName string) (string, string, error) {
start:
	sslrecord := c.getCertRecord()

	if sslrecord["Type"] == "process" {
		fmt.Println("证书审核中等待 1 分钟！")
		time.Sleep(60 * time.Second)
		goto start
	}
	if sslrecord["Type"] == "domain_verify" {
		fmt.Println("等待dns 检验 1 分钟！")
		time.Sleep(60 * time.Second)
		goto start
	}
	if sslrecord["Type"] == "certificate" {
		sslpemfile := sslrecord["certificate"]
		sslkeyfile := sslrecord["privateKey"]
		return sslpemfile, sslkeyfile, nil
	}
	return "", "", nil
}

func BuyGetCert(domainName string) {
	aliyunClient := alidnsmng.GetAccessKeyFromDomainName(domainName)
	if aliyunClient != nil {
		certObj := NewCertManager(domainName, aliyunClient)
		certObj.getCerd()
		//certObj.CertId = "xAip7gJq"
		certObj.getOrderId()
		certObj.getCertRecord()
		//time.Sleep(20 * time.Second)
		certObj.addDnsRecord()
		//certObj.getDnsAutoAuth()
		//certObj.getCerd()
		time.Sleep(30 * time.Second)
		certObj.DownloadCertFile()
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
		go BuyGetCert(domainlist[i])
	}
}
