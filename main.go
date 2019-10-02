/*package main

import (
	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"
	"github.com/unrolled/render"
	"net/http"
)

var renderer *render.Render

func init() {
	renderer = render.New()
}

func mainpage(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	renderer.HTML(w, http.StatusOK, "index", map[string]string{"title": "MHW IB Guiding Land Monster List"})
}

func main() {
	router := httprouter.New()

	router.GET("/", mainpage)

	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":8080")
}

*/

package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/api/sheets/v4"
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

func getData(monsters []Monster) error {
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

	makeMonsterList(monsters, mlist)

/*	readRange = "출현 몬스터 Ver.가로!B18:I38"
	forest, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
		return err
	}

	readRange = "출현 몬스터 Ver.가로!B43:I62"
	wildspire, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
		return err
	}

	readRange = "출현 몬스터 Ver.가로!B67:I83"
	coral, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
		return err
	}

	readRange = "출현 몬스터 Ver.가로!B88:I100"
	rotten, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
		return err
	}*/

	/*if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		fmt.Println("Name, Major:")
		for _, row := range resp.Values {
			// Print columns A and E, which correspond to indices 0 and 4.
			fmt.Printf("%s, %s, %s, %s\n", row[0], row[1], row[2], row[3])
		}
	}*/
	return nil
}

func makeMonsterList(monsters []Monster, v *sheets.ValueRange) {
	if len(v.Values) == 0 {
		log.Println("No data found.")
		return
	}
	for i, row := range v.Values {
		monsters[i].Name = row[0].(string)
	}
}

const (
	Normal		= iota
	Tempered
)

type Monster struct {
	Name 		string
	Info		map[string]MonsterInfo
}

type MonsterInfo struct {
	Type		int
	Item 		string
	Habitat		[4][7]int
}

func marshalSheet(valueRange *sheets.ValueRange) {

}

func main() {
	monsters := make([]Monster, 42)
	getData(monsters)
}