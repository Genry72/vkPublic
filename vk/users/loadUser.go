package users

import (
	"fmt"
	"log"
	"time"
)

//Прокачиваем пользоватля
//1. Проверяем количество фоток у пользователя, если меньше трех, то удаляем и клюнируем с другого акка
//2. Ищем клона
//1. Подписываемся на все подписки из списка
//2. Вносим в бд историчную запись о подписке
// 1. Получаем список подписок
//2. В каждом из сообщест получаем последний объект
// 2. Репоситим последнюю запись последнего сообщения в подписке

//LoadUser Прокачка пользователя и получение обновлений
func LoadUser(username, password, firstName, lastName, myID, sex, versinonAPIVk string, vkErrUserChan chan error, vkMsgUserChan chan string) (err error) {
	log.Printf("Начинаем прокачку пользователя %v myID = %v", username, myID)
	//300 лайков, постов и репостов в сутки
	//Получаем токен пользоватлея
	// user.username, user.password, user.naviDate.String(), user.naviUser, user.mailboxID, user.sex
	token, _, err := GetToken(username, password)
	if err != nil {
		err = fmt.Errorf("Ошибка получения токена для пользователя %v: %v", username, err)
		return err
	}
	// //Получаем имя и фамилию
	// firstName, lastName, sex, err := GetInfo(myID, token, versinonAPIVk, false, queueChanZapros, queueChanOtvet)
	// if err != nil {
	// 	err = fmt.Errorf("Не удалось получить имя пользвателя по id %v: %v", myID, err)
	// 	vkErrUserChan <- err
	// 	return
	// }
	//Получаем обновления для пользователя VK
	GetUpdateForUser(versinonAPIVk, token, firstName, lastName, myID, vkErrUserChan, vkMsgUserChan)

	go func() { //Ищем клона, если у нас 0 фоток в профиле
		SearchClone(versinonAPIVk, token, myID, sex, vkErrUserChan)
	}()

	Subscriber(sex, versinonAPIVk, token, myID, vkErrUserChan) //Подписываем пользоватле на указанные группы
	//Делаем репосты
	go func() {
		for {
			//Получаем список объектов для репоста
			postsIDs, err := NewsfeedGet(versinonAPIVk, token, myID, vkErrUserChan)
			if err != nil {
				vkErrUserChan <- err
				continue
			}
			for _, object := range postsIDs { //Репостим
				err = WallRepost(versinonAPIVk, token, object, myID, vkErrUserChan)
				if err != nil {
					vkErrUserChan <- err
					continue
				}
			}
		}
	}()
	//Проставлем лайки 1 на аву и 2 на пост со стены
	// go func() {
	// 	for {
	// 		SetLikes(token, versinonAPIVk, myID, vkErrUserChan)
	// 	}
	// }()

	// Добавляем в друзья тех кому поставили лайк
	// go func() {
	// 	for {
	// 		AddFrends(myID, token, versinonAPIVk, vkErrUserChan)
	// 	}
	// }()

	//Принимаем входящие заявкив друзья
	go func() {
		for {
			InFrends(versinonAPIVk, token, myID, vkErrUserChan)
			time.Sleep(1 * time.Hour)
		}
	}()
	//Лайкаем через сервис freelikes.online
	go func() {
		err = FreeLikes(versinonAPIVk, token, myID, username, password, vkErrUserChan)
		if err != nil {
			vkErrUserChan <- err
		}
	}()
	//Проверяем количество друзей пользователя
	go func() {
		for {
			dt := time.Now()
			b := fmt.Sprintf(dt.Format("15"))
			if b == "19" {
				count, err := FrendsGet(versinonAPIVk, token)
				if err != nil {
					vkErrUserChan <- err
				}
				vkMsgUserChan <- fmt.Sprintf("У пользователя %v подписчиков", count)
				time.Sleep(2 * time.Hour)
			}
			time.Sleep(30 * time.Minute)
		}
	}()

	// log.Printf("Прогрузка пользователя %v завершена", myID)
	return err
}
