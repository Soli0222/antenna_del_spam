package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	host := os.Getenv("HOST")
	token := os.Getenv("TOKEN")
	antennaId := os.Getenv("ANTENNA_ID")

	antenna_url := "https://" + host + "/api/antennas/notes"

	requestData := map[string]interface{}{
		"i":         token,
		"antennaId": antennaId,
		"limit":     100,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		fmt.Println("JSON marshaling failed:", err)
		return
	}

	// リクエストを作成
	req, err := http.NewRequest("POST", antenna_url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// HTTPリクエストを送信
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// レスポンスを読み取り
	fmt.Println("Response Status:", resp.Status)
	var responseBody []interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		fmt.Println("Error decoding response body:", err)
		return
	}

	// JSONデータを処理
	for i, element := range responseBody {
		fmt.Printf("Element %d:\n", i+1)
		processElement(host, token, element)
	}
}

func processElement(host string, token string, element interface{}) {
	switch elem := element.(type) {
	case map[string]interface{}:
		// idを取得
		id, ok := elem["id"].(string)
		if !ok {
			fmt.Println("ID is missing or not a string")
			return
		}

		// user情報を取得
		user, ok := elem["user"].(map[string]interface{})
		if !ok {
			fmt.Println("User information is missing or not a map")
			return
		}

		userId, _ := user["id"].(string)
		userHost, _ := user["host"].(string)

		// idとuser情報を表示
		fmt.Println("ID:", id)
		fmt.Println("User ID:", userId)
		fmt.Println("User Host:", userHost)

		// mentionsキーの要素が2つ以上ある場合の処理
		mentions, ok := elem["mentions"].([]interface{})
		if !ok || len(mentions) < 2 {
			fmt.Println("Mentions is missing or does not have at least 2 elements")
			return
		}

		requestData_note := map[string]interface{}{
			"i":      token,
			"noteId": id,
		}
		delnote_url := "https://" + host + "/api/notes/delete"
		postRequest(requestData_note, delnote_url)

		requestData_user := map[string]interface{}{
			"i":      token,
			"userId": userId,
		}
		deluser_url := "https://" + host + "/api/admin/delete-account"
		postRequest(requestData_user, deluser_url)

		// APIから設定情報を取得
		settings, err := getSettings(host, token)
		if err != nil {
			fmt.Println("Error getting settings:", err)
			return
		}

		// blockedHosts フィールドが存在するかどうかをチェック
		blockedHostsInterface, ok := settings["blockedHosts"]
		if !ok {
			fmt.Println("blocked hosts field not found")
			return
		}

		// blockedHosts フィールドの値を []string として取得
		blockedHosts, err := getStringSlice(blockedHostsInterface)
		if err != nil {
			fmt.Println("error getting blocked hosts:", err)
			return
		}

		// userHostがblockedHostsに含まれているか確認
		if !contains(blockedHosts, userHost) {
			// userHostが含まれていない場合、blockedHostsに追加
			blockedHosts = append(blockedHosts, userHost)

			// 更新したblockedHostsを設定情報に反映
			settings["blockedHosts"] = blockedHosts

			// 更新した設定情報のみをAPIに送信して更新
			err := updateBlockedHosts(host, token, blockedHosts)
			if err != nil {
				fmt.Println("error updating blocked hosts:", err)
				return
			}
			fmt.Println("blocked hosts updated successfully")
		} else {
			fmt.Println("user host already in blocked hosts")
		}

	default:
		fmt.Println("Unexpected type:", element)
	}
}

func postRequest(requestData interface{}, url string) {
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		fmt.Println("JSON marshaling failed:", err)
		return
	}

	// リクエストを作成
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// HTTPリクエストを送信
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// レスポンスを読み取り
	fmt.Println("Response Status:", resp.Status)
}

func getSettings(host string, token string) (map[string]interface{}, error) {
	url := "https://" + host + "/api/admin/meta"

	requestData := map[string]interface{}{
		"i": token,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("json marshaling failed: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		return nil, fmt.Errorf("error decoding response body: %v", err)
	}

	return responseBody, nil
}

func updateBlockedHosts(host string, token string, blockedHosts []string) error {
	url := "https://" + host + "/api/admin/update-meta"

	requestData := map[string]interface{}{
		"i":            token,
		"blockedHosts": blockedHosts,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("json marshaling failed: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	return nil
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func getStringSlice(data interface{}) ([]string, error) {
	switch v := data.(type) {
	case []interface{}:
		var result []string
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			} else {
				return nil, fmt.Errorf("value is not a string: %v", item)
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("value is not a slice: %v", data)
	}
}
