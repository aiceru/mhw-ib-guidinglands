package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/api/sheets/v4"

	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"
	"github.com/unrolled/render"
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

func getData() error {
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
	spreadsheetId := "1YTO5fcXNmfpU7XptQRuxmZIkvWq5bly-L8j3qq22oy8"
	readRange := "인도하는 땅 몬스터 소재!B12:G53"
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

	readRange = "출현 몬스터 Ver.가로!B43:I62"
	wildspire, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
		return err
	}
	makeHabitatList(Wildspire, wildspire)

	readRange = "출현 몬스터 Ver.가로!B67:I83"
	coral, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
		return err
	}
	makeHabitatList(Coral, coral)

	readRange = "출현 몬스터 Ver.가로!B88:I100"
	rotten, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
		return err
	}
	makeHabitatList(Rotten, rotten)

	return nil
}

func makeMonsterList(v *sheets.ValueRange) {
	if len(v.Values) == 0 {
		log.Println("No data found.")
		return
	}
	for _, row := range v.Values {
		name := row[0].(string)
		if row[2].(string) != "X" {
			m := &MonsterInfo{
				Name:       name,
				Difficulty: Normal,
				Item:       row[2].(string),
				Habitat:    [4][7]int{},
			}
			monsters[m.Difficulty.String()+name] = m
		}
		if row[4].(string) != "X" {
			m := &MonsterInfo{
				Name:       name,
				Difficulty: Tempered,
				Item:       row[4].(string),
				Habitat:    [4][7]int{},
			}
			monsters[m.Difficulty.String()+name] = m
		}
	}
}

func makeHabitatList(field int, v *sheets.ValueRange) {
	var diff difficulty

	for _, row := range v.Values {
		for i, star := range row[1:] {
			starString := star.(string)
			nFreq := strings.Count(starString, "★")
			if nFreq > 0 {
				diff = Normal
				monsterInfo := monsters[diff.String()+row[0].(string)]
				monsterInfo.Habitat[field][i] = nFreq
			} else {
				tFreq := strings.Count(starString, "☆")
				if tFreq > 0 {
					diff = Tempered
					monsterInfo := monsters[diff.String()+row[0].(string)]
					monsterInfo.Habitat[field][i] = tFreq
				}
			}
		}
	}
}

const (
	Forest		= iota
	Wildspire
	Coral
	Rotten
)

type difficulty int
func (t difficulty) String() string {
	switch t {
	case Normal:
		return ""
	case Tempered:
		return "역전 "
	}
	return ""
}

const (
	Normal		= iota
	Tempered
)

type MonsterInfo struct {
	Name		string
	Difficulty	difficulty
	Item 		string
	Habitat		[4][7]int
}

func (m MonsterInfo) String() string {
	return m.Difficulty.String() + m.Name
}

var renderer *render.Render
var monsters map[string]*MonsterInfo

func init() {
	renderer = render.New(render.Options{
		Layout: "layout_main",
		Funcs: []template.FuncMap{
			{
				"Iterate": func(start, count int) []int {
					var Items []int
					for i := start; i < start+count; i++ {
						Items = append(Items, i)
					}
					return Items
				},
			},
			{
				"Newline": func(i int) bool {
					if i % 4 == 3 {
						return true
					}
					return false
				},
			},
		},
	})
	monsters = make(map[string]*MonsterInfo)
}

func mainpage(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	renderer.HTML(w, http.StatusOK, "monster_lists", map[string]interface{}{
		"flv": 1,
		"wlv": 1,
		"clv": 1,
		"rlv": 1,
	})
}

func findleft(list [7]int, pos int) bool {
	for i := pos; i >= 0; i-- {
		if list[i] > 0 {
			return true
		}
	}
	return false
}

func findright(list [7]int, pos int) bool {
	for i := pos; i < 7; i++ {
		if list[i] > 0 {
			return true
		}
	}
	return false
}

func contains(s []string, t string) bool {
	for _, i := range s {
		if i == t {
			return true
		}
	}
	return false
}

func whatYouCannotSeeHere(result *[4][]*MonsterInfo, lvs []int) {
	for _, monsterInfo := range monsters {
		for field, current_lv := range lvs {
			if findleft(monsterInfo.Habitat[field], current_lv-2) && !findright(monsterInfo.Habitat[field], current_lv-1) {
				result[field] = append(result[field], monsterInfo)
			}
		}
	}
}

func whatYouCannotSee(result *[]*MonsterInfo, mlist *[4][]*MonsterInfo, lvs []int) {
	for field := Forest; field <= Rotten; field++ {
		for _, monsterInfo := range mlist[field] {
			appearsInOtherField := false
			for otherField := Forest; otherField <= Rotten; otherField++ {
				if otherField == field {
					continue
				}
				if findright(monsterInfo.Habitat[otherField], lvs[otherField]-1) {
					appearsInOtherField = true
					break
				}
			}
			if !appearsInOtherField {
				*result = append(*result, monsterInfo)
			}
		}
	}
}

func calculateLists(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	req.ParseForm()
	fLv, _ :=	strconv.Atoi(req.FormValue("forest_lev")[0:])
	wLv, _ :=	strconv.Atoi(req.FormValue("wild_lev")[0:])
	cLv, _ :=	strconv.Atoi(req.FormValue("coral_lev")[0:])
	rLv, _ :=	strconv.Atoi(req.FormValue("rotten_lev")[0:])

	currentLvs := []int{fLv, wLv, cLv, rLv}
	var cannotSeeLists [4][]*MonsterInfo
	var totalLists []*MonsterInfo
	var cannotSeeIfFields [4][]*MonsterInfo
	var cannotSeeIfs [4][]*MonsterInfo

	whatYouCannotSeeHere(&cannotSeeLists, currentLvs)
	whatYouCannotSee(&totalLists, &cannotSeeLists, currentLvs)

	futureLvs := make([]int, 4)
	for field := Forest; field <= Rotten; field++ {
		if currentLvs[field] >= 7 {
			continue
		}
		copy(futureLvs, currentLvs)
		futureLvs[field]++

		for i := 0; i < 4; i++ {
			cannotSeeIfFields[i] = cannotSeeIfFields[i][:0]
		}
		whatYouCannotSeeHere(&cannotSeeIfFields, futureLvs)
		whatYouCannotSee(&cannotSeeIfs[field], &cannotSeeIfFields, futureLvs)
	}


	renderer.HTML(w, http.StatusOK, "monster_lists",
		map[string]interface{}{
			"flv":                 fLv,
			"wlv":                 wLv,
			"clv":                 cLv,
			"rlv":                 rLv,
			"NotAppearList":       totalLists,
			"ForestNotAppearList": cannotSeeLists[Forest],
			"WildNotAppearList":   cannotSeeLists[Wildspire],
			"CoralNotAppearList":  cannotSeeLists[Coral],
			"RottenNotAppearList": cannotSeeLists[Rotten],
			"NotIfForestUp":       cannotSeeIfs[Forest],
			"NotIfWildUp":         cannotSeeIfs[Wildspire],
			"NotIfCoralUp":        cannotSeeIfs[Coral],
			"NotIfRottenUp":       cannotSeeIfs[Rotten],
		})
}

func main() {
	getData()

/*	for k, m := range monsters {
		fmt.Println(k, m.Info[Normal], m.Info[Tempered])
	}*/

	router := httprouter.New()

	router.GET("/", mainpage)
	router.POST("/", calculateLists)



	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":8080")
}