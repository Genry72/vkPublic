package users

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"bot.my/utils"
)

//FindUsers производит поиск пользоватлей. Возвращает 1000 id
func FindUsers(token, sex, versinonAPIVk, myID string, vkErrUserChan chan error) (ids []string, err error) {
	err = utils.Antiban(myID, "FindUsers", 300)
	if err != nil {
		vkErrUserChan <- err
	}
	vozrastFrom := "18"
	vozrastTo := "24"
	var pol string
	if sex == "Мужчина" {
		pol = "2"
	}
	if sex == "Женщина" {
		pol = "1"
	}
	url := "https://api.vk.com/method/users.search?v=" + versinonAPIVk + "&access_token=" + token + "&age_from=" + vozrastFrom + "&age_to=" + vozrastTo + "&country=1&has_photo=1&sex=" + pol + "&count=1000&is_closed=false"
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return ids, err
	}

	res, err := client.Do(req)
	if err != nil {
		return ids, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ids, err
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("Ошибка поиска пользователей " + res.Status + " " + string(body))
		return ids, err
	}
	if strings.Contains(string(body), "failed") == true || strings.Contains(string(body), "error") {
		err = fmt.Errorf("Ошибка поиска пользователей: %v", string(body))
		return ids, err
	}

	m := findUsersStruct{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("Ошибка парсинга ответа на запрос FindUsers: %v Боди: %v", err, string(body))
		return ids, err
	}
	for _, items := range m.Response.Items {
		ids = append(ids, fmt.Sprintf("%v", items.ID))
	}
	//Добавляем запись в историчну таблицу о выполнении запроса
	err = utils.InsertToDB(fmt.Sprintf("INSERT INTO %v VALUES('%v', '%v', '%v', %v);", "history", time.Now().UnixNano(), "FindUsers", myID, "NOW()"))
	if err != nil {
		vkErrUserChan <- err
	}
	return ids, err
}

type findUsersStruct struct {
	Response struct {
		Count int `json:"count"`
		Items []struct {
			FirstName        string `json:"first_name"`
			ID               int    `json:"id"`
			LastName         string `json:"last_name"`
			CanAccessClosed  bool   `json:"can_access_closed"`
			IsClosed         bool   `json:"is_closed"`
			CanInviteToChats bool   `json:"can_invite_to_chats"`
			TrackCode        string `json:"track_code"`
		} `json:"items"`
	} `json:"response"`
}
