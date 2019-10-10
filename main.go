package main

import (
	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"
	"github.com/unrolled/render"
	"html/template"
	"net/http"
	"sort"
	"strconv"
)

const (			// Fields
	Forest = iota
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
	Normal = iota
	Tempered
)

type MonsterInfo struct {
	Code       int
	Name       string
	Difficulty difficulty
	Item       string
	Habitat    [4][7]int
}

func (m MonsterInfo) String() string {
	return m.Difficulty.String() + m.Name
}

func (m MonsterInfo) Copy(l *MonsterInfo) {
	for i := Forest; i <= Rotten; i++ {
		for j := 0; j < 7; j++ {
			l.Habitat[i][j] = m.Habitat[i][j]
		}
	}
	l.Code = m.Code
	l.Name = m.Name
	l.Difficulty = m.Difficulty
	l.Item = m.Item
}

type MonsterList []*MonsterInfo

func (l MonsterList) Len() int	{ return len(l) }
func (l MonsterList) Less(i, j int) bool { return l[i].Code < l[j].Code }
func (l MonsterList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

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

func whatYouCannotSeeHere(result *[4]MonsterList, lvs []int) {
	for _, monsterInfo := range monsters {
		for field, current_lv := range lvs {
			if findleft(monsterInfo.Habitat[field], current_lv-2) && !findright(monsterInfo.Habitat[field], current_lv-1) {
				result[field] = append(result[field], monsterInfo)
			}
		}
	}
}

func whatYouCannotSee(result *MonsterList, mlist *[4]MonsterList, lvs []int) {
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

	currentLvs := []int{fLv, wLv, cLv, rLv}
	var cannotSeeLists [4]MonsterList
	var totalLists MonsterList
	var cannotSeeIfFields [4]MonsterList
	var cannotSeeIfs [4]MonsterList

	whatYouCannotSeeHere(&cannotSeeLists, currentLvs)
	whatYouCannotSee(&totalLists, &cannotSeeLists, currentLvs)
	sort.Sort(totalLists)
	for _, l := range cannotSeeLists {
		sort.Sort(l)
	}

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

	for _, l := range cannotSeeIfs {
		sort.Sort(l)
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

func displayAppearLists(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	var appearList [4]MonsterList

	for _, info := range monsters {
		if info.Difficulty == Tempered {
			if info.Name == "상처입은 얀가루루가" {
				// TODO
			}
		} else {
			m := MonsterInfo{}
			info.Copy(&m)

			if tempered := monsters["역전 " + info.Name]; tempered != nil {
				for i := Forest; i <= Rotten; i++ {
					for j := 0; j < 7; j++ {
						if tempered.Habitat[i][j] != 0 {
							m.Habitat[i][j] = tempered.Habitat[i][j] + 3
						}
					}
				}
			}

			for i := Forest; i <= Rotten; i++ {
				for j := 0; j < 7; j++ {
					if m.Habitat[i][j] > 0 {
						appearList[i] = append(appearList[i], &m)
					}
				}
			}
		}
	}
	renderer.HTML(w, http.StatusOK, "habitat_list",
		map[string]interface{}{
			"ForestList": appearList[Forest],
			"WildList": appearList[Wildspire],
			"CoralList": appearList[Coral],
			"RottenList": appearList[Rotten],
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
	router.GET("/appearlist", displayAppearLists)

	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":8080")
}
