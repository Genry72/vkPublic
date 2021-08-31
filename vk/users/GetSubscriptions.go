package users

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bot.my/utils"
)

//GetSubscriptions Возвращает список идентификаторов групп, на которые подписан пользователь
func GetSubscriptions(versinonAPIVk, token, userID, myID string, vkErrUserChan chan error) (subscrybers []string, err error) {
	err = utils.Antiban(myID, "GetSubscriptions", 300)
	if err != nil {
		vkErrUserChan <- err
	}
	url := "https://api.vk.com/method/groups.get?v=" + versinonAPIVk + "&access_token=" + token + "&user_id=" + userID
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return subscrybers, err
	}
	res, err := client.Do(req)
	if err != nil {
		return subscrybers, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return subscrybers, err
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("Ошибка получения списка групп " + res.Status + " " + string(body))
		return subscrybers, err
	}
	if strings.Contains(string(body), "failed") == true || strings.Contains(string(body), "error") {
		err = fmt.Errorf("Ошибка получения списка групп: %v", string(body))
		return subscrybers, err
	}
	m := getSubscriptionsStruct{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("Ошибка парсинга ответа на запрос GetSubscriptions: %v Боди: %v", err, string(body))
		return subscrybers, err
	}
	for _, group := range m.Response.Items {
		subscrybers = append(subscrybers, strconv.FormatInt(int64(group), 10))
	}
	//Добавляем запись в историчну таблицу о выполнении запроса
	err = utils.InsertToDB(fmt.Sprintf("INSERT INTO %v VALUES('%v', '%v', '%v', %v);", "history", time.Now().UnixNano(), "GetSubscriptions", myID, "NOW()"))
	if err != nil {
		vkErrUserChan <- err
	}
	log.Printf("Пользоватлем %v запрошена информация о подписках пользователя %v", myID, userID)
	return subscrybers, err
}

type getSubscriptionsStruct struct {
	Response struct {
		Count int   `json:"count"`
		Items []int `json:"items"`
	} `json:"response"`
}