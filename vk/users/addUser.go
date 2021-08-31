package users

import (
	"fmt"

	// "time"

	"bot.my/utils"
)

//AddUser Прописывает пользователя в БД
// 1. Получаем токен и id
// 2. Создаем новый почтовый ящик
//2. В каждом из сообщест получаем последний объект
// 2. Репоситим последнюю запись последнего сообщения в подписке
func AddUser(username, password string, tokenTamTam, domainMail, pddToken, versinonAPIVk string, vkErrUserChan chan error) (err error) {
	funcName := "AddUser"
	//Получаем токен пользоватлея
	token, userID, err := GetToken(username, password)
	if err != nil {
		err = fmt.Errorf("Ошибка получения токена для пользователя %v: %v", username, err)
		return err
	}
	usersMap, err := utils.GetUsersFromDB() //Получаем список пользователей
	if err != nil {
		return err
	}

	//Проверяем существование пользователя в БД
	for id := range usersMap {
		if id == userID {
			err = fmt.Errorf("Пользователь %v уже существует", userID)
			return err
		}
	}
	//Получаем информацию о пользователе
	firstName, lastName, sex, _, err := GetInfo(userID, token, versinonAPIVk, userID, false, vkErrUserChan)
	if err != nil {
		err = fmt.Errorf("Не удалось получить имя пользвателя по id %v: %v. Функция %v", userID, err, funcName)
		return err
	}
	//Создаем почтовый ящик
	passwordMail := utils.GeneratePWD() //Генерируем пароль
	mailboxID, err := utils.MakeMailBox(domainMail, fmt.Sprintf("mail-%v", userID), passwordMail, pddToken)
	if err != nil {
		return err
	}
	//Вносим в БД инфу о пользователе
	err = utils.InsertToDB(fmt.Sprintf("INSERT INTO %v VALUES('%v', '%v', '%v', %v, '%v', '%v', '%v', '%v', '%v', '%v');", "users", userID, username, password, "NOW()", "Новый пользоватлель", mailboxID, sex, "y", firstName, lastName))
	if err != nil {
		return err
	}

	//Вносим в БД инфу о почтовом ящике
	err = utils.InsertToDB(fmt.Sprintf("INSERT INTO %v VALUES('%v', '%v', '%v', %v);", "mailbox", mailboxID, fmt.Sprintf("mail-%v@%v", userID, domainMail), passwordMail, "NOW()"))
	if err != nil {
		return err
	}
	return err
}
