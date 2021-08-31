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

//AddFrends Добавление друзей
func AddFrends(myID, token, versinonAPIVk string, vkErrUserChan chan error) {
	//Селектим из бд пользоватлей которым проставили лайки
	log.Printf("Запустили добавление в друзья пользователю %v", myID)
	userSlice, err := utils.GetLikedUsers(myID)
	if err != nil {
		vkErrUserChan <- err
		return
	}
	//Идем по слайсу с потенциальными друзьями
	for _, user := range userSlice {
		//Скипаем елси это наш клон
		cloneID, err := utils.GetOneRowDB("select clone_id from users t where t.user_id = '" + myID + "'")
		if err != nil {
			vkErrUserChan <- err
		}
		if cloneID == myID {
			continue
		}
		//Добавляем в друзья
		err = addFrendsAPI(versinonAPIVk, token, user, myID, vkErrUserChan, false)
		if err != nil {
			vkErrUserChan <- err
		}
	}
	log.Printf("Закончили добавление в друзья пользователю %v", myID)
}

//addFrendsAPI Отправка заявки на добавление в друзья. Так же принимаем заявки
func addFrendsAPI(versinonAPIVk, token, userID, myID string, vkErrUserChan chan error, inFrend bool) (err error) {
	//Проверяем что этого пользователя ранее не добавляли в друзья
	countAddThisFrend, err := utils.GetOneRowDB("select count(*) from history t where t.user_id = '" + myID + "' and t.hist_type = 'Заявка в друзья' and hist_id like '%" + userID + "%'")
	if err != nil {
		return err
	}
	if countAddThisFrend != "0" {
		// fmt.Println("countAddThisFrend: " + countAddThisFrend)
		log.Printf("Пользоватлея %v уже добавляли в друзья пользователю %v", userID, myID)
		// time.Sleep(10 * time.Minute)
		return err
	}
	log.Printf("Пользователю %v добавляем в друзья пользоватля %v", myID, userID)

	if inFrend == false { //Проверку на антибан выполняем только если это не входящие друзья
		err = utils.Antiban(myID, "addFrendsAPI", 7200) //Раз в 2 часа
		if err != nil {
			vkErrUserChan <- err
		}
		//Проверяем что за сегодня было меньше 15 запросов
		countRepostSTR, err := utils.GetOneRowDB("select count(*) from history t where t.user_id = '" + myID + "' and t.navi_date > 'today' and t.hist_type = 'Заявка в друзья' and t.hist_id not like 'infrend-%'") //Исключаем из проверки входящие заявки
		if err != nil {
			vkErrUserChan <- err
		}
		countRepost, err := strconv.Atoi(countRepostSTR)
		if err != nil {
			vkErrUserChan <- err
			return err
		}
		if countRepost > 15 {
			log.Println("Количество запросов в друзья больше 15")
			time.Sleep(30 * time.Minute)
			return err
		}
	}
	url := "https://api.vk.com/method/friends.add?v=" + versinonAPIVk + "&access_token=" + token + "&user_id=" + userID + "&follow=0"
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
		err = fmt.Errorf("Ошибка добавления в друзья " + res.Status + " " + string(body))
		return err
	}
	if strings.Contains(string(body), "failed") == true || strings.Contains(string(body), "error") {
		if strings.Contains(string(body), "Cannot add this user to friends as user not found") == true {//Не выводим ошибки по удаленным или заблокированным пользоватлеям
			log.Printf("Пользователь %v не существует, скипаем", userID)
			return err
		}
		err = fmt.Errorf("добавления в друзья: %v", string(body))
		time.Sleep(5 * time.Minute)
		return err
	}
	//Добавляем запись в историчну таблицу о добавлении в друзья
	var info string
	if inFrend == false { //Если это не входящаяя заявка
		info = "addfrend"
	} else {
		info = "infrend"
	}
	err = utils.InsertToDB(fmt.Sprintf("INSERT INTO %v VALUES('%v', '%v', '%v', %v);", "history", info+"-"+myID+"_"+userID, "Заявка в друзья", myID, "NOW()"))
	if err != nil {
		vkErrUserChan <- err
	}
	log.Printf("Пользователю %v добавили в друзья пользоватля %v", myID, userID)
	return err
}

//InFrends Проверяем входящие заявки в друзья и дабавляем
func InFrends(versinonAPIVk, token, myID string, vkErrUserChan chan error) (err error) {
	// err = utils.Antiban(myID, "InFrends", 7200) //Раз в 2 часа
	// if err != nil {
	// 	vkErrUserChan <- err
	// }
	url := "https://api.vk.com/method/friends.getRequests?v=" + versinonAPIVk + "&access_token=" + token
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
		err = fmt.Errorf("Ошибка запроса входящих предложений дружить " + res.Status + " " + string(body))
		return err
	}
	if strings.Contains(string(body), "failed") == true || strings.Contains(string(body), "error") {
		err = fmt.Errorf("Ошибка запроса входящих предложений дружить: %v", string(body))
		return err
	}
	frends := inFrendsStruct{}
	err = json.Unmarshal(body, &frends)
	if err != nil {
		err = fmt.Errorf("Ошибка парсинга боди на запрос friends.getRequests: %v/ Боди: %v", err, string(body))
		return err
	}
	if frends.Response.Count != 0 {
		log.Printf("Для пользователя %v %v заявок в друзья", myID, frends.Response.Count)
		for _, frendID := range frends.Response.Items { //Идем по слайсу с id
			//Добавляем в друзья
			err = addFrendsAPI(versinonAPIVk, token, fmt.Sprintf("%v", frendID), myID, vkErrUserChan, true)
			if err != nil {
				vkErrUserChan <- err
			}
			time.Sleep(1 * time.Minute)
		}
	}
	// token = frends.AccessToken
	return err
}

type inFrendsStruct struct {
	Response struct {
		Count       int   `json:"count"`
		Items       []int `json:"items"`
		CountUnread int   `json:"count_unread"`
		LastViewed  int   `json:"last_viewed"`
	} `json:"response"`
}
