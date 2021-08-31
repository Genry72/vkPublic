package groups

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"bot.my/utils"
	"bot.my/vk/photo"
	"bot.my/vk/users"
)

//LoadGroups Публикует фото в ленту грурппы
func LoadGroups(groupID, versinonAPIVk, username, password string, vkErrGroupChan chan error) {
	photoFolder := "./foto" //Куда скачиваем фото
	//Создаем папку для будущих фоток
	os.MkdirAll(photoFolder, 0777)
	//Получаем токен для пользователя, так как публикация в сообщество только от пользователя
	token, _, err := users.GetToken(username, password)
	if err != nil {
		err = fmt.Errorf("Ошибка получения токена для пользователя %v: %v", username, err)
		vkErrGroupChan <- err
		return
	}
	//Проверка что это не первая публикация
	countPub, err := utils.GetOneRowDB("select count(*) from history where hist_type = 'Публикация фото на страницу группы' and hist_id like '" + groupID + "%'")
	if err != nil {
		err = fmt.Errorf("Ошибка проверки количества публикаций для группы %v: %v", groupID, err)
		vkErrGroupChan <- err
		return
	}
	if countPub == "0" { //Если ранее не было публикации, то вносим первую запись
		err = utils.InsertToDB(fmt.Sprintf("INSERT INTO history (hist_id, hist_type, user_id, navi_date, group_id, clone_id) VALUES('%v', '%v', '%v', '%v', '%v', '%v');", groupID+"_first", "Публикация фото на страницу группы", "", "NOW()", groupID, ""))
		if err != nil {
			err = fmt.Errorf("Ошибка внесения первой записи в бд для группы %v о публикации фото: %v", groupID, err)
			vkErrGroupChan <- err
			return
		}
	}
Metka: //Метка возврата в начало цикла
	for {
		//Публикация раз в час
		//Время последней публикации string
		lastChangeSTR, err := utils.GetOneRowDB("select navi_date from history where hist_type = 'Публикация фото на страницу группы' and hist_id like '" + groupID + "%' ORDER BY navi_date DESC limit 1")
		if err != nil {
			err = fmt.Errorf("Ошибка получения времени последней публикации фото в группу %v: %v", groupID, err)
			vkErrGroupChan <- err
			return
		}
		lastChange, err := time.Parse("2006-01-02T15:04:05-07:00", lastChangeSTR) //Получаем время из стринга
		if err != nil {
			err = fmt.Errorf("Ошибка парасинга времени: %v", err)
			vkErrGroupChan <- err
		}
		proshloVremeni := time.Since(lastChange).Seconds() //Считаем сколько прошло времени с последнего запроса
		log.Printf("Прошло %v секунд c последней публикации фото", proshloVremeni)
		if proshloVremeni < 3600 { //Не чаще 1 часа
			log.Println("Прошло менее часа с момента последней публикации фото, спим")
			time.Sleep(10 * time.Minute) //Спим 10 минут между запусками
			continue
		}
		for i := 0; i < 1000; i++ {
			//Парсим фишки
			urlList, err := ParseFishki(fmt.Sprintf("%v", i))
			if err != nil {
				vkErrGroupChan <- err
				time.Sleep(5 * time.Minute)
				continue
			}
			//Скачиваем фотки
			for _, url := range urlList {
				_, fileName := path.Split(url)                          //Получаем имя файла
				filePath := fmt.Sprintf("%v/%v", photoFolder, fileName) //Путь для загрузки
				//Проверяем, скачивали ли мы ранее это фотку
				countFoto, err := utils.GetOneRowDB(fmt.Sprintf("select count(*) from history t where t.group_id = '%v' and t.hist_id = '%v'", groupID, groupID+"_"+fileName))
				if err != nil {
					vkErrGroupChan <- err
					continue
				}
				if countFoto != "0" {
					continue
				}
				err = photo.DownloadPhoto(filePath, url)
				if err != nil {
					vkErrGroupChan <- err
					continue
				}
				//Публикуем в группу
				err = UploadFotoToWall(versinonAPIVk, token, filePath, groupID)
				if err != nil {
					err = fmt.Errorf("Ошибка публикации фото в группу %v: %v", groupID, err)
					vkErrGroupChan <- err
					continue
				}
				//Записываем в БД информацию о том что отправили в группу
				err = utils.InsertToDB(fmt.Sprintf("INSERT INTO history (hist_id, hist_type, user_id, navi_date, group_id, clone_id) VALUES('%v', '%v', '%v', '%v', '%v', '%v');", groupID+"_"+fileName, "Публикация фото на страницу группы", "", "NOW()", groupID, ""))
				if err != nil {
					vkErrGroupChan <- err
					continue
				}
				//Удаляем загруженную фотку
				os.Remove(filePath)
				continue Metka
			}
			fmt.Println("ОК")
		}
	}
}
