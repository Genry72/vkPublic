package users

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bot.my/utils"

	"github.com/PuerkitoBio/goquery"
	"github.com/sclevine/agouti"
)

//FreeLikes для работы с сайтом freelikes.online. username, password от ВК
func FreeLikes(versinonAPIVk, token, myID, usernameVK, passwordVK string, vkErrUserChan chan error) (err error) {
	//Один раз получаем PHPSESSID
	//Раз в 30 минут получаем токен и апдейтим PHPSESSID
	rand.Seed(time.Now().Unix()) ////Seed для рандома
	randomChislo := 100 + rand.Intn(500)
	idStr := fmt.Sprintf("%v%v", time.Now().Unix(), randomChislo) //Уникальный id для выполнения проверки
	sessID, err := GetphpSessID(usernameVK, usernameVK)
	if err != nil {
		err = fmt.Errorf("Кука ВК не получена, лайков не будет:%v", err)
		return err
	}
	err = getUloginToken(usernameVK, passwordVK, sessID, myID, vkErrUserChan)
	if err != nil {
		vkErrUserChan <- err
		return err
	}

	go func() { //Беcконечно получает токен и апдейтит sessID раз в 30 минут
		for {
			err = getUloginToken(usernameVK, passwordVK, sessID, myID, vkErrUserChan)
			if err != nil {
				vkErrUserChan <- err
				time.Sleep(50 * time.Minute)
			}
			time.Sleep(30 * time.Minute)
		}
	}()
	go func() { //Лайкаем
		for {
			if sessID == "" {
				continue
			}
			mapa, err := findLikes(sessID, "vklike", myID, vkErrUserChan)
			if err != nil {
				vkErrUserChan <- err
				time.Sleep(5 * time.Minute)
				continue
			}
			if len(mapa) == 0 {
				err = fmt.Errorf("FreeLikes, отвалиась авторизация, кука = %v", sessID)
				time.Sleep(5 * time.Minute)
				continue
			}
			var typeObg string
			var ownerID string
			var itemID string
			for i, zadanie := range mapa {
				if strings.Contains(zadanie[2], "photo") == true {
					typeObg = "photo"
					s := strings.Split(zadanie[2], typeObg)
					b := strings.Split(s[1], "_")
					ownerID = b[0]
					itemID = b[1]
				}
				if strings.Contains(zadanie[2], "wall") == true {
					typeObg = "post"
					s := strings.Split(zadanie[2], "wall")
					// fmt.Println(s)
					b := strings.Split(s[1], "_")
					ownerID = b[0]
					itemID = b[1]
					// continue
				}
				if strings.Contains(zadanie[2], "video") == true {
					typeObg = "video"
					s := strings.Split(zadanie[2], typeObg)
					// fmt.Println(s)
					b := strings.Split(s[1], "_")
					ownerID = b[0]
					itemID = b[1]
				}
				if typeObg == "" {
					continue
				}
				//Лайкаем
				err = LikesAdd(versinonAPIVk, token, typeObg, ownerID, itemID, myID, vkErrUserChan)
				if err != nil {
					vkErrUserChan <- err
					continue
				}
				//Проверка задания
				id, err := strconv.ParseInt(idStr, 10, 64)
				err = check(zadanie[0], zadanie[1], fmt.Sprintf("%v", id+int64(i)), sessID, "vklike")
				if err != nil {
					vkErrUserChan <- err
					continue
				}
			}

		}
	}()

	go func() { //Добавляем в друзья
		for {
			if sessID == "" {
				continue
			}
			mapa, err := findLikes(sessID, "vkfriend", myID, vkErrUserChan)
			if err != nil {
				vkErrUserChan <- err
				time.Sleep(5 * time.Minute)
				continue
			}
			if len(mapa) == 0 {
				err = fmt.Errorf("FreeLikes, отвалиась авторизация, кука = %v", sessID)
				time.Sleep(5 * time.Minute)
				continue
			}
			for i, zadanie := range mapa {
				s := strings.Split(zadanie[2], "id") //Пилим по id, достаем id пользоватлея   https://vk.com/id529643144
				if len(s) == 0 {
					err = fmt.Errorf("Хреновый пользователь %v", zadanie)
					vkErrUserChan <- err
					continue
				}
				user := s[1]
				//Добавляем в друзья
				err = addFrendsAPI(versinonAPIVk, token, user, myID, vkErrUserChan, false)
				if err != nil {
					vkErrUserChan <- err
					continue
				}
				//Проверка задания
				id, err := strconv.ParseInt(idStr, 10, 64)
				err = check(zadanie[0], zadanie[1], fmt.Sprintf("%v", id+int64(i)), sessID, "vkfriend")
				if err != nil {
					vkErrUserChan <- err
					continue
				}
			}

		}
	}()
	return err
}

//GetphpSessID достает из кук sessID c сайта http://freelikes.online/
func GetphpSessID(username, password string) (sessID string, err error) {
	url := "http://freelikes.online"
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return sessID, err
	}
	res, err := client.Do(req)
	if err != nil {
		return sessID, err
	}
	defer res.Body.Close()
	if err != nil {
		return sessID, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return sessID, err
	}
	for _, kuka := range res.Cookies() {
		if kuka.Name == "PHPSESSID" {
			sessID = kuka.Value
		}
	}
	if sessID == "" {
		err = fmt.Errorf("Пустаяя кука:%v %v", res.Status, string(body))
	}
	return sessID, err
}

//GetTokenFreelikes Получает токен и апдейтит sessID
func getUloginToken(login, pass, sessID, myID string, vkErrUserChan chan error) (err error) {
	log.Printf("Выполняем получение токена Utoken для пользователя %v", myID)
	err = utils.Antiban(myID, "getUloginToken", 300)
	if err != nil {
		vkErrUserChan <- err
	}
	var token string
	// driver := agouti.PhantomJS()
	// driver := agouti.ChromeDriver()
	driver := agouti.ChromeDriver(
		// agouti.ChromeOptions("args", []string{"--headless", "--disable-gpu", "--no-sandbox"}),
		agouti.ChromeOptions("args", []string{"--headless", "--disable-gpu", "--no-sandbox"}),
	)

	err = driver.Start()
	if err != nil {
		return err
	}

	page, err := driver.NewPage()
	if err != nil {
		return err
	}
	// page.Size(1920, 1080)
	page.SetImplicitWait(15000000) //Таймаут ожиданя вейта

	if err := page.Navigate("https://oauth.vk.com/authorize?v=5.62&client_id=3280318&scope=friends,schools,email&display=page&response_type=code&redirect_uri=https://ulogin.ru/auth.php?name=vkontakte"); err != nil {
		return err
	}
	_, err = page.FindByID(`install_allow`).Visible()
	if err != nil {
		err = fmt.Errorf("Не дождались загрузки кнопки войти:%v", err)
		return err
	}
	//Вводим логин
	err = page.FindByXPath(`/html/body/div/div/div/div[2]/form/div/div/input[6]`).SendKeys(login)
	if err != nil {
		return err
	}
	//Вводим пароль
	err = page.FindByXPath(`/html/body/div/div/div/div[2]/form/div/div/input[7]`).SendKeys(pass)
	if err != nil {
		return err
	}
	//Нажимаем вход
	err = page.FindByID(`install_allow`).Submit()
	if err != nil {
		return err
	}
	//Ждем прогрузки страницы. Когда появится кнопка закрыть окно
	_, err = page.FindByXPath(`/html/body/div/button`).Visible()
	if err != nil {
		err = fmt.Errorf("Не дождались загрузки кнопки войти:%v", err)
		return err
	}

	//Забираем боди, там токен
	z, err := page.HTML()
	if err != nil {
		return err
	}

	//Парсим боди
	scanner := bufio.NewScanner(strings.NewReader(z)) //Построчно читаем боди
	for scanner.Scan() {
		var out string
		out = scanner.Text()
		var clearstring string
		ansiCode := []string{
			`</script>`,
			`;`,
			"'",
			"=",
			` `,
			// `"`,
			// `>`,
		}
		//Идем по циклу с цветами и удаляем их
		for _, color := range ansiCode {
			// fmt.Println(color)
			clearstring = strings.Replace(out, color, "", -1)
			// fmt.Println("Нашли")
			out = clearstring

		}
		if strings.Contains(out, "token") == true {
			s := strings.Split(out, "token")
			token = s[1]
		}
	}
	err = driver.Stop()
	if err != nil {
		return err
	}
	//Продлеваем имеющуюся куку
	err = updatePhpSessID(token, sessID)
	if err != nil {
		return err
	}
	log.Printf("Gолучение токена Utoken для пользователя %v выполнено", myID)
	return err
}

func updatePhpSessID(uloginToken, sessID string) (err error) {
	url := "http://freelikes.online/login.php"
	method := "POST"
	payload := strings.NewReader("token=" + uloginToken)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return err
	}
	req.Header.Add("Host", " freelikes.online")
	req.Header.Add("User-Agent", " Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:82.0) Gecko/20100101 Firefox/82.0")
	req.Header.Add("Accept", " text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Add("Accept-Language", " ru-RU,ru;q=0.8,en-US;q=0.5,en;q=0.3")
	req.Header.Add("Accept-Encoding", " gzip, deflate")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", " 38")
	req.Header.Add("Origin", " http://freelikes.online")
	req.Header.Add("Connection", " keep-alive")
	req.Header.Add("Referer", " http://freelikes.online/")
	req.Header.Add("Cookie", " PHPSESSID="+sessID)
	req.Header.Add("Upgrade-Insecure-Requests", " 1")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if strings.Contains(string(body), "Undefined index") == true {
		err = fmt.Errorf("Ошибка обновления PHPSESSID")
		return err
	}
	return err
}

//Проверка выполнения задания
func check(divid, taskid, id, sessID, typeGet string) (err error) {
	url := "https://freelikes.online/ajax.php?divid=" + divid + "&taskid=" + taskid + "&task=" + typeGet + "&_=" + id
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Referer", "https://freelikes.online/earn/vkontakte/"+typeGet)
	req.Header.Add("Accept-Language", "ru-RU,ru;q=0.9")
	req.Header.Add("Cookie", "PHPSESSID="+sessID)

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if strings.Contains(string(body), "Задание") == true {
		// err = fmt.Errorf("Ошибка проверки задания")
		return err
	}
	return err
}

//Собираем задания. Для лайков и друзей. Добавляем задание если баланс выше 100
func findLikes(sessID, typeGet, myID string, vkErrUserChan chan error) ([][]string, error) { //typeGet vklike либо vkfriend
	log.Printf("Ищем задания Freelikes %v", typeGet)
	minprice := 4           //Минимальная цена задания
	var bigslice [][]string // ,Большой слайс слайсов
	// resultMapa := make(map[string][][]string) //Мапа для возврата результата. Ключь unixtimestamp + рандом от 100 до 900. 0 и 1 элемент id. 2-й элемент урл фото
	url := "http://freelikes.online/earn/vkontakte/" + typeGet
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return bigslice, err
	}
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Cache-Control", "max-age=0")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	req.Header.Add("Sec-Fetch-User", "?1")
	req.Header.Add("Sec-Fetch-Dest", "document")
	req.Header.Add("Referer", "https://freelikes.online/")
	req.Header.Add("Accept-Language", "ru-RU,ru;q=0.9")
	req.Header.Add("Cookie", "PHPSESSID="+sessID)
	res, err := client.Do(req)
	if err != nil {
		return bigslice, err
	}
	defer res.Body.Close()
	//Парсим html
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return bigslice, err
	}

	document.Find(`article > div`).Each(func(index int, item *goquery.Selection) {
		title := item.Text() //Получаем цену
		linkTag := item.Find("#div2 > div > div.task-body > font > a")
		link, _ := linkTag.Attr("onclick") //Получаем урл с заданием
		scanner := bufio.NewScanner(strings.NewReader(title))
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "баллов") == true {
				ansiCode := []string{
					"баллов",
					"\t\u00a0",
					" ",
				}
				//Идем по циклу с цветами и удаляем их
				title = scanner.Text()
				for _, color := range ansiCode {
					title = strings.Replace(title, color, "", -1)
				}
			}
		}
		price, err := strconv.Atoi(title)
		if err != nil {
			err = fmt.Errorf("Ошибка парсинга цены задания %v: %v", title, err)
			vkErrUserChan <- err
			return
		}
		// fmt.Printf("Цена: %v\n", price)
		ansiCode := []string{
			`)`,
			`(`,
			`"`,
			`;`,
			"proverka",
		}
		for _, color := range ansiCode {
			link = strings.Replace(link, color, "", -1)
		}
		r := strings.Split(link, ",")
		if price > minprice-1 {
			divid := strings.Replace(r[0], " ", "", -1)
			taskid := strings.Replace(r[1], " ", "", -1) //
			likeID := strings.Replace(r[3], " ", "", -1) //Урл для лайка
			var smolSlice []string                       //Маленький слайс
			smolSlice = append(smolSlice, divid, taskid, likeID)
			bigslice = append(bigslice, smolSlice)
		}
	})
	//Парсим баланс
	var balance int
	document.Find(`#points2`).Each(func(index int, item *goquery.Selection) {
		balance, err = strconv.Atoi(item.Text())
		if err != nil {
			err = fmt.Errorf("Не удалось получить текущий баланс: %v", err)
			vkErrUserChan <- err
			return
		}
	})
	if balance > 100 {
		log.Printf("Баланс %v, создаем задание", balance)
		err = addTask(myID, sessID, balance)
		if err != nil {
			log.Println(err)
		}
	}
	log.Printf("Нашли %v заданий Freelikes %v", len(bigslice), typeGet)
	return bigslice, err
}

//Добавляем таску на добавление друзей
func addTask(myID, sessID string, balance int) (err error) {
	price := 6               //Цена за 1 добавление в друзья
	count := balance / price //Целое число от деления
	log.Printf("Добавляем таску на добавление друзей пользователю %v", myID)
	url := "http://freelikes.online/newvk"
	method := "POST"
	payload := strings.NewReader("typetask=vkgroup&name=groups" + fmt.Sprintf("%v", time.Now().Minute()) + "&link=https%3A%2F%2Fvk.com%2Ffuck_realnost&amount=" + strconv.FormatInt(int64(count), 10) + "&price=" + strconv.FormatInt(int64(price), 10) + "&addorder=") //Взял минуты для уникальности задания

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return err
	}
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Cache-Control", "max-age=0")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Origin", "http://freelikes.online")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.193 Safari/537.36")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Add("Referer", "http://freelikes.online/newvk")
	req.Header.Add("Accept-Language", "ru-RU,ru;q=0.9")
	req.Header.Add("Cookie", "PHPSESSID="+sessID)

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(strings.NewReader(string(body))) //Построчно читаем инфу о запросе
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "alert alert") == true {
			if strings.Contains(scanner.Text(), "success") == false {
				// Такое задание уже выполняется
				if strings.Contains(scanner.Text(), "Такое задание уже выполняется") == true { //Пропускаем если прошлое задание еще выполняется
					log.Println("Прошлое задание еще выполняется")
					continue
				}
				err = fmt.Errorf("Ошибка добавления таска %v боди запроса: %v", scanner.Text(), payload)
				return err
			}
		}
	}
	// fmt.Println(string(body))
	log.Printf("Добавлили таску на прокачку группы %v", myID)
	return err
}
