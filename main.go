package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func main() {
	http.HandleFunc("/single", AliyunMail)
	http.HandleFunc("/batch", AliyunMail)
	fmt.Println("server listen on 9093")
	err := http.ListenAndServe(":9093", nil)
	if err != nil {
		fmt.Println("http.ListenAndServe error: ", err.Error())
	}
}

func AliyunMail(w http.ResponseWriter, r *http.Request) {
	var mailInter interface{}
	var requestUrl string
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintln(w, (fmt.Sprintf("ioutil.ReadAll error: %s", err.Error())))
		return
	}

	if r.URL.Path == "/batch" {
		mail := DefaultBatch()
		mail.MailBase = DefaultBase()
		err = json.Unmarshal(data, mail)
		if err != nil {
			fmt.Fprintln(w, (fmt.Sprintf("json.Unmarshal to SingleMailSend error: %s", err.Error())))
			return
		}
		fmt.Printf("%+v\n", mail)
		mailInter = mail
		requestUrl = urlMap[regionType(mail.RegionId)]
	} else if r.URL.Path == "/single" {
		mail := DefaultSingle()
		mail.MailBase = DefaultBase()
		err = json.Unmarshal(data, mail)
		if err != nil {
			fmt.Fprintln(w, (fmt.Sprintf("json.Unmarshal to BatchMailSend error: %s", err.Error())))
			return
		}
		fmt.Printf("%+v\n", mail)
		mailInter = mail
		requestUrl = urlMap[regionType(mail.RegionId)]
	} else {
		fmt.Fprintln(w, (fmt.Sprintf("unsupport url path: %s", r.URL.Path)))
		return
	}
	err = validMailPara(mailInter)
	if err != nil {
		fmt.Fprintln(w, (fmt.Sprintf("validMailPara error: %s", err.Error())))
		return
	}
	reqStr, err := urlEncode(mailInter)
	if err != nil {
		fmt.Fprintln(w, (fmt.Sprintf("urlEncode error: %s", err.Error())))
		return
	}

	req, err := http.NewRequest("POST", requestUrl, strings.NewReader(reqStr))
	if err != nil {
		fmt.Fprintln(w, (fmt.Sprintf("http.NewRequest: %s", err.Error())))
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintln(w, (fmt.Sprintf("http.PostForm error: %s, request url: %s, \nrequest data: %s", err.Error(), requestUrl, reqStr)))
		return
	} else {
		fmt.Println("request data: ", reqStr)
	}
	defer resp.Body.Close()
	fmt.Printf("%+v", resp)
	respData, _ := ioutil.ReadAll(r.Body)
	w.WriteHeader(resp.StatusCode)
	w.Write(respData)
}
