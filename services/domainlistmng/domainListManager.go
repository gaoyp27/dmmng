package domainlistmng

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func GetDomainList() ([]string, error) {
	domainNameFile := "/Users/snail/devwork/gop/test/src/aliyungop/domainmng/list"
	file, err := os.Open(domainNameFile)
	s := make([]string, 0, 10)
	if err != nil {
		fmt.Println("open file error:", err)
		return nil, err
	}
	defer file.Close()
	r := bufio.NewReader(file)
	for {
		str, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("read error:", err)
			return nil, err
		}
		domain := string(str)
		domainname := strings.TrimSpace(domain)
		s = append(s, domainname)
	}
	return s, err
}
