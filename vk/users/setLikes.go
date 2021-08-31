package users

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"bot.my/utils"
)

//SetLikes Проставляет лайки
func SetLikes(token, versinonAPIVk, myID string, vkErrUserChan chan error) {
	var countLikesWall = 2 //Количество лайков постов
	//Ищем пользоватлей
	ids, err := FindUsers(token, "", versinonAPIVk, myID, vkErrUserChan)
	if err != nil {
		vkErrUserChan <- err
		return
	}
	for _, id := range ids { //Идем по всем пользоватлелям
		//Скипаем елси это наш клон
		cloneID, err := utils.GetOneRowDB("select clone_id from users t where t.user_id = '" + myID + "'")
		if err != nil {
			vkErrUserChan <- err
			continue
		}
		if cloneID == myID {
			continue
		}
		//Переходим к следующему пользоватлею, если у этого дастаточное количество лайков
		countLikes, err := utils.GetOneRowDB("select count(*) from history t where t.user_id = '" + myID + "' and t.hist_id like '%-" + id + "_%' and t.hist_type='Лайк'")
		if err != nil {
			vkErrUserChan <- err
			continue
		}
		countLike, err := strconv.Atoi(countLikes)
		if err != nil {
			vkErrUserChan <- err
			continue
		}
		if countLike > countLikesWall {
			log.Printf("Пользователь %v уже отлайкан", id)
			continue
		}
		//Если ава не лайкнута, то лайкаем
		_, _, _, photoID, err := GetInfo(id, token, versinonAPIVk, myID, false, vkErrUserChan) //олучаем id авы
		if err != nil {
			err = fmt.Errorf("Ошибка получения id авы: %v", err)
		}
		countLikesAva, err := utils.GetOneRowDB(fmt.Sprintf("select count(*) from history t where t.user_id = '%v' and t.hist_id like 'photo-%v' and t.hist_type='Лайк'", myID, photoID))
		if err != nil {
			err = fmt.Errorf("Ошибка получения количества лайков авы")
			vkErrUserChan <- err
			continue
		}
		if countLikesAva == "0" { //Лайкаем аву
			s := strings.Split(string(photoID), "_") //Бьем photoID для получения itemID
			if len(s) < 2 {
				err = fmt.Errorf("Хрень с photoID: %v", photoID)
				vkErrUserChan <- err
				continue
			}
			err = LikesAdd(versinonAPIVk, token, "photo", id, s[1], myID, vkErrUserChan)
			if err != nil {
				vkErrUserChan <- err
				continue
			}
		}
		//Получаем список постов со стены пользователя
		mapPosts, err := WallGet(versinonAPIVk, token, id, myID, vkErrUserChan)
		if err != nil {
			vkErrUserChan <- err
			continue
		}
		//Идем по мапе с постами
		for postID, value := range mapPosts { // Ключь id объекта, значение список: тип объекта и количество лайков
			//Проверяем что лайков этого пользователя не больше 3-х
			countLikes, err = utils.GetOneRowDB("select count(*) from history t where t.user_id = '" + myID + "' and t.hist_id like '%-" + id + "_%' and t.hist_type='Лайк'")
			if err != nil {
				vkErrUserChan <- err
				continue
			}
			countLike, err := strconv.Atoi(countLikes)
			if err != nil {
				vkErrUserChan <- err
			}
			if countLike > countLikesWall {
				// log.Printf("Пользователь %v уже отлайкан", id)
				continue
			}
			// //Проверяем что ранее мы не лайкали этот пост
			// countLikesThisPost, err := utils.GetOneRowDB(fmt.Sprintf("select count(*) from history t where t.user_id = '%v' and t.hist_id = '%v' and t.hist_type='Лайк'", myID, value[0]+"-"+id+"_"+postID))
			// if err != nil {
			// 	vkErrUserChan <- err
			// 	continue
			// }
			// if countLikesThisPost != "0" {
			// 	log.Printf("Пост %v уже лайкался пользователем %v", postID, myID)
			// 	continue
			// }
			//Лайкаем пост
			err = LikesAdd(versinonAPIVk, token, value[0], id, postID, myID, vkErrUserChan)
			if err != nil {
				vkErrUserChan <- err
				continue
			}
		}
	}
}
