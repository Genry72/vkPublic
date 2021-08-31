package users

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//GetUpdateForUser получение обновлений для пользователя
func GetUpdateForUser(versinonAPIVk, token, myfirstName, mylastName, myID string, vkErrUserChan chan error, vkMsgUserChan chan string) {
	var ts int
	var server string
	var key string
	//Получаем параметры подключения к долгим запросам на обновление
	key, server, ts, err := getLongPollServer(token, versinonAPIVk)
	if err != nil {
		err = fmt.Errorf("Ошибка получения параметров подключения к серверу догих запросов: %v", err)
		vkErrUserChan <- err
		return
	}
	var tsNew int
	var text string
	timerKey := time.Now() //Таймер срока действия ключа (действует 1 час)
	//Выполняем в цикле получение обновления
	go func() {
		for {
			timerKeyElapsed := time.Since(timerKey).Minutes() //сколько прошло времени после получения клча
			if timerKeyElapsed > 50 {                         //Перепалучаем ключ и адрес сервера при истечении 50 минут
				key, server, ts, err = getLongPollServer(token, versinonAPIVk)
				if err != nil {
					err = fmt.Errorf("Ошибка получения параметров подключения к серверу догих запросов: %v", err)
					vkErrUserChan <- err
					return
				}
				timerKey = time.Now() //Обнуляем счетчик
			}
			tsNew, text, err = getUpdates(key, server, ts)
			if err != nil {
				if strings.Contains(err.Error(), "Ночная ошибка") != true {
					vkErrUserChan <- err //Если это не ночная ошибка то отправляем в канал
				}
				continue
			}
			// fmt.Printf("ts: %v, text:%v",ts, text)
			// fmt.Printf("tsNew: %v, text:%v",tsNew, text)
			var mapupdates = map[float64]string{
				1:  "Замена флагов сообщения (FLAGS:=$flags).",
				2:  "Установка флагов сообщения (FLAGS|=$mask).",
				3:  "Сброс флагов сообщения (FLAGS&=~$mask).",
				4:  "Добавление нового сообщения.",
				5:  "Редактирование сообщения.",
				6:  "Прочтение всех входящих сообщений в $peer_id, пришедших до сообщения с $local_id.",
				7:  "Прочтение всех исходящих сообщений в $peer_id, пришедших до сообщения с $local_id.",
				8:  "Друг $user_id стал онлайн. $extra не равен 0, если в mode был передан флаг 64. В младшем байте (остаток от деления на 256) числа extra лежит идентификатор платформы (см. 7. Платформы). $timestamp — время последнего действия пользователя $user_id на сайте.",
				9:  "Друг $user_id стал оффлайн ($flags равен 0, если пользователь покинул сайт и 1, если оффлайн по таймауту ) . $timestamp — время последнего действия пользователя $user_id на сайте.",
				10: "Сброс флагов диалога $peer_id. Соответствует операции (PEER_FLAGS &= ~$flags). Только для диалогов сообществ.",
				11: "Замена флагов диалога $peer_id. Соответствует операции (PEER_FLAGS:= $flags). Только для диалогов сообществ.",
				12: "Установка флагов диалога $peer_id. Соответствует операции (PEER_FLAGS|= $flags). Только для диалогов сообществ.",
				13: "Удаление всех сообщений в диалоге $peer_id с идентификаторами вплоть до $local_id.",
				14: "Восстановление недавно удаленных сообщений в диалоге $peer_id с идентификаторами вплоть до $local_id.",
				20: "Изменился $major_id в диалоге $peer_id.",
				21: "Изменился $minor_id в диалоге $peer_id.",
				51: "Один из параметров (состав, тема) беседы $chat_id были изменены. $self — 1 или 0 (вызваны ли изменения самим пользователем).",
				52: "Изменение информации чата $peer_id с типом $type_id, $info — дополнительная информация об изменениях, зависит от типа события. См. 3.2. Дополнительные поля чатов",
				61: "Пользователь $user_id набирает текст в диалоге. Событие приходит раз в ~5 секунд при наборе текста. $flags = 1.",
				62: "Пользователь $user_id набирает текст в беседе $chat_id.",
				63: "Пользователи $user_ids набирают текст в беседе $peer_id. Максимально передается пять участников беседы, общее количество печатающих указывается в $total_count. $ts — время генерации этого события",
				64: "Пользователи $user_ids записывают аудиосообщение в беседе $peer_id.",
				70: "Пользователь $user_id совершил звонок с идентификатором $call_id.",
				80: "Счетчик в левом меню стал равен $count.",
				114: "	Изменились настройки оповещений. $peer_id — идентификатор чата/собеседника, '$sound — 1/0, включены/выключены звуковые оповещения, $disabled_until — выключение оповещений на необходимый срок (-1: навсегда, ''0",
			}
			ts = tsNew
			//Парисим боди
			m := getUpdateStruct{}
			err = json.Unmarshal([]byte(text), &m)
			if err != nil {
				err = fmt.Errorf("Ошибка парсинга боди на запрос getUpdates: %v Боди: %v", err, string(text))
				vkErrUserChan <- err
			}
			for _, update := range m.Updates {
				c := update[0].(float64) //Определенному полю интерфейса присваиваем тип float64
				b := "Тип: " + mapupdates[c]
				if c == 4 { //Добавление нового сообщения
					dialog := strconv.FormatInt(int64(int(update[3].(float64))), 10) //С каким пользователем происходит диалог
					idMsg := update[1].(float64)                                     //id сообщения
					firstName, lastName, _, _, err := GetInfo(dialog, token, versinonAPIVk, myID, false, vkErrUserChan)
					if err != nil {
						err = fmt.Errorf("Не удалось получить имя пользвателя по id %v: %v", dialog, err)
					}
					textMsg, err := getHistory(dialog, myID, token, versinonAPIVk, idMsg)
					if err != nil {
						err = fmt.Errorf("Не удалось получить историю сообщений dialog %v: %v", dialog, err)
					}
					if textMsg != "" { //От самого себя либо пустое
						vkMsgUserChan <- fmt.Sprintf("Для бота %v %v получено сообщение от пользователя %v %v c текстом: %v", myfirstName, mylastName, firstName, lastName, textMsg)
					}
					continue
				}
				if c == 61 {
					dialog := strconv.FormatInt(int64(int(update[1].(float64))), 10) //С каким пользователем происходит диалог
					firstName, lastName, _, _, err := GetInfo(dialog, token, versinonAPIVk, myID, false, vkErrUserChan)
					if err != nil {
						err = fmt.Errorf("Не удалось получить имя пользвателя по id %v: %v", dialog, err)
					}
					vkMsgUserChan <- fmt.Sprintf("Пользователь %v %v набираем текст в диалоге с %v %v", firstName, lastName, myfirstName, mylastName)
					continue
				}
				if c == 8 { //Друг стал онлайн
					continue
					// dialog := int(update[1].(float64)) //С каким пользователем происходит диалог
					// userName, err := GetUserGroupByID(dialog*(-1), token, versinonAPIVk, false, queueChanZapros, queueChanOtvet)
					// if err != nil {
					// 	err = fmt.Errorf("Не удалось получить имя пользвателя по id %v: %v", dialog, err)
					// }
					// vkMsgUserChan <- fmt.Sprintf("Друг %v ползователя %v стал онлайн", userName, myName)
					// continue
				}
				if c == 9 { //Друг стал офлайн
					continue
					// dialog := int(update[1].(float64)) //С каким пользователем происходит диалог
					// userName, err := GetUserGroupByID(dialog*(-1), token, versinonAPIVk, false, queueChanZapros, queueChanOtvet)
					// if err != nil {
					// 	err = fmt.Errorf("Не удалось получить имя пользвателя по id %v: %v", dialog, err)
					// }
					// vkMsgUserChan <- fmt.Sprintf("Друг %v ползователя %v стал офлайн", userName, myName)
					// continue
				}
				if c == 81 { //сообщение не прочитано
					continue
				}
				fmt.Println(text)
				vkMsgUserChan <- b //Отправляем все обновления (для дебага)
			}

		}
	}()

}

//LimitAPIRequest Устанавливаем лимит для запросов в апи ВК
func LimitAPIRequest(queueChanZapros, queueChanOtvet chan string, vkErrUserChan chan error) {
	var countRest = 0            //Счетчик количества запросов
	rand.Seed(time.Now().Unix()) //Seed для рандома
	timeLast := time.Now()       //Запоминаем время последнего запроса
	for {
		//Проверяем количество сообщений в каналах
		if len(queueChanZapros) > 2 || len(queueChanOtvet) > 2 {
			err := fmt.Errorf("В канале queueChanZapros %v сообщений, в канале queueChanOtvet %v сообщений", len(queueChanZapros), len(queueChanOtvet))
			vkErrUserChan <- err
		}
		<-queueChanZapros
		//Делаем проверку. Был ли последний запрос более секунды назад
		if countRest == 3 { //Все проверки производим перед выполнением четвертого по счету запросом
			proshloVremeni := time.Since(timeLast).Seconds() //Считаем сколько прошло времени с последнего запроса
			if proshloVremeni < 1 {                          //Последний запрос был менее секунды назад
				// log.Printf("Спим %v", 5)
				time.Sleep(time.Duration(1+rand.Intn(5)) * time.Second) //Рандом от 1 до 10
				//Сбрасываем счетчики
				countRest = 0
			}
		}
		timeLast = time.Now()
		countRest++                     //При получении увеличиваем счетчик
		queueChanOtvet <- "Согласовано" //Даем добро если это не четвертый запрос
	}
}

//Получаем адрес сервера, ключь и метку для получения последнего обновления
func getLongPollServer(token, versinonAPIVk string) (key, server string, ts int, err error) {
	url := "https://api.vk.com/method/messages.getLongPollServer?access_token=" + token + "&lp_version=3&v=" + versinonAPIVk
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return key, server, ts, err
	}
	res, err := client.Do(req)
	if err != nil {
		return key, server, ts, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if strings.Contains(string(body), "error") == true { //Если в боди вернулась ошибка то прекращаем
		err = fmt.Errorf(string(body))
		return key, server, ts, err
	}
	m := getLongPollServerStruct{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("Ошибка парсинга ответа на запрос getLongPollServer: %v/ Боди: %v", err, string(body))
		return key, server, ts, err
	}
	return m.Response.Key, m.Response.Server, m.Response.Ts, err
}

//Вызаваем сами долгие запросы на получение обновления
func getUpdates(key, server string, ts int) (tsNew int, text string, err error) {
	url := "https://" + server + "?act=a_check&key=" + key + "&ts=" + strconv.FormatInt(int64(ts), 10) + "&wait=25&version=3"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return tsNew, text, err
	}
	res, err := client.Do(req)
	if err != nil {
		return tsNew, text, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return tsNew, text, err
	}
	if strings.Contains(string(body), "failed") == true {
		m := failedStruct{}
		err = json.Unmarshal(body, &m)
		if err != nil {
			err = fmt.Errorf("Ошибка парсинга ответа на запрос getUpdates: %v Боди: %v", err, string(body))
			return tsNew, text, err
		}
		if m.Failed == 2 { //Проверка на ночную ошибку
			err = fmt.Errorf("Ночная ошибка: %v", string(body))
			return tsNew, text, err
		}
	}
	m := getUpdateStruct{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("Ошибка парсинга ответа на запрос getUpdates: %v Боди: %v", err, string(body))
		return tsNew, text, err
	}
	tsNew = m.Ts
	return tsNew, string(body), err
}

//getHistory получает историю сообщений по id пользователя. Возвращает текст сообщения и ошибку. Не воспользовался долгим запросом так как там не понять от кого пришло сообщение
func getHistory(userID, myID string, token, versinonAPIVk string, msgID float64) (textMsg string, err error) {
	url := "https://api.vk.com/method/messages.getHistory?user_id=" + userID + "&access_token=" + token + "&v=" + versinonAPIVk + "&count=10"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return textMsg, err
	}
	res, err := client.Do(req)
	if err != nil {
		return textMsg, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return textMsg, err
	}
	if strings.Contains(string(body), "error") == true {
		err = fmt.Errorf("Ошибка при получении getHistory: %v", string(body))
		fmt.Println(url)
		return textMsg, err
	}
	m := getHistoryStruct{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("Ошибка парсинга ответа на запрос getHistory: %v Боди: %v", err, string(body))
		return textMsg, err
	}
	// fromID := m.Response.Items
	for _, msg := range m.Response.Items { //Идем по полученным сообщениям и ищем наш id
		fromID := strconv.FormatInt(int64(msg.FromID), 10)
		if float64(msg.ID) == msgID { //Нашли наше сообщение в истории
			if fromID != myID { //Проверка что это сообщение не от нас самих же
				textMsg = msg.Text
			} else { //Если сообщение от нас то выходим без ошибки
				return textMsg, err
			}
		}
	}
	if textMsg == "" {
		err = fmt.Errorf("Не нашли id: %v Боди: %v", msgID, string(body))
	}
	return textMsg, err
}

type getUpdateStruct struct {
	Ts      int             `json:"ts"`
	Updates [][]interface{} `json:"updates"`
}
type failedStruct struct {
	Failed int `json:"failed"`
}

type getLongPollServerStruct struct {
	Response struct {
		Key    string `json:"key"`
		Server string `json:"server"`
		Ts     int    `json:"ts"`
	} `json:"response"`
}

type getHistoryStruct struct {
	Response struct {
		Count int `json:"count"`
		Items []struct {
			Date                  int           `json:"date"`
			FromID                int           `json:"from_id"`
			ID                    int           `json:"id"`
			Out                   int           `json:"out"`
			PeerID                int           `json:"peer_id"`
			Text                  string        `json:"text"`
			ConversationMessageID int           `json:"conversation_message_id"`
			FwdMessages           []interface{} `json:"fwd_messages"`
			Important             bool          `json:"important"`
			RandomID              int           `json:"random_id"`
			Attachments           []interface{} `json:"attachments"`
			IsHidden              bool          `json:"is_hidden"`
		} `json:"items"`
	} `json:"response"`
}
