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

//LikesAdd Лайкает указанный пост
func LikesAdd(versinonAPIVk, token, typeObg, ownerID, itemID, myID string, vkErrUserChan chan error) (err error) {
	log.Printf("Лайкаем пост %v пользоватлея %v пользоватлем %v", itemID, ownerID, myID)
	maxLikes := 150 //Максимальное количество лайков в день
	err = utils.Antiban(myID, "LikesAdd", 300)
	if err != nil {
		vkErrUserChan <- err
	}
	//Не больше 100 лайков в сутки
	//Проверяем что за сегодня было меньше 100 лайков
	countRepostSTR, err := utils.GetOneRowDB(fmt.Sprintf("select count(*) from history t where t.user_id = '%v' and t.navi_date > 'today' and t.hist_type = 'Лайк'", myID))
	countRepost, err := strconv.Atoi(countRepostSTR)
	if err != nil {
		return
	}
	if countRepost > maxLikes {
		log.Printf("Количество лайков больше %v", maxLikes)
		time.Sleep(30 * time.Minute)
		return
	}
	//Проверяем что ранее этот объект не лайкался этим польоватлем
	countLikesThisPost, err := utils.GetOneRowDB("select count(*) from history t where t.user_id = '" + myID + "' and t.hist_id = '" + typeObg + "-" + ownerID + "_" + itemID + "' and t.hist_type='Лайк'")
	if err != nil {
		vkErrUserChan <- err
		return
	}
	if countLikesThisPost != "0" {
		log.Printf("Пост %v уже лайкался пользователем %v", itemID, myID)
		return
	}
	url := "https://api.vk.com/method/likes.add?v=" + versinonAPIVk + "&access_token=" + token + "&type=" + typeObg + "&owner_id=" + ownerID + "&item_id=" + itemID
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
		err = fmt.Errorf("Ошибка при лайке " + res.Status + " " + string(body))
		return err
	}
	if strings.Contains(string(body), "failed") == true || strings.Contains(string(body), "error") {
		if strings.Contains(string(body), "This profile is private") == true { //Не выводим ошибку по рпиватному профилю
			log.Printf("Профиль аккаунта %v приватный, скипаем", ownerID)
			return err
		}
		if strings.Contains(string(body), "photo is private") == true {
			log.Printf("Фотка аккаунта %v приватная, скипаем", ownerID)
			return err
		}
		if strings.Contains(string(body), "no access to this group") == true {
			log.Printf("Нет доступа к группе %v, скипаем", ownerID)
			return err
		}
		if strings.Contains(string(body), "object not found") == true {
			log.Printf("object not found %v, скипаем", ownerID)
			return err
		}
		if strings.Contains(string(body), "Access denied") == true {
			log.Printf("Access denied %v, скипаем", ownerID)
			return err
		}
		if strings.Contains(string(body), "Flood control") == true {
			log.Printf("Flood control %v, скипаем", ownerID)
			time.Sleep(20 * time.Minute) // Спим 20 минут при подозрении в флуде
			return err
		}
		if strings.Contains(string(body), "Captcha needed") == true {
			log.Printf("Captcha needed %v, скипаем", ownerID)
			time.Sleep(20 * time.Minute) // Спим 20 минут при при запросе капчи
			return err
		}
		err = fmt.Errorf("Ошибка при лайке, запрос: %v: %v", url, string(body))
		time.Sleep(1 * time.Minute) // Спим минуту
		return err
	}
	//Добавляем запись в историчну таблицу о лайке
	err = utils.InsertToDB(fmt.Sprintf("INSERT INTO %v VALUES('%v', '%v', '%v', %v);", "history", fmt.Sprintf("%v-%v_%v", typeObg, ownerID, itemID), "Лайк", myID, "NOW()"))
	if err != nil {
		vkErrUserChan <- err
	}
	log.Printf("Лайкнули пост %v пользоватлея %v пользоватлем %v", itemID, ownerID, myID)
	return err
}
