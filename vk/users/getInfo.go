package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"bot.my/utils"
)

//GetInfo Отдает информацию по пользователю или группе
func GetInfo(id, keyAccesVk, versinonAPIVk, myID string, group bool, vkErrUserChan chan error) (firstName, lastName, sex, photoID string, err error) {
	log.Printf("Запрашиваем тнформацию пользователем %v для пользователя %v", myID, id)
	err = utils.Antiban(myID, "GetInfo", 300)
	if err != nil {
		err := fmt.Errorf("Проблема с антибаном: %v", err)
		vkErrUserChan <- err
	}
	funcName := "getUserGroupByID"
	var url string
	if group == false {
		url = "https://api.vk.com/method/users.get?user_ids=" + id + "&v=" + versinonAPIVk + "&access_token=" + keyAccesVk + "&fields=sex,photo_id"
	} else {
		url = "https://api.vk.com/method/groups.getById?group_ids=" + id + "&v=" + versinonAPIVk + "&access_token=" + keyAccesVk
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return firstName, lastName, sex, photoID, err
	}
	res, err := client.Do(req)
	if err != nil {
		return firstName, lastName, sex, photoID, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return firstName, lastName, sex, photoID, err
	}
	if group == false { //Получаем имя пользователя
		un := userInfoStruct{}
		err = json.Unmarshal(body, &un)
		if err != nil {
			err = errors.New(funcName + " %Парсинг боди запроса на получение имени пользователя по id: " + err.Error() + "Боди: " + string(body))
			return firstName, lastName, sex, photoID, err
		}
		if len(un.Response) == 0 {
			err = fmt.Errorf("Не улолось получить имя пользователя: %v url:%v", string(body), url)
			vkErrUserChan <- err
			return firstName, lastName, sex, photoID, err
		}
		firstName = un.Response[0].FirstName
		lastName = un.Response[0].LastName
		if un.Response[0].Sex == 1 {
			sex = "Женщина"
		}
		if un.Response[0].Sex == 2 {
			sex = "Мужчина"
		}
		photoID = un.Response[0].PhotoID
		//Добавляем запись в историчну таблицу о выполнении запроса
		err = utils.InsertToDB(fmt.Sprintf("INSERT INTO %v VALUES('%v', '%v', '%v', %v);", "history", time.Now().UnixNano(), "GetInfo", myID, "NOW()"))
		if err != nil {
			vkErrUserChan <- err
		}
		log.Printf("Пользоватлем %v выполнен запрос информации по пользоватлею %v", myID, id)
	} else { //Получаем имя группы
		gn := groupInfoStruct{}
		err = json.Unmarshal(body, &gn)
		if err != nil {
			err = errors.New(funcName + " Парсинг боди запроса на получение имени пользователя по id: " + err.Error() + "Боди: " + string(body))
			return firstName, lastName, sex, photoID, err
		}
		firstName = gn.Response[0].Name
		sex = ""
	}
	return firstName, lastName, sex, photoID, err
}

//userInfoStruct структура ответа на запрос получения имени пользователя по id
type userInfoStruct struct {
	Response []struct {
		ID               int    `json:"id"`
		FirstName        string `json:"first_name"`
		LastName         string `json:"last_name"`
		CanAccessClosed  bool   `json:"can_access_closed"`
		IsClosed         bool   `json:"is_closed"`
		Sex              int    `json:"sex"`
		CanInviteToChats bool   `json:"can_invite_to_chats"`
		PhotoID          string `json:"photo_id"`
	} `json:"response"`
}

//groupInfoStruct Структура ответа на запрос получения имени группы по id
type groupInfoStruct struct {
	Response []struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		ScreenName string `json:"screen_name"`
		IsClosed   int    `json:"is_closed"`
		Type       string `json:"type"`
		Photo50    string `json:"photo_50"`
		Photo100   string `json:"photo_100"`
		Photo200   string `json:"photo_200"`
	} `json:"response"`
}
