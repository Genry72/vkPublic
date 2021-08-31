package users

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"bot.my/utils"
)

//GroupsJoin Вступление в группу
func GroupsJoin(versinonAPIVk, token, groupID, myID string, vkErrUserChan chan error) (err error) {
	log.Printf("Пользоватлеь %v подписывается к %v\n", myID, groupID)
	err = utils.Antiban(myID, "GroupsJoin", 300)
	if err != nil {
		vkErrUserChan <- err
	}
	url := "https://api.vk.com/method/groups.join?v=" + versinonAPIVk + "&access_token=" + token + "&group_id=" + groupID
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
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
	if res.StatusCode != 200 {
		err = fmt.Errorf("Ошибка вступления в группу " + res.Status + " " + string(body))
		return err
	}
	if strings.Contains(string(body), "failed") == true || strings.Contains(string(body), "error") {
		err = fmt.Errorf("Ошибка вступления в группу: %v", string(body))
		return err
	}
	//Добавляем запись в историчну таблицу о вступлении в группу
	err = utils.InsertToDB(fmt.Sprintf("INSERT INTO %v VALUES('%v', '%v', '%v', %v, '%v');", "history", time.Now().UnixNano(), "Вступление в группу", myID, "NOW()", groupID))
	if err != nil {
		vkErrUserChan <- err
	}
	log.Printf("Пользователь %v подписался к группе %v\n", myID, groupID)
	return err
}
