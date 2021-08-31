package groups

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

//GetMembers возвращает количество подписчиков в группе
func GetMembers(groupID, versinonAPIVk, token string) (count int, err error) {
	url := "https://api.vk.com/method/groups.getMembers?v=" + versinonAPIVk + "&access_token=" + token + "&group_id=" + groupID + "&count=1"
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
	fmt.Println(string(body))
	if res.StatusCode != 200 {
		err = fmt.Errorf("Ошибка получения количества подписчиков " + res.Status + " " + string(body))
		return count, err
	}
	m := getMembersStruct{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("Ошибка парсинга боди на запрос GetMembers: %v Боди: %v", err, string(body))
		return count, err
	}
	return m.Response.Count, err
}

type getMembersStruct struct {
	Response struct {
		Count int   `json:"count"`
		Items []int `json:"items"`
	} `json:"response"`
}
