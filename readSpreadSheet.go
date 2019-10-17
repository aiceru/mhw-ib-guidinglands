package main

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "data/token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getDataFromGoogleSheet() error {
	b, err := ioutil.ReadFile("data/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.NewService(context.Background(),
		option.WithScopes(sheets.SpreadsheetsReadonlyScope),
		option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// Prints the names and majors of students in a sample spreadsheet:
	// https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
	spreadsheetId := "1w80MacGIadcJ4b5LtgLPdBHH-gTymScTVXz7-e4qEJQ"
	readRange := "인도하는 땅 몬스터 소재!B12:G60"
	mlist, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
		return err
	}

	makeMonsterList(mlist)

	readRange = "출현 몬스터 Ver.가로!B18:I38"
	forest, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
		return err
	}
	makeHabitatList(Forest, forest)

	readRange = "출현 몬스터 Ver.가로!B43:I63"
	wildspire, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
		return err
	}
	makeHabitatList(Wildspire, wildspire)

	readRange = "출현 몬스터 Ver.가로!B68:I85"
	coral, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
		return err
	}
	makeHabitatList(Coral, coral)

	readRange = "출현 몬스터 Ver.가로!B90:I103"
	rotten, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
		return err
	}
	makeHabitatList(Rotten, rotten)

	readRange = "출현 몬스터 Ver.가로!B108:I127"
	lava, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
		return err
	}
	makeHabitatList(Lava, lava)

	return nil
}

func makeMonsterList(v *sheets.ValueRange) {
	if len(v.Values) == 0 {
		log.Println("No data found.")
		return
	}
	for i, row := range v.Values {
		if len(row) > 0 {
			name := row[0].(string)
			if row[2].(string) != "X" {
				m := &MonsterInfo{
					Code:       i * 2,
					Name:       name,
					Ename: row[1].(string),
					Difficulty: Normal,
					Item:       row[2].(string),
					Habitat:    [FieldMax][7]int{},
				}
				monsters[m.Difficulty.String()+name] = m
			}
			if row[4].(string) != "X" {
				m := &MonsterInfo{
					Code:       (i * 2) + 1,
					Name:       name,
					Ename: row[1].(string),
					Difficulty: Tempered,
					Item:       row[4].(string),
					Habitat:    [FieldMax][7]int{},
				}
				if name == "얀가루루가" {
					m.Difficulty = Wounded
				}
				monsters[m.Difficulty.String()+name] = m
			}
		}
	}
}

func makeHabitatList(field int, v *sheets.ValueRange) {
	var diff difficulty

	for _, row := range v.Values {
		for i, star := range row[1:] {
			name := row[0].(string)
			starString := star.(string)
			nFreq := strings.Count(starString, "★")
			if nFreq > 0 {
				diff = Normal
				monsterInfo := monsters[diff.String()+name]
				monsterInfo.Habitat[field][i] = nFreq
			} else {
				tFreq := strings.Count(starString, "☆")
				if tFreq > 0 {
					diff = Tempered
					if name == "얀가루루가" {
						diff = Wounded
					}
					monsterInfo := monsters[diff.String()+name]
					monsterInfo.Habitat[field][i] = tFreq
				}
			}
		}
	}
}
