package utils

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	// "time"
	_ "github.com/lib/pq"
)

// CREATE TABLE mailbox (
// 	mailbox_id    varchar(20) NOT NULL,
// 	username   varchar(255) NOT NULL,
// 	password  varchar(255) NOT NULL,
// 	navi_date   TIMESTAMPTZ NOT NULL
//   );
// CREATE TABLE users (
// 	user_id    varchar(20) NOT NULL,
// 	username   varchar(255) NOT NULL,
// 	password  varchar(255) NOT NULL,
// 	navi_date   TIMESTAMPTZ NOT NULL,
// 	navi_user   varchar(255) NOT NULL,
// mailbox_id    varchar(20) NOT NULL,
// sex varchar(20) NOT NULL,
// active varchar(20) NOT NULL,
// first_name varchar(255) NOT NULL DEFAULT '',
// last_name varchar(255) NOT NULL DEFAULT '',
// clone_id varchar(20) NOT NULL DEFAULT ''--id пользователя, с которого тянем фотки
//   );
//   CREATE TABLE groups_vk (
// 	group_id    varchar(20) NOT NULL,
// 	token  varchar(255) NOT NULL,
// 	navi_date   TIMESTAMPTZ NOT NULL,
// 	navi_user   varchar(255) NOT NULL,
// active varchar(20) NOT NULL,
// group_name varchar(255) NOT NULL DEFAULT ''
// admin_group varchar(255) NOT NULL DEFAULT ''
//   );
// CREATE TABLE history ( -- историчная таблица для выдерживания лимитов
// 	hist_id    varchar(255) NOT NULL, -- ключь, id истории, чтобы не было повторов
// 	hist_type varchar(255) NOT NULL, -- тип действия (репост, лайк и тд)
// 	user_id varchar(20) NOT NULL, -- под каким пользователем
// 	navi_date   TIMESTAMPTZ NOT NULL, -- время
// group_id varchar(20) NOT NULL DEFAULT '', -- ид группы, если это репост или вступление в группу
// clone_id varchar(20) NOT NULL DEFAULT '' --id пользователя, с которого тянем фотки
//   );
// ALTER TABLE mailbox ADD PRIMARY KEY (mailbox_id); --создаем ключь
// ALTER TABLE users ADD PRIMARY KEY (user_id); --создаем ключь
// ALTER TABLE history ADD PRIMARY KEY (hist_id); --создаем ключь
// ALTER TABLE groups_vk ADD PRIMARY KEY (group_id); --создаем ключь
var db *sql.DB

func init() {
	userDB := "postgres"
	passwordDB := "XXXX"
	hostDB := "localhost"
	nameDB := "botdb"
	credsDB := []string{userDB, passwordDB, hostDB, nameDB}
	var err error
	db, err = sql.Open("postgres", "postgres://"+credsDB[0]+":"+credsDB[1]+"@"+credsDB[2]+"/"+credsDB[3])
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}

//InsertToDB Вставка в базу
func InsertToDB(qweryInsertDB string) (err error) {
	result, err := db.Exec(qweryInsertDB)
	if err != nil {
		err = fmt.Errorf("Ошибка записи в БД: %v Запрос: %v", err, qweryInsertDB)
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		err = fmt.Errorf("Ошибка записи в БД: %v Запрос: %v", err, qweryInsertDB)
		return err
	}

	return err
}

// GetUsersFromDB возвращает мапу с пользователями из бд
func GetUsersFromDB() (map[string][]string, error) {
	usersMap := make(map[string][]string)
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		err = fmt.Errorf("Ошибка подключения к БД: %v", err)
		return usersMap, err
	}
	defer rows.Close()

	users := make([]*usersStruct, 0)
	for rows.Next() {
		user := new(usersStruct)
		err := rows.Scan(&user.userID, &user.username, &user.password, &user.naviDate, &user.naviUser, &user.mailboxID, &user.sex, &user.active, &user.firstName, &user.lastName, &user.cloneID)
		if err != nil {
			err = fmt.Errorf("Ошибка поиска пользоватлей: %v", err)
			return usersMap, err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		err = fmt.Errorf("Ошибка поиска пользоватлей: %v", err)
		return usersMap, err
	}

	for _, user := range users {
		usersMap[user.userID] = []string{user.username, user.password, user.naviDate.String(), user.naviUser, user.mailboxID, user.sex, user.active, user.firstName, user.lastName, user.cloneID}
	}
	return usersMap, err
}

// GetGroupsFromDB возвращает мапу с группами из бд
func GetGroupsFromDB() (map[string][]string, error) {
	groupsMap := make(map[string][]string)
	rows, err := db.Query("SELECT * FROM groups_vk")
	if err != nil {
		err = fmt.Errorf("Ошибка подключения к БД: %v", err)
		return groupsMap, err
	}
	defer rows.Close()

	groups := make([]*groupsStruct, 0)
	for rows.Next() {
		group := new(groupsStruct)
		err := rows.Scan(&group.groupID, &group.token, &group.naviDate, &group.naviUser, &group.active, &group.groupName, &group.adminGroup)
		if err != nil {
			err = fmt.Errorf("Ошибка поиска пользоватлей: %v", err)
			return groupsMap, err
		}
		groups = append(groups, group)
	}
	if err = rows.Err(); err != nil {
		err = fmt.Errorf("Ошибка поиска пользоватлей: %v", err)
		return groupsMap, err
	}

	for _, group := range groups {
		groupsMap[group.groupID] = []string{group.token, group.naviDate.String(), group.naviUser, group.active, group.groupName, group.adminGroup}
	}
	// log.Println(groupsMap)
	return groupsMap, err
}

// GetLikedUsers возвращает уникальный слайс пользователей, которым поставили лайк
func GetLikedUsers(myID string) (userSlice []string, err error) {
	type getLikeUsers struct {
		user string
	}
	rows, err := db.Query(fmt.Sprintf("select hist_id from history t where t.user_id = '%v' and t.hist_type='Лайк'", myID))
	if err != nil {
		err = fmt.Errorf("Ошибка подключения к БД: %v", err)
		return userSlice, err
	}
	defer rows.Close()

	users := make([]*getLikeUsers, 0)
	for rows.Next() {
		user := new(getLikeUsers)
		err := rows.Scan(&user.user)
		if err != nil {
			err = fmt.Errorf("Ошибка поиска пользоватлей: %v", err)
			return userSlice, err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		err = fmt.Errorf("Ошибка поиска пользоватлей: %v", err)
		return userSlice, err
	}

	for _, user := range users {
		//Парсим hist_id. %v-%v_%v", typeObg, ownerID, itemID
		usertmp1 := strings.Split(string(user.user), "-") //Разбиваем пробелами для получения ownerID_itemID
		usertm2 := strings.Split(usertmp1[1], "_")        //Разбиваем пробелами для получения ownerID
		userSlice = append(userSlice, usertm2[0])
	}
	return Unique(userSlice), err
}

//Antiban спит от 1 до 10 минут между запросами в разрезе пользоватлея, если прошлы запрос был менее минуты назад
func Antiban(userID, function string, sleepSeconds int) (err error) {
	//Проверка, что текущее время в интервале от 9 до 23 то выходим
	var chasNachala = 9
	if time.Now().Hour() < chasNachala || time.Now().Hour() > 22 {
		log.Println(function + ": В это время пользователь должен спать, ждем")
		var day int                          //Число
		if time.Now().Hour() < chasNachala { //Если текущий час меньше 9 то число оставляем
			day = time.Now().Day()
		}
		if time.Now().Hour() > 22 { //Если текущий час 23 то для расчета берем следующий день
			day = time.Now().Day() + 1
		}
		vremyaNachala := time.Date(time.Now().Year(), time.Now().Month(), day, chasNachala, 0, 0, 0, time.Local)
		log.Printf(function+": Спим %v", time.Since(vremyaNachala))
		time.Sleep(time.Duration(time.Since(vremyaNachala).Seconds()) * (-1) * time.Second)
	}
	var lastChange time.Time
	rand.Seed(time.Now().Unix()) //Seed для рандома
	qwery := `select navi_date from history t where t.user_id = '` + userID + `' and t.hist_id not like 'infrend-%' ORDER BY navi_date DESC limit 1`
	row := db.QueryRow(qwery)
	err = row.Scan(&lastChange)
	if err != nil {
		err = fmt.Errorf("Ошибка антибана: %v Запрос %v userID=%v", err, qwery, userID)
		time.Sleep(30 * time.Second) //Спим 30 если проблемы с подключением к бд
		return err
	}
	proshloVremeni := time.Since(lastChange).Seconds() //Считаем сколько прошло времени с последнего запроса
	if proshloVremeni < 60 {                           //Если времени с последнего изменения прошло меньше минуты то ждем от 1 до 10 минут
		son := time.Duration(60+rand.Intn(sleepSeconds)) * time.Second // Сон, рандом от 1 мин до 5 минут
		log.Printf(function+": Последний запрос пользоватлея %v выполнен в %v %v секунд назад. Спим %v до следущего задания. Запуск в %v", userID, lastChange, proshloVremeni, son, time.Now().Add(son))
		time.Sleep(son)
	} else {
		log.Printf(function+": Последний запрос пользоватлея %v выполнен в %v, больше минуты назад, %v секунд", userID, lastChange, proshloVremeni)
	}
	return err
}

//GetOneRowDB возвращает одно значение из таблицы
func GetOneRowDB(qwery string) (result string, err error) {
	row := db.QueryRow(qwery)
	err = row.Scan(&result)
	if err != nil {
		err = fmt.Errorf("Ошибка выполнения единичного запроса в БД %v: %v", qwery, err)
		return result, err
	}
	return result, err
}

//Структура таблицы users
type usersStruct struct {
	userID    string
	username  string
	password  string
	naviDate  time.Time
	naviUser  string
	mailboxID string
	sex       string
	active    string
	firstName string
	lastName  string
	cloneID   string
}

//Структура таблицы groups_vk
type groupsStruct struct {
	groupID    string
	token      string
	naviDate   time.Time
	naviUser   string
	active     string
	groupName  string
	adminGroup string
}
