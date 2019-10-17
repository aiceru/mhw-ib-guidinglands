package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"
	"github.com/unrolled/render"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
)

const ( // Fields
	Forest = iota
	Wildspire
	Coral
	Rotten
	Lava
	FieldMax
)

type difficulty int

func (t difficulty) String() string {
	switch t {
	case Normal:
		return ""
	case Tempered:
		return "역전 "
	case Wounded:
		return "상처입은 "
	}
	return ""
}

const (
	Normal = iota
	Tempered
	Wounded
)

type MonsterInfo struct {
	Code       int
	Name       string
	Ename string
	Difficulty difficulty
	Item       string
	Habitat    [FieldMax][7]int
}

func (m MonsterInfo) Copy(l *MonsterInfo) {
	for i := Forest; i < FieldMax; i++ {
		for j := 0; j < 7; j++ {
			l.Habitat[i][j] = m.Habitat[i][j]
		}
	}
	l.Code = m.Code
	l.Name = m.Name
	l.Difficulty = m.Difficulty
	l.Item = m.Item
}

func (m MonsterInfo) String() string {
	return m.Difficulty.String() + m.Name
}

type MonsterList []*MonsterInfo
type MonsterMap map[string]*MonsterInfo

func (l MonsterList) Len() int           { return len(l) }
func (l MonsterList) Less(i, j int) bool { return l[i].Code < l[j].Code }
func (l MonsterList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

var renderer *render.Render
var monsters MonsterMap

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
					if i%4 == 3 {
						return true
					}
					return false
				},
			},
			{
				"GetHabitat": func(name string, field int) [7]int {
					return monsters[name].Habitat[field]
				},
			},
			{
				"GetTemperedMonster": func(name string) MonsterInfo {
					if m := monsters["역전 "+name]; m != nil {
						return *m
					} else if m := monsters["상처입은 "+name]; m != nil {
						return *m
					} else {
						return MonsterInfo{}
					}
				},
			},
		},
	})
	monsters = make(map[string]*MonsterInfo)
}

func cannotSee(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	renderer.HTML(w, http.StatusOK, "monster_lists", map[string]interface{}{
		"flv": 1,
		"wlv": 1,
		"clv": 1,
		"rlv": 1,
		"llv": 1,
	})
}

func whatYouCannotSeeHere(result *[FieldMax]MonsterList, lvs []int) {
	for _, monsterInfo := range monsters {
		for field, currentLv := range lvs {
			if findleft(monsterInfo.Habitat[field], currentLv-2) && !findright(monsterInfo.Habitat[field], currentLv-1) {
				result[field] = append(result[field], monsterInfo)
			}
		}
	}
}

func whatYouCannotSee(result *MonsterList, mlist *[FieldMax]MonsterList, lvs []int) {
	for field := Forest; field < FieldMax; field++ {
		for _, monsterInfo := range mlist[field] {
			appearsInOtherField := false
			for otherField := Forest; otherField < FieldMax; otherField++ {
				if otherField == field {
					continue
				}
				if findright(monsterInfo.Habitat[otherField], lvs[otherField]-1) {
					appearsInOtherField = true
					break
				}
			}
			if !appearsInOtherField && !contains(*result, monsterInfo) {
				*result = append(*result, monsterInfo)
			}
		}
	}
}

func calculateLists(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	req.ParseForm()
	fLv, _ := strconv.Atoi(req.FormValue("forest_lev")[0:])
	wLv, _ := strconv.Atoi(req.FormValue("wild_lev")[0:])
	cLv, _ := strconv.Atoi(req.FormValue("coral_lev")[0:])
	rLv, _ := strconv.Atoi(req.FormValue("rotten_lev")[0:])
	lLv, _ := strconv.Atoi(req.FormValue("lava_lev")[0:])

	currentLvs := []int{fLv, wLv, cLv, rLv, lLv}
	var cannotSeeLists [FieldMax]MonsterList
	var totalLists MonsterList
	var cannotSeeIfFields [FieldMax]MonsterList
	var cannotSeeIfs [FieldMax]MonsterList

	whatYouCannotSeeHere(&cannotSeeLists, currentLvs)
	whatYouCannotSee(&totalLists, &cannotSeeLists, currentLvs)
	sort.Sort(totalLists)
	for _, l := range cannotSeeLists {
		sort.Sort(l)
	}

	futureLvs := make([]int, FieldMax)
	for field := Forest; field < FieldMax; field++ {
		if currentLvs[field] >= 7 {
			continue
		}
		copy(futureLvs, currentLvs)
		futureLvs[field]++

		for i := Forest; i < FieldMax; i++ {
			cannotSeeIfFields[i] = cannotSeeIfFields[i][:0]
		}
		whatYouCannotSeeHere(&cannotSeeIfFields, futureLvs)
		whatYouCannotSee(&cannotSeeIfs[field], &cannotSeeIfFields, futureLvs)
	}

	for _, l := range cannotSeeIfs {
		sort.Sort(l)
	}

	renderer.HTML(w, http.StatusOK, "monster_lists",
		map[string]interface{}{
			"flv":                 fLv,
			"wlv":                 wLv,
			"clv":                 cLv,
			"rlv":                 rLv,
			"llv":                 lLv,
			"NotAppearList":       totalLists,
			"ForestNotAppearList": cannotSeeLists[Forest],
			"WildNotAppearList":   cannotSeeLists[Wildspire],
			"CoralNotAppearList":  cannotSeeLists[Coral],
			"RottenNotAppearList": cannotSeeLists[Rotten],
			"LavaNotAppearList":   cannotSeeLists[Lava],
			"NotIfForestUp":       cannotSeeIfs[Forest],
			"NotIfWildUp":         cannotSeeIfs[Wildspire],
			"NotIfCoralUp":        cannotSeeIfs[Coral],
			"NotIfRottenUp":       cannotSeeIfs[Rotten],
			"NotIfLavaUp":         cannotSeeIfs[Lava],
		})
}

func displayAppearLists(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	var appearList [FieldMax]MonsterList

	for _, info := range monsters {
		if info.Difficulty == Normal {
			m := MonsterInfo{}
			info.Copy(&m)

			tempered := monsters["역전 "+m.Name]

			if m.Name == "얀가루루가" {
				tempered = monsters["상처입은 "+m.Name]
			}

			if tempered != nil {
				for i := Forest; i < FieldMax; i++ {
					for j := 0; j < 7; j++ {
						if tempered.Habitat[i][j] != 0 {
							m.Habitat[i][j] = tempered.Habitat[i][j] + 3
						}
					}
				}
			}

			for i := Forest; i < FieldMax; i++ {
				for j := 0; j < 7; j++ {
					if m.Habitat[i][j] > 0 {
						appearList[i] = append(appearList[i], &m)
						break
					}
				}
			}
		}
	}

	for i := Forest; i < FieldMax; i++ {
		sort.Sort(appearList[i])
	}
	renderer.HTML(w, http.StatusOK, "habitat_list",
		map[string]interface{}{
			"ForestList": appearList[Forest],
			"WildList":   appearList[Wildspire],
			"CoralList":  appearList[Coral],
			"RottenList": appearList[Rotten],
			"LavaList":   appearList[Lava],
		})
}

func itemlist(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	var list MonsterList
	for _, monster := range monsters {
		if monster.Difficulty == Normal {
			list = append(list, monster)
		}
	}

	sort.Sort(list)

	renderer.HTML(w, http.StatusOK, "item_list",
		map[string]interface{}{
			"ItemList": list,
		})
}

func updateData(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	getDataFromGoogleSheet()
	saveDataToJson()
	http.Redirect(w, req, "/", http.StatusFound)
}

func saveDataToJson() {
	file, _ := json.MarshalIndent(monsters, "", " ")
	_ = ioutil.WriteFile("data/monsters.json", file, 0644)
}

func getDataFromJson() error {
	jsonFile, err := os.Open("data/monsters.json")
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &monsters)
	return nil
}

func displayInfo(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	target := monsters[ps.ByName("name")]
	m := MonsterInfo{}

	if target != nil {
		target.Copy(&m)

		tempered := monsters["역전 "+m.Name]

		if m.Name == "얀가루루가" {
			tempered = monsters["상처입은 "+m.Name]
		}

		if tempered != nil {
			for i := Forest; i < FieldMax; i++ {
				for j := 0; j < 7; j++ {
					if tempered.Habitat[i][j] != 0 {
						m.Habitat[i][j] = tempered.Habitat[i][j] + 3
					}
				}
			}
		}
	}

	renderer.HTML(w, http.StatusOK, "monster_info",
		map[string]interface{}{
			"Info": m,
		})
}

func main() {
	err := getDataFromJson()
	if err != nil {
		getDataFromGoogleSheet()
		saveDataToJson()
		fmt.Println("Reading from google sheet...")
	}

	router := httprouter.New()

	router.GET("/update", updateData)
	router.GET("/", itemlist)
	router.GET("/itemlist", itemlist)
	router.GET("/appearlist", displayAppearLists)
	router.GET("/youcannotsee", cannotSee)
	router.POST("/youcannotsee", calculateLists)
	router.GET("/monster_info/:name", displayInfo)
	router.ServeFiles("/data/*filepath", http.Dir("./data"))

	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":8080")
}
