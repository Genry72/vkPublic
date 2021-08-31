package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

//MakeMailBox Создаем почтовый ящик
func MakeMailBox(domainMail, loginMail, passwordMail, pddToken string) (uid string, err error) {
	urls := "https://pddimp.yandex.ru/api2/admin/email/add"
	method := "POST"

	payload := strings.NewReader("domain=" + domainMail + "&login=" + loginMail + "&password=" + url.QueryEscape(passwordMail))

	client := &http.Client{}
	req, err := http.NewRequest(method, urls, payload)
	if err != nil {
		return uid, err
	}
	req.Header.Add("PddToken", pddToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return uid, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return uid, err
	}
	m := makeMailBoxStruct{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		return uid, err
	}
	if m.Success != "ok" {
		err = fmt.Errorf("Ошибка создания почтового ящика %v с паролем %v: %v", loginMail, passwordMail, string(body))
		return uid, err

	}
	return m.UID, err
}

type makeMailBoxStruct struct {
	Success string `json:"success"`
	UID     string `json:"uid"`
	Login   string `json:"login"`
	Domain  string `json:"domain"`
}
