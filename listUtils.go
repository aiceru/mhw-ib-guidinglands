package main

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

func contains(list []*MonsterInfo, m *MonsterInfo) bool {
	for _, i := range list {
		if i == m {
			return true
		}
	}
	return false
}
