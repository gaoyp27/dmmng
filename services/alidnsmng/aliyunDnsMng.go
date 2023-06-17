package alidnsmng

import (
	"dmmng/services/domainlistmng"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"sync"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

var wg sync.WaitGroup

//定义 dnsmng 结构体

type DNSDomainRecordManager struct {
	DomainName   string `json:"domainName"`
	region       string `json:"region"`
	accessKeyID  string `json:"accessKeyID"`
	accessSecret string `json:"accessSecret"`
	Client       *alidns.Client

	IP     net.IP `json:"ip"`
	Scheme string `json:"scheme"` // 默认 https
	Type   string `json:"type"`   //@

	// json 解析字段 //解析 response 返回的数据
	RR       string `json:"RR"`
	RecordId string `json:"recordId"`
	Value    string `json:"Value"` //已经映射的的 ip 地址

	pageNumber      int `json:"page_number"`
	pageSize        int `json:"page_size"`
	totalPageNumber int `json:"totalPageNumber"`
}

type DdrmResponse struct {
	RequestId     string `json:"RequestId"`
	DomainRecords struct {
		Record []DNSDomainRecordManager
	} `json:"DomainRecords"`
}

// 定义设置默认参数函数

// 定义构造函数，返回结构体

func NewDNSDomainRecordManager(domainName, region string, client *alidns.Client, Type, RR, value string) *DNSDomainRecordManager {
	return &DNSDomainRecordManager{
		DomainName:      domainName,
		region:          region,
		Client:          client,
		Type:            Type,
		RR:              RR,
		Value:           value,
		Scheme:          "https",
		pageNumber:      1,
		pageSize:        100,
		totalPageNumber: 1,
	}
}

// func (m *DNSDomainRecordManager) set_request(commonArg) {
// 	request = alidnsmng.CreateDescribeDomainInfoRequest()
// 	if commonArg {
//        request.
// 	}
// }

func (m *DNSDomainRecordManager) auth_check(domainName string) (err error) {
	m.pageNumber = 1
	m.pageSize = 1
	m.DomainName = domainName
	request := alidns.CreateDescribeDomainInfoRequest()
	request.DomainName = m.DomainName
	_, err = m.Client.DescribeDomainInfo(request)
	return err
}

func (m *DNSDomainRecordManager) create_client(regionId, accessKeyId, accessKeySecret string) *alidns.Client {
	client, err := alidns.NewClientWithAccessKey(regionId, accessKeyId, accessKeySecret)
	m.region = regionId
	m.Client = client
	if err != nil {
		return nil
	}
	return client
}

// func (m *DNSDomainRecordManager) get_reponse(domainName string) {}

func (m *DNSDomainRecordManager) get_count() int {
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.DomainName = m.DomainName
	response, err := m.Client.DescribeDomainRecords(request)
	if err != nil {
		return -1
	}
	count := response.TotalCount
	return int(count)
}

func (m *DNSDomainRecordManager) get_total_page_number() int {
	count := m.get_count()
	mod := count % m.pageSize
	if mod != 0 {
		total_page_number := int(count/m.pageSize) + 1
		return total_page_number
	}
	total_page_number := int(count / m.pageSize)
	return total_page_number
}

// 查看域名解析
func (m *DNSDomainRecordManager) get_dns_record_list() []map[string]string {
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.DomainName = m.DomainName
	request.PageNumber = requests.NewInteger(m.pageNumber)
	request.PageSize = requests.NewInteger(m.pageSize)
	response, err := m.Client.DescribeDomainRecords(request)
	if err != nil {
		fmt.Printf("出错了！%s", err)
		return nil
	}
	b := response.GetHttpContentBytes()
	var resp DdrmResponse
	json.Unmarshal([]byte(b), &resp)
	records := resp.DomainRecords.Record

	recordList := make([]map[string]string, 0, 10)

	for _, record := range records {
		// var recordDict = make(map[string]string, 6)
		recordDict := map[string]string{"DomainName": record.DomainName,
			"RecordId": record.RecordId,
			"RR":       record.RR,
			"Value":    record.Value,
			"Type":     record.Type,
		}
		recordList = append(recordList, recordDict)
	}
	return recordList
}

func (m *DNSDomainRecordManager) get_total_dns_record_list() []map[string]string {
	total_record_list := make([]map[string]string, 0, 10)
	if m.totalPageNumber > 1 {
		n := m.totalPageNumber
		for i := 0; i < n; i++ {
			record_list := m.get_dns_record_list()
			total_record_list = append(total_record_list, record_list...)
			m.pageNumber += 1
		}
		return total_record_list
	}
	m.pageNumber = 1
	total_record_list = m.get_dns_record_list()
	return total_record_list
}

func (m *DNSDomainRecordManager) get_dns_record_id_by_name(recordName string) map[string]string {
	record_list := m.get_total_dns_record_list()

	for _, record := range record_list {
		if record["RR"] == recordName {
			return record
		}
	}
	return nil
}

// 增加域名解析
func (m *DNSDomainRecordManager) add_dns_record(recordName, value, Type string) string {
	request := alidns.CreateAddDomainRecordRequest()
	request.Scheme = "https"
	request.Type = Type
	request.Value = value
	request.TTL = "600"
	request.RR = recordName
	request.DomainName = m.DomainName

	response, err := m.Client.AddDomainRecord(request)
	recordId := response.RecordId
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Printf("response is %#v,recordId is %#v\n", response, recordId)
	return recordId

}

func (m *DNSDomainRecordManager) Get_or_add_dns_record(recordName, value, Type string) string {
	record := m.get_dns_record_id_by_name(recordName)
	if record != nil {
		return record["RecordId"]
	}
	recordId := m.add_dns_record(recordName, value, Type)
	return recordId

}

// 改动域名解析
func (m *DNSDomainRecordManager) updateDNSRecord(recordName, subName, value, Type string) string {
	RecordId := m.Get_or_add_dns_record(recordName, value, Type)
	//    RecordId == ""
	if len(RecordId) == 0 {
		return ""
	}
	request := alidns.CreateUpdateDomainRecordRequest()
	request.RecordId = RecordId
	request.Type = Type
	request.Value = value
	request.RR = subName
	response, err := m.Client.UpdateDomainRecord(request)
	//response, err := m.Client.UpdateDomainRecord(request)
	if err != nil {
		fmt.Println(err.Error())
	}
	//fmt.Printf("dns 更新完成，Response:%v ", response)
	return response.RecordId

}

// 阿里云域名管理系统
// 1.阿里云账户 key 和 secret
//   1.1定义 key 和 secret
// const accessKeyID = os.Getenv("ALIYUN_ACCESSKEYID")
// const accessSecret = os.Getenv("ALIYUN_ACCESSSECRET")

// var domainName = os.Getenv("DOMAINNAME")

//判断元素是否在被检测的类型（切片，数组等）中
func IsExistItem(value interface{}, array interface{}) bool {
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(value, s.Interface()) {
				return true
			}
		}
	}
	return false
}

// 2.验证域名归属权
//    2.1 如果域名在这个 key 下返回对象
//    2.2 如果域名不再这个 key 下，返回报错信息
//func GetAccessKeyFromDomainName(domainName string) *alidns.Client {
//
//	type domainclient = *alidns.Client
//
//	client1, _ := alidns.NewClientWithAccessKey("cn-hangzhou", "aliyunkey", "aliyunsecret")
//	client2, _ := alidns.NewClientWithAccessKey("cn-hangzhou", "LTAI5tJNHK3DksjTkcHXnCXG", "8aX3WO7cYBF2I46IwA0cgJugC2ofTA")
//	client3, _ := alidns.NewClientWithAccessKey("cn-hangzhou", "LTAI5tJkWLyj58ERpmZWTGbw", "mRH3xGnXzJ64jl1eo9Sm3VxQ579a9v")
//
//	clientslice := [3]domainclient{client1, client2, client3}
//
//	request := alidns.CreateDescribeDomainInfoRequest()
//	request.Scheme = "https"
//	request.AcceptFormat = "json"
//	request.DomainName = domainName
//	request.Lang = "zh_cn"
//
//	domainInDnsPodList := []string{"dui88.com", "tuia.cn", "duiba.com.cn"}
//	ok := IsExistItem(domainName, domainInDnsPodList)
//	if ok {
//		fmt.Println("域名在 DNSPOD 里！")
//		return nil
//	}
//	// 调试
//	// fmt.Println(len(clientslice))
//
//	for i := 0; i < len(clientslice); i++ {
//		_, err := clientslice[i].DescribeDomainInfo(request)
//		if err != nil {
//			fmt.Print(err.Error())
//			return nil
//		}
//		clientrequest := clientslice[i]
//		return clientrequest
//		// fmt.Printf("response is %#v\n", response)
//	}
//	return nil
//}

func GetAccessKeyFromDomainName(domainName string) *alidns.Client {

	domainclient, _ := alidns.NewClientWithAccessKey("cn-hangzhou", "aliyunkey", "aliyunsecret")

	request := alidns.CreateDescribeDomainInfoRequest()
	request.Scheme = "https"
	request.AcceptFormat = "json"
	request.DomainName = domainName
	request.Lang = "zh_cn"

	domainInDnsPodList := []string{"dui88.com", "tuia.cn", "duiba.com.cn"}
	ok := IsExistItem(domainName, domainInDnsPodList)
	if ok {
		fmt.Println("域名在 DNSPOD 里！")
		return nil
	}
	_, err := domainclient.DescribeDomainInfo(request)
	if err != nil {
		fmt.Print(err.Error())
		return nil
	}
	return domainclient
}

func ExecManager(domainName string) {
	defer wg.Done()
	aliyunClient := GetAccessKeyFromDomainName(domainName)

	if aliyunClient != nil {
		//添加解析配置
		obj := NewDNSDomainRecordManager(domainName, "cn-hangzhou", aliyunClient, "A", "@", "120.26.53.4")
		obj.totalPageNumber = obj.get_total_page_number()
		subName := "*"
		//domainrecordlist := obj.get_total_dns_record_list()
		RecordId := obj.Get_or_add_dns_record(subName, obj.Value, obj.Type)
		fmt.Println(RecordId)
	} else {
		fmt.Println("域名不存在当前accesskey下")
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
		wg.Add(1)
		ExecManager(domainlist[i])
	}
	wg.Wait()
}
