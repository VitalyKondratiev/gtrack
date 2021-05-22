package toggl

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"../config"
	"github.com/Jeffail/gabs"
	"github.com/manifoldco/promptui"
)

type Toggl struct {
	Credentials config.TogglConfig
	isLoggedIn  bool
}

func (toggl Toggl) IsLoggedIn() bool {
	return toggl.isLoggedIn
}

func (toggl Toggl) SetCredentials() Toggl {
	prompt := promptui.Prompt{
		Label: "Enter API key of your Toggl.com account",
	}
	result, err := prompt.Run()
	if err != nil {
		toggl.isLoggedIn = false
		return toggl
	}
	toggl.Credentials.ApiKey = result

	username, statusCode := toggl.GetUser()
	if statusCode == 403 {
		toggl.isLoggedIn = false
		fmt.Printf("API key %s is not correct\n", toggl.Credentials.ApiKey)
	}
	if statusCode == 200 {
		toggl.isLoggedIn = true
		fmt.Printf("You sucessfully logged in Toggl as %s\n", *username)
	}

	return toggl
}

func (toggl Toggl) GetTimeEntries() (bool, int) {
	client := http.Client{}
	req, err := http.NewRequest("GET", "https://api.track.toggl.com/api/v8/time_entries", nil)
	if err != nil {
		return false, 0
	}
	req.SetBasicAuth(toggl.Credentials.ApiKey, "api_token")
	resp, err := client.Do(req)
	if err != nil {
		return false, 0
	}
	defer resp.Body.Close()
	// if resp.StatusCode == 200 {
	// 	data, _ := ioutil.ReadAll(resp.Body)
	// 	jsonParsed, err := gabs.ParseJSON(data)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	since := jsonParsed.Path("since").Data().(float64)
	// 	fmt.Printf("%d\n", since)
	// }
	return true, resp.StatusCode
}

func (toggl Toggl) GetUser() (*string, int) {
	client := http.Client{}
	req, err := http.NewRequest("GET", "https://api.track.toggl.com/api/v8/me", nil)
	if err != nil {
		return nil, 0
	}
	req.SetBasicAuth(toggl.Credentials.ApiKey, "api_token")
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		data, _ := ioutil.ReadAll(resp.Body)
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			panic(err)
		}
		fullname := jsonParsed.Path("data.fullname").Data().(string)
		return &fullname, resp.StatusCode
	}
	return nil, resp.StatusCode
}
