package users

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

//FrendsGet возвращает количество друзей текущего пользоватля
func FrendsGet(versinonAPIVk, token string) (count int, err error) {
	url := "https://api.vk.com/method/friends.get?v=" + versinonAPIVk + "&access_token=" + token + "&count=1"
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return count, err
	}
	res, err := client.Do(req)
	if err != nil {
		return count, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return count, err
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("Ошибка получения списка друзей пользователя: %v %v", res.Status, string(body))
		return count, err
	}
	m := frendsGetStruct{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("Ошибка парсинга боди на запрос FrendsGet: %v Боди: %v", err, string(body))
		return count, err
	}
	count = m.Response.Count
	return count, err
}

type frendsGetStruct struct {
	Response struct {
		Count int   `json:"count"`
		Items []int `json:"items"`
	} `json:"response"`
}
