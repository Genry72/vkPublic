package vk

import (
	"strings"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func sendFromUser(token, msg, userID, versinonAPIVk string, queueChanZapros, queueChanOtvet chan string) (err error) {
	queueChanZapros <- "Согласуй"
	<-queueChanOtvet //Ожидание одобрения
	url := "https://api.vk.com/method/messages.send?v=" + versinonAPIVk + "&access_token=" + token + "&user_id=" + userID + "&message=" + url.QueryEscape(msg)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if strings.Contains(string(body), "error") == true { //Если в боди вернулась ошибка то прекращаем
		err = fmt.Errorf(string(body))
		return err
	}
	return err
}
