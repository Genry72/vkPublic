package users

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bot.my/utils"
)

//WallRepost Делает репост указанного поста
func WallRepost(versinonAPIVk, token, object, myID string, vkErrUserChan chan error) (err error) {
	log.Printf("Делаем репост %v пользоватлем %v", object, myID)
	err = utils.Antiban(myID, "WallRepost", 7200) //Раз в 2 часа
	if err != nil {
		return err
	}
	//Проверяем что за сегодня было меньше 10 репостов
	countRepostSTR, err := utils.GetOneRowDB(fmt.Sprintf("select count(*) from history t where t.user_id = '%v' and t.navi_date > 'today' and t.hist_type = 'Репост'", myID))
	countRepost, err := strconv.Atoi(countRepostSTR)
	if err != nil {
		return err
	}
	if countRepost > 10 {
		log.Println("Количество репостов больше 10")
		time.Sleep(30 * time.Minute)
		return err
	}
	//Проверяем что этот репост мы ранее не делали
	countThisRepost, err := utils.GetOneRowDB(fmt.Sprintf("select count(*) from history t where t.user_id = '%v' and t.hist_type = 'Репост' and t.hist_id = '%v'", myID, object))
	if err != nil {
		return err
	}
	if countThisRepost != "0" {
		log.Printf("Этот репост %v мы делали раньше", object)
		return err
	}
	url := "https://api.vk.com/method/wall.repost?v=" + versinonAPIVk + "&access_token=" + token + "&object=" + object
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
		err = fmt.Errorf("Ошибка репоста " + res.Status + " " + string(body))
		return err
	}
	if strings.Contains(string(body), "failed") == true || strings.Contains(string(body), "error") {
		err = fmt.Errorf("Ошибка репоста: %v", string(body))
		return err
	}
	//Добавляем запись в историчну таблицу о выполнении запроса
	err = utils.InsertToDB(fmt.Sprintf("INSERT INTO %v VALUES('%v', '%v', '%v', %v);", "history", object, "Репост", myID, "NOW()"))
	if err != nil {
		vkErrUserChan <- err
	}
	log.Printf("Пользоватлем %v сделан репост %v объекта", myID, object)
	return err
}
