package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

type Site struct {
	ID                     int
	x                      int
	y                      int
	radius                 int
	ignore1                int
	ignore2                int
	structureType          int
	owner                  int
	param1                 int
	param2                 int
	distanceFromMyQueen    int
	distanceFromEnemyQueen int
}
type Sites map[int]*Site
type Unit struct {
	x        int
	y        int
	health   int
	owner    int
	unitType int
}
type Game struct {
	numberOfBarracks BarracksCount
	numberOfTowers   int
	numberOfMyUnits  UnitCount
	touchedSite      int
	gold             int
	myQueen          Unit
	enemyQueen       Unit
	myUnits          []Unit
	enemyUnits       []Unit
	sites            Sites
}
type BarracksCount map[int]int
type UnitCount map[int]int

const MaxKnightBarracks = 2
const MaxArcherBarracks = 0
const MaxTowers = 3
const MaxKnights = 8
const MaxArcher = 4

const Tower = 1
const Barracks = 2
const Queen = -1
const Friendly = 0
const Neutral = -1
const Enemy = 1
const Knight = 0
const Archer = 1
const Giant = 2

func main() {
	var game Game
	game.numberOfBarracks = BarracksCount{
		Knight: 0,
		Archer: 0,
		Giant:  0,
	}
	var numSites int
	fmt.Scan(&numSites)
	game.sites = make(Sites)
	for i := 0; i < numSites; i++ {
		site := Site{}
		fmt.Scan(&site.ID, &site.x, &site.y, &site.radius)
		site.owner = -1 // Default no owner
		game.sites[site.ID] = &site
		fmt.Fprintln(os.Stderr, site.ID, site.x, site.y, site.radius)
	}
	for {
		fmt.Scan(&game.gold, &game.touchedSite)

		for i := 0; i < numSites; i++ {
			// structureType: -1 = No structure, 2 = Barracks
			// owner: -1 = No structure, 0 = Friendly, 1 = Enemy
			var siteID, ignore1, ignore2, structureType, owner, param1, param2 int
			fmt.Scan(&siteID, &ignore1, &ignore2, &structureType, &owner, &param1, &param2)
			game.changeSite(siteID, structureType, owner, param1, param2)
		}
		var numUnits int
		game.myUnits = []Unit{}
		game.enemyUnits = []Unit{}
		game.numberOfMyUnits = UnitCount{
			Knight: 0,
			Archer: 0,
			Giant:  0,
		}
		fmt.Scan(&numUnits)
		for i := 0; i < numUnits; i++ {
			// unitType: -1 = QUEEN, 0 = KNIGHT, 1 = ARCHER
			var x, y, owner, unitType, health int
			fmt.Scan(&x, &y, &owner, &unitType, &health)
			game.buildUnit(x, y, owner, unitType, health)
		}
		game.sites.setDistancesFromQueens(game.myQueen, game.enemyQueen)

		fmt.Println(game.GetQueenAction())
		fmt.Println(game.getTrainAction())
	}
}

func (game Game) buildUnit(x int, y int, owner int, unitType int, health int) {
	newUnit := Unit{
		x:        x,
		y:        y,
		owner:    owner,
		unitType: unitType,
		health:   health,
	}
	if unitType == Queen {
		if owner == Friendly {
			game.myQueen = newUnit
		} else {
			game.enemyQueen = newUnit
		}
	} else {
		if owner == Friendly {
			game.myUnits = append(game.myUnits, newUnit)
			game.numberOfMyUnits[newUnit.unitType]++
		} else {
			game.enemyUnits = append(game.enemyUnits, newUnit)
		}
	}
}

func (game Game) changeSite(ID int, structureType int, owner int, param1 int, param2 int) {
	// Substract changing sites from count
	//fmt.Fprintln(os.Stderr, "changin site ", ID, structureType, param2)
	if structureType != game.sites[ID].structureType && owner == Friendly {
		if structureType == Tower {
			game.numberOfTowers++
			fmt.Fprintln(os.Stderr, "Add Game towers, total", game.numberOfTowers)
		}
		if structureType == Barracks {
			game.numberOfBarracks[param2]++
			fmt.Fprintln(os.Stderr, "Add number to ", param2, " To get total of ", strconv.Itoa(game.numberOfBarracks[param2]))
		}
	}
	game.sites[ID].structureType = structureType
	game.sites[ID].owner = owner
	game.sites[ID].param1 = param1
	game.sites[ID].param2 = param2
}

func (game Game) getTrainAction() string {
	// Decide train step
	trainingLocations := ""
	if game.gold > 80 && game.numberOfMyUnits[Knight] < MaxKnights {
		closestAttackKnightsSiteID, _ := game.sites.findClosestSiteID(game.enemyQueen.x, game.enemyQueen.y, true, false, false, true, false, false)
		if closestAttackKnightsSiteID != -1 {
			trainingLocations = trainingLocations + " " + strconv.Itoa(closestAttackKnightsSiteID)
			game.gold = game.gold - 80
		}
	}
	if game.gold > 100 && game.numberOfMyUnits[Archer] < MaxArcher {
		closestArcherBarracksID, _ := game.sites.findClosestSiteID(game.myQueen.x, game.myQueen.y, true, false, false, false, false, true)
		if closestArcherBarracksID != -1 {
			trainingLocations = trainingLocations + " " + strconv.Itoa(closestArcherBarracksID)
			game.gold = game.gold - 100
		}
	}

	return "TRAIN" + trainingLocations
}

func (game Game) GetQueenAction() string {
	// Decide build step
	if game.touchedSite != Neutral && game.sites[game.touchedSite].owner == Neutral {
		fmt.Fprintln(os.Stderr, "knights and archer barracks", game.numberOfBarracks[Knight], game.numberOfBarracks[Archer])
		buildType := "BARRACKS-KNIGHT"
		if game.numberOfBarracks[Knight] >= MaxKnightBarracks {
			buildType = "BARRACKS-ARCHER"
		}
		if game.numberOfBarracks[Knight] >= MaxKnightBarracks && game.numberOfBarracks[Archer] >= MaxArcherBarracks {
			buildType = "TOWER"
		}
		return "BUILD " + strconv.Itoa(game.touchedSite) + " " + buildType
	} else {
		// Decide move step
		closestSiteID, distance := game.sites.findClosestSiteID(game.myQueen.x, game.myQueen.y, false, false, true, false, false, false)
		if distance > 500 {
			closestSiteID = -1
		}
		fmt.Fprintln(os.Stderr, "Number of game towers", game.numberOfTowers)
		if closestSiteID == -1 || game.numberOfTowers >= MaxTowers {
			closestSiteID, _ = game.sites.findClosestSiteID(game.myQueen.x, game.myQueen.y, true, false, false, false, true, false)
			fmt.Fprintln(os.Stderr, "Moving to closest friendly tower!", closestSiteID)
		}
		fmt.Fprintln(os.Stderr, "closestSiteID ", closestSiteID, "xy", game.sites[closestSiteID].x, game.sites[closestSiteID].y)
		return "MOVE " + strconv.Itoa(game.sites[closestSiteID].x) + " " + strconv.Itoa(game.sites[closestSiteID].y)
	}
}

func (sites Sites) findSiteIDs(owned bool, enemy bool, neutral bool) []int {
	IDs := []int{}
	for id, site := range sites {
		if (owned == true && site.owner == Friendly) ||
			(enemy == true && site.owner == Enemy) ||
			(neutral == true && site.owner == Neutral) {
			IDs = append(IDs, id)
		}
	}
	return IDs
}

func (sites Sites) setDistancesFromQueens(myQueen Unit, enemyQueen Unit) {
	for _, site := range sites {
		distanceFromMyQueen := distanceBetween(myQueen.x, myQueen.y, site.x, site.y)
		site.distanceFromMyQueen = int(distanceFromMyQueen)
		distanceFromEnemyQueen := distanceBetween(enemyQueen.x, enemyQueen.y, site.x, site.y)
		site.distanceFromEnemyQueen = int(distanceFromEnemyQueen)
	}
}

func (sites Sites) findClosestSiteID(x int, y int, owned bool, enemy bool, neutral bool, knightBarracks bool, tower bool, archerBarracks bool) (int, float64) {
	//fmt.Fprintln(os.Stderr, "Checking for location x", x, " and Y:", y)
	closestSiteID := -1
	closestDistance := 9999999.0
	for id, site := range sites {
		if owned == false && site.owner == Friendly {
			continue
		}
		if enemy == false && site.owner == Enemy {
			continue
		}
		if neutral == false && site.owner == Neutral {
			continue
		}
		if knightBarracks == false && site.structureType == Barracks && site.param2 == Knight {
			continue
		}
		if archerBarracks == false && site.structureType == Barracks && site.param2 == Archer {
			continue
		}
		if tower == false && site.structureType == Tower {
			continue
		}

		distance := distanceBetween(x, y, site.x, site.y)
		if distance < closestDistance {
			closestSiteID = id
			closestDistance = distance
			//fmt.Fprintln(os.Stderr, "SiteID", site.ID, " has a distance of ", distance)
		}
	}
	return closestSiteID, closestDistance
}

func distanceBetween(x int, y int, targetX int, targetY int) float64 {
	return math.Sqrt(math.Pow(float64(x-targetX), 2) + math.Pow(float64(y-targetY), 2))
}
