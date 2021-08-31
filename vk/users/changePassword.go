package users

import (
	"strings"
  "fmt"
  "net/http"
  "io/ioutil"
)
//ChangePasswordVK Смена пароля пользователя
func ChangePasswordVK(versinonAPIVk, token, oldPWD, newPWD string) (err error) {
  url := "https://api.vk.com/method/account.changePassword?v="+versinonAPIVk+"&access_token="+token+"&old_password="+oldPWD+"&new_password="+newPWD
  method := "GET"
  client := &http.Client {}
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
	err = fmt.Errorf("Ошибка сброса пароля " + res.Status + " " + string(body))
	return err
}
if strings.Contains(string(body), "failed") == true || strings.Contains(string(body), "error") {
	err = fmt.Errorf("Ошибка сброса пароля: %v", string(body))
	return err
}
  return err
}