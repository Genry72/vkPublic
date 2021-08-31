package users

import (
	"fmt"
	"log"
	"os"

	"bot.my/utils"
	"bot.my/vk/photo"
)

//SearchClone ищет клона и загружает фотки аватарки
func SearchClone(versinonAPIVk, token, myID, sex string, vkErrUserChan chan error) {
	//Проверяем количество фоток на своей аве, если ноль то ищем клона, если он не определен
	countMyPhoto, _, err := photo.GetPhotoList(versinonAPIVk, token, myID, myID, vkErrUserChan)
	if err != nil {
		vkErrUserChan <- err
	}
	if countMyPhoto == 0 { //Ищем клона
		idsMap, err := FindUsers(token, sex, versinonAPIVk, myID, vkErrUserChan)
		if err != nil {
			vkErrUserChan <- err
		}
		for _, cloneID := range idsMap { //Идем по слайсу cloneID
			countPhoto, urlsPhoto, err := photo.GetPhotoList(versinonAPIVk, token, cloneID, myID, vkErrUserChan) //Получаем url с фотками
			if err != nil {
				vkErrUserChan <- err
			}
			if countPhoto > 15 { //Если на аватарке больше 15 фоток, то нам подходит этот клон
				//Проверяем что клон не занят
				result, err := utils.GetOneRowDB(fmt.Sprintf("select count(*) from users t where t.clone_id = '%v'", cloneID))
				if err != nil {
					vkErrUserChan <- err
				}
				if result != "0" { //Если клон занят то пропускаем
					log.Printf("Для пользователя %v клон %v занят", myID, cloneID)
					continue
				}
				//Дабвляем в БД инфу, что клон занят
				err = utils.InsertToDB(fmt.Sprintf("UPDATE %v SET clone_id = '%v' WHERE user_id = '%v'", "users", cloneID, myID))
				if err != nil {
					vkErrUserChan <- err
				}
				log.Printf("Для пользователя %v найден клон %v", myID, cloneID)
				//Создаем папку для фоток
				photoFolder := fmt.Sprintf("./%v_%v", myID, cloneID)
				os.MkdirAll(photoFolder, 0777)
				//Скачиваем фотки на сервер
				for i, urlPhoto := range urlsPhoto {
					photoFile := fmt.Sprintf("%v/%v.jpg", photoFolder, i)
					err = photo.DownloadPhoto(photoFile, urlPhoto) //Загружаем фотки на сервер
					if err != nil {
						vkErrUserChan <- err
					}
					//После загрузки сразу добавляем фото в профиль
					err = photo.UploadPhoto(cloneID, versinonAPIVk, token, photoFile, myID, vkErrUserChan)
					if err != nil {
						vkErrUserChan <- err
					}
				}
				//Удаляем временный каталог с фотками
				err = os.RemoveAll(photoFolder)
				if err != nil {
					vkErrUserChan <- err
				}
				break //выхолим из цикла, если нашли клона на предыдущем шаге
			}
		}
	}
}
