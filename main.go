package main

import (
	"time"
	// "fmt"
	"fmt"
	"net/http"

	"bot.my/tamtam"

	"bot.my/telegram"
	"bot.my/utils"
	"bot.my/vk/groups"
	"bot.my/vk/users"
	log "github.com/sirupsen/logrus"
)

//Акк VK http://vk-retriv.ru/
//Домен получен на https://my.freenom.com/cart.php?a=checkout
//Серты настраивал по инструкции https://invs.ru/support/chastie-voprosy/ustanovka-besplatnogo-ssl-ot-cloudflare/
//На самом мервере сомоподписные openssl req -newkey rsa:2048 -sha256 -nodes -x509 -days 365 -keyout YOURPRIVATE.key -out YOURPUBLIC.pem -subj "/C=RU/ST=Saint-Petersburg/L=Saint-Petersburg/O=Example Inc/CN=cloudvsemenov.tk"
// 1
// www.cloudvsemenov.tk/*
// Forwarding URL (Status Code: 301 - Permanent Redirect, Url: https://cloudvsemenov.tk/$1)
// 2
// https://cloudvsemenov.tk/*
// Always Use HTTPS
//https://dash.cloudflare.com/6fff0487cca407f7e516e9b7b7109b2d/cloudvsemenov.tk/page-rules
func main() {
	//Параметры для создания почтовых ящиков
	// domainMail := "dripmen.ru"
	// pddToken := "H5A5UBHBKOORTMNIOH4JOVBKWMUWVT3436PRWWSTJPL5UWIA7LWQ" //https://pddimp.yandex.ru/api2/admin/get_token?domain_name=dripmen.ru
	tokenTamTam := "XXXX"
	portWeb := "443"
	urlWebHookTamTam := "https://domen.ru:" + portWeb + "/" + tokenTamTam
	tokenTelegram := "XXXX"
	// urlWebHookTelegram := "https://cloudvsemenov.tk:" + portWeb + "/" + tokenTelegram
	keyAccesVk := "XXXX"
	passAPIVk := "XXXX" //Пароль из настроек сообщества, для проверки что запрос точно из ВК
	versinonAPIVk := "5.89"               //Версия апишки VK https://vk.com/dev/versions
	publicCert := "YOURPUBLIC.pem"
	privKey := "YOURPRIVATE.key"
	//Логирование
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05.999999"
	customFormatter.FullTimestamp = true
	log.SetFormatter(customFormatter)
	//Отправка сообщения о запуске
	// err := tamtam.SendToChat(tokenTamTam, "Я стартанул!", "", "Валентин Семенов", 0)
	// if err != nil {
	// 	log.Error("Ошибка отправки сообщения в там-там: " + err.Error())
	// }
	err := telegram.SendMsg(tokenTelegram, "Я стартанулл!", 153123826)
	if err != nil {
		log.Error("Ошибка отправки сообщения в телеграмм: " + err.Error())
	}
	log.Info("Привет, я стартанулл!")
	msgTamtamChan := make(chan map[int64]string, 10000)   //Канал для входящих сообщений из тамтам
	tamTamErrChan := make(chan error, 10000)              //Канал для фиксации ошибок из тамтам
	msgTelegramChan := make(chan map[int64]string, 10000) //Канал для входящих сообщений из телеги
	telegramErrChan := make(chan error, 10000)            //Канал для фиксации ошибок из телеги
	vkMsgGroupChan := make(chan string, 10000)            //В канале VK только само сообщение для группы
	vkErrGroupChan := make(chan error, 10000)             //Канал для фиксации ошибок из контакта для группы
	// usernameVK := "992985562627"                          //Логин пользователя VK +79200180048
	// passwordVK := "Nifi2in8u"                             //Пароль пользователя vk Nifi2in8u
	// usernameVK := "96597288036"        //Логин пользователя VK
	// passwordVK := "rVlsk6NOqA"         //Пароль пользователя vk
	vkMsgUserChan := make(chan string, 10000) //В канале VK только само сообщение для пользователя
	vkErrUserChan := make(chan error, 10000)  //Канал для фиксации ошибок из контакта для пользователя
	//Ловим сообщения tam-tam
	go func() {
		for {
			m := <-msgTamtamChan
			for chatID, msg := range m {
				log.Info("Получено сообщение из там-там: " + msg)
				err := tamtam.SendToChat(tokenTamTam, msg, "", "", chatID)
				if err != nil {
					log.Error("Ошибка отправки сообщения в там-там: " + err.Error())
				}
			}

		}
	}()
	//Ловим ошибки tam-tam
	go func() {
		for {
			err := <-tamTamErrChan
			if err != nil {
				log.Error("там-там: " + err.Error())
			}
		}
	}()
	//Ловим сообщения telegram
	go func() {
		for {
			m := <-msgTelegramChan
			for chatID, msg := range m {
				log.Info("Получено сообщение из telegram: " + msg)
				err := telegram.SendMsg(tokenTelegram, msg, chatID)
				if err != nil {
					log.Error("Ошибка отправки сообщения в телеграм: " + err.Error())
				}
			}

		}
	}()
	//Ловим ошибки telegram
	go func() {
		for {
			err := <-telegramErrChan
			if err != nil {
				log.Error("telegram: " + err.Error())
			}
		}
	}()
	//Ловим сообщения ВК и отправляем в там-там для группы
	go func() {
		for {
			msg := <-vkMsgGroupChan
			log.Info("Получено уведомление из контакта (группа): " + msg)
			// err := tamtam.SendToChat(tokenTamTam, msg, "", "Валентин Семенов", 0)
			// if err != nil {
			// 	log.Error("Ошибка отправки сообщения в там-там: " + err.Error())
			// }
			err := telegram.SendMsg(tokenTelegram, msg, 153123826)
			if err != nil {
				log.Error("Ошибка отправки сообщения в телеграмм: " + err.Error())
			}
		}
	}()
	//Ловим ошибки ВК для группы
	go func() {
		for {
			err := <-vkErrGroupChan
			if err != nil {
				log.Error("контакт группа: " + err.Error())
				err := telegram.SendMsg(tokenTelegram, "контакт группа: "+err.Error(), 153123826)
				if err != nil {
					err = fmt.Errorf("Ошибка отправки сообщения в телеграм: %v", err)
					log.Error(err)
				}
			}
		}
	}()
	//Ловим сообщения ВК и отправляем в там-там для пользователя
	go func() {
		for {
			msg := <-vkMsgUserChan
			log.Info("Получено уведомление из контакта для пользователя: " + msg)
			// err := tamtam.SendToChat(tokenTamTam, msg, "", "Валентин Семенов", 0)
			// if err != nil {
			// 	log.Error("Ошибка отправки сообщения в там-там: " + err.Error())
			// }
			err := telegram.SendMsg(tokenTelegram, msg, 153123826)
			if err != nil {
				err = fmt.Errorf("Ошибка отправки сообщения в телеграм: %v", err)
				log.Error(err)
			}

		}
	}()
	//Ловим ошибки ВК для пользователя
	go func() {
		for {
			err := <-vkErrUserChan
			if err != nil {
				log.Error("контакт пользователь: " + err.Error())
				// err := tamtam.SendToChat(tokenTamTam, "контакт пользователь: "+err.Error(), "", "Валентин Семенов", 0)
				// if err != nil {
				// 	log.Error("Ошибка отправки сообщения в там-там: " + err.Error())
				// }
				err := telegram.SendMsg(tokenTelegram, "контакт пользователь: "+err.Error(), 153123826)
				if err != nil {
					err = fmt.Errorf("Ошибка отправки сообщения в телеграм: %v", err)
					log.Error(err)
				}
			}
		}
	}()
	//Вызываем функцию получения сообщений из тамт-там
	go func() {
		err = tamtam.GetMsg(msgTamtamChan, tamTamErrChan, tokenTamTam, urlWebHookTamTam)
		if err != nil {
			log.Error("там-там: " + err.Error())
		}
	}()
	//Вызываем функцию получения сообщений из telegram
	go func() {
		telegram.GetMsg(msgTelegramChan, telegramErrChan, tokenTelegram)
	}()
	// if err != nil {
	// 	log.Error("telegram: " + err.Error())
	// }
	//Вызываем функцию получения сообщений из контакта для группы
	go func() {
		groups.GetMsg(vkMsgGroupChan, vkErrGroupChan, keyAccesVk, passAPIVk, versinonAPIVk)
	}()
	//Создаем пользователя в БД
	// err = users.AddUser(usernameVK, passwordVK, tokenTamTam, domainMail, pddToken, versinonAPIVk, vkErrUserChan)
	// if err != nil {
	// 	vkErrUserChan <- err
	// }
	go func() {
		usersMap, err := utils.GetUsersFromDB() //Получаем список пользователей
		if err != nil {
			vkErrUserChan <- err
			return
		}
		//Прокачиваем пользователя и получаем для него обновления
		for myID, value := range usersMap {
			go func(myID string, value []string) {
				err = users.LoadUser(value[0], value[1], value[7], value[8], myID, value[5], versinonAPIVk, vkErrUserChan, vkMsgUserChan)
				if err != nil {
					vkErrUserChan <- err
					return
				}
			}(myID, value)

		}
	}()
	go func() {
		groupsMap, err := utils.GetGroupsFromDB() //Получаем список групп
		if err != nil {
			vkErrGroupChan <- err
			return
		}
		//Прокачиваем группу
		for groupID := range groupsMap {
			if groupsMap[groupID][5] == "" {
				err = fmt.Errorf("Не задан админ для группы %v", groupID)
				vkErrGroupChan <- err
				continue
			}
			//Получаем логин админа по id из БД
			username, err := utils.GetOneRowDB(fmt.Sprintf("select username from users t where t.user_id = '%v'", groupsMap[groupID][5]))
			if err != nil {
				err = fmt.Errorf("Ошибка получения логина админа группы: %v", err)
				vkErrGroupChan <- err
				continue
			}
			//Получаем пароль админа по id из БД
			password, err := utils.GetOneRowDB(fmt.Sprintf("select password from users t where t.user_id = '%v'", groupsMap[groupID][5]))
			if err != nil {
				err = fmt.Errorf("Ошибка получения логина админа группы: %v", err)
				vkErrGroupChan <- err
				continue
			}

			go func(groupID string) {
				log.Info("Начинаем прокачку группы " + groupID)
				groups.LoadGroups(groupID, versinonAPIVk, username, password, vkErrGroupChan)
			}(groupID)
		}
	}()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { //Слушаем на всех урлах для отображения кривых запросов
		fmt.Fprintf(w, "Кто ты? Что тебе нужно?!")
		ips := fmt.Sprintf("X-Forwarded-For: %v, X-REAL-IP: %v, RemoteAddr: %v\n", r.Header.Get("X-REAL-IP"), r.Header.Get("X-Forwarded-For"), r.RemoteAddr)
		log.Error(fmt.Sprintf("Запрос некорректного url %v c %v. Передали %v", r.RequestURI, ips, r))
	})
	time.Sleep(1 * time.Second)                                                 //Перед выходом ждем отправку сообщений
	log.Fatal(http.ListenAndServeTLS(":"+portWeb+"", publicCert, privKey, nil)) //Слушаем порт для всех вебхуков
}
