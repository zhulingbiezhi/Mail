package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"time"
)

var (
	AccessKeyId     string     = "LTAIptPMe9Ayia39"
	AccessKeySecret string     = "testsecret"
	RegionID        regionType = HangZhou
)

type regionType string

const (
	HangZhou   regionType = "cn-hangzhou"
	SouthEast1 regionType = "ap-southeast-1"
	SouthEast2 regionType = "ap-southeast-2"
)

var versionMap = map[regionType]string{
	HangZhou:   "2015-11-23",
	SouthEast1: "2017-06-22",
	SouthEast2: "2017-06-22",
}
var urlMap = map[regionType]string{
	HangZhou:   "https://dm.aliyuncs.com",
	SouthEast1: "https://dm.ap-southeast-1.aliyuncs.com",
	SouthEast2: "https://dm.ap-southeast-2.aliyuncs.com",
}

type MailBase struct {
	Format           string `json:"Format"`
	Version          string `json:"Version" required:"ture"`
	AccessKeyId      string `json:"AccessKeyId" required:"ture"`
	Signature        string `json:"Signature" required:"ture"`
	SignatureMethod  string `json:"SignatureMethod" required:"ture"`
	Timestamp        string `json:"Timestamp" required:"ture"`
	SignatureVersion string `json:"SignatureVersion" required:"ture"`
	SignatureNonce   string `json:"SignatureNonce" required:"ture"`
	RegionId         string `json:"RegionId"`
}

func DefaultBase() MailBase {
	return MailBase{
		Format:           "JSON",
		Version:          versionMap[RegionID],
		AccessKeyId:      AccessKeyId,
		SignatureMethod:  "HMAC-SHA1",
		Timestamp:        getFormatTime(),
		SignatureVersion: "1.0",
		SignatureNonce:   generateRandom(32),
		RegionId:         string(RegionID),
	}
}

type SingleMailSend struct {
	MailBase
	Action         string `json:"Action" required:"ture"`
	AccountName    string `json:"AccountName" required:"ture"`
	ReplyToAddress bool   `json:"ReplyToAddress" required:"ture"`
	AddressType    int    `json:"AddressType" required:"ture"`
	ToAddress      string `json:"ToAddress" required:"ture"`
	FromAlias      string `json:"FromAlias"`
	Subject        string `json:"Subject" required:"ture"`
	HtmlBody       string `json:"HtmlBody" required:"ture"`
	TextBody       string `json:"TextBody" required:"ture"`
	ClickTrace     string `json:"ClickTrace"`
}

func DefaultSingle() *SingleMailSend {
	return &SingleMailSend{
	//Action:         "SingleSendMail",
	//ReplyToAddress: false,
	//AddressType:    1,
	//		FromAlias:      "admin",
	//		Subject:        "mailTitle",
	//		ClickTrace:     "1",
	}
}

type BatchMailSend struct {
	MailBase
	Action        string `json:"action" required:"ture"`
	AccountName   string `json:"account_name" required:"ture"`
	AddressType   int    `json:"address_type" required:"ture"`
	TemplateName  string `json:"template_name" required:"ture"`
	ReceiversName string `json:"receivers_name" required:"ture"`
	TagName       string `json:"tag_name"`
	ClickTrace    string `json:"click_trace"`
}

func DefaultBatch() *BatchMailSend {
	return &BatchMailSend{
		Action:      "BatchSendMail",
		AddressType: 1,
		ClickTrace:  "1",
	}
}

func getFormatTime() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

func generateRandom(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
func getHmac(key, msg string) string {
	message := "POST&%2F&" + msg
	fmt.Println("final sign---", string(message))

	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(message))
	expectedMAC := mac.Sum(nil)
	fmt.Printf("expectedMAC----%x", expectedMAC)
	return base64.RawURLEncoding.EncodeToString(expectedMAC)
}

func urlEncodeReplace(val string) string {
	val = url.QueryEscape(val)
	val = strings.Replace(val, "+", "%20", -1)
	val = strings.Replace(val, "*", "%2A", -1)
	val = strings.Replace(val, "%7E", "~", -1)
	return val
}

func urlEncode(inter interface{}) (string, error) {
	jsonBytes, err := json.Marshal(inter)
	if err != nil {
		return "", err
	}
	//fmt.Println(string(jsonBytes))
	var paramsMap map[string]interface{}
	err = json.Unmarshal(jsonBytes, &paramsMap)
	if err != nil {
		return "", err
	}

	var keys []string
	for k, _ := range paramsMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var resultString, signString string
	for _, tag := range keys {
		valStr := fmt.Sprint(paramsMap[tag])
		if valStr != "" {
			urlKey := urlEncodeReplace(tag)
			urlValue := urlEncodeReplace(valStr)
			signString = signString + urlKey + "=" + urlValue + "&"
		}
	}
	resultString = strings.TrimSuffix(signString, "&")
	signString = urlEncodeReplace(resultString)
	signString = strings.Replace(signString, "%26", "&", -1)
	signResult := getHmac(AccessKeySecret+"&", signString)
	resultString = resultString + "&signature=" + urlEncodeReplace(signResult)
	return resultString, nil
}

func validMailPara(v interface{}) error {
	refVal := reflect.ValueOf(v)
	refType := reflect.TypeOf(v)
	for i := 0; i < refVal.Elem().NumField(); i++ {
		tag := refType.Elem().Field(i).Tag.Get("required")
		if tag == "ture" {
			val := refVal.Elem().Field(i).String()
			if val == "" {
				return fmt.Errorf("the field %s required, can't be empty !", refType.Elem().Field(i).Name)
			}
		}
	}
	return nil
}

func replaceSpecial(in string) string {
	in = strings.Replace(in, "+", "%20", -1)
	in = strings.Replace(in, "*", "%2A", -1)
	return strings.Replace(in, "%7E", "~", -1)
}
