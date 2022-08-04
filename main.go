package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
)

type Site struct {
	ID                           int
	position                     Position
	radius                       int
	ignore1                      int
	ignore2                      int
	structureType                int
	owner                        int
	param1                       int
	param2                       int
	distanceFromMyQueen          int
	distanceFromEnemyQueen       int
	distanceFromStartingLocation int
	maxMineSize                  int
	goldRemaining                int
}

type Sites map[int]*Site

type Unit struct {
	position Position
	health   int
	owner    int
	unitType int
}

type Game struct {
	numberOfBarracks                *BarracksCount
	numberOfTowers                  int
	numberOfMyUnits                 UnitCount
	touchedSite                     int
	gold                            int
	remainingGold                   int
	myQueen                         Unit
	enemyQueen                      Unit
	myUnits                         []Unit
	enemyUnits                      []Unit
	sites                           Sites
	turn                            int
	startingHealth                  int
	myQueenStartingPosition         Position
	sitesOrderedByDistanceFromStart SitesByDistanceFromStart
	enemyTowers                     Sites
	unitBuildQueue                  []int
	strategy                        int
}

type Position struct {
	x int
	y int
}

type BarracksCount map[int]int
type UnitCount map[int]int

/************************************************
Configurable values
*************************************************/

// MaxKnightBarracks How many Knight Barracks do we build?
const MaxKnightBarracks = 1

// MaxGoldMines How many Gold mines should we have?
const MaxGoldMines = 3

// MaxArcherBarracks How many Archer Barracks should we build?
const MaxArcherBarracks = 0

// MaxTowers How many Towers should we have at all times?
const MaxTowers = 3

// MaxKnights How many Knights do we want to have at one time?
const MaxKnights = 12

// MaxArcher How many Archers do we want to have at one time?
const MaxArcher = 4

// MinTowerRangeConstruction Until what range should we "grow" our towers?
const MinTowerRangeConstruction = 400

// IgnoreGoldmine Change this goldmine into a Tower, if gold remaining is less than this.
const IgnoreGoldmine = 10

// Layered sites constants

/************************************************
Building Constants
*************************************************/

const Goldmine = 0
const Tower = 1
const Barracks = 2

// GiantBarracks (is actually 2 in-game)
const GiantBarracks = 3

// ArcherBarracks (is actually 2 in-game)
const ArcherBarracks = 4

/************************************************
Unit Constants
*************************************************/

const Queen = -1
const Knight = 0
const Archer = 1
const Giant = 2

/************************************************
Unit Costs
*************************************************/

const KnightCost = 80
const ArcherCost = 100
const GiantCost = 140

/************************************************
Owner Constants
*************************************************/

const Friendly = 0
const Neutral = -1
const Enemy = 1

/************************************************
Field Settings
*************************************************/

const FieldWidth = 1920
const FieldHeight = 1000

/************************************************
Strategy
*************************************************/

const DefaultStrategy = 0
const TooManyTowersStrategy = 1

/************************************************
MAIN FUNCTION
*************************************************/
func main() {
	game := Game{
		numberOfBarracks: &BarracksCount{
			Knight: 0,
			Archer: 0,
			Giant:  0,
		},
		numberOfTowers:  0,
		numberOfMyUnits: UnitCount{},
		touchedSite:     0,
		gold:            0,
		myQueen:         Unit{},
		enemyQueen:      Unit{},
		myUnits:         []Unit{},
		enemyUnits:      []Unit{},
		enemyTowers:     Sites{},
		sites:           nil,
		turn:            1,
		strategy:        DefaultStrategy,
	}

	var numSites int
	fmt.Scan(&numSites)
	game.sites = make(Sites)
	for i := 0; i < numSites; i++ {
		site := &Site{}
		fmt.Scan(&site.ID, &site.position.x, &site.position.y, &site.radius)
		site.owner = Neutral // Default no owner
		game.sites[site.ID] = site
		//game.sites[site.ID].distanceFromStartingLocation = distanceBetween(game.sites[site.ID].position, )
		fmt.Fprintln(os.Stderr, site.ID, site.position.x, site.position.y, site.radius)
	}
	for {
		fmt.Scan(&game.gold, &game.touchedSite)

		for i := 0; i < numSites; i++ {
			var siteID, goldRemaining, maxMineSize, structureType, owner, param1, param2 int
			fmt.Scan(&siteID, &goldRemaining, &maxMineSize, &structureType, &owner, &param1, &param2)
			game.changeSite(siteID, structureType, owner, param1, param2, goldRemaining, maxMineSize)
			//fmt.Fprintln(os.Stderr, siteID, "maxsize", maxMineSize, goldRemaining)
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
			var x, y, owner, unitType, health int
			fmt.Scan(&x, &y, &owner, &unitType, &health)
			game.buildUnit(x, y, owner, unitType, health)
		}
		if game.turn == 1 {
			game.myQueenStartingPosition = Position{
				x: game.myQueen.position.x,
				y: game.myQueen.position.y,
			}
			game.startingHealth = game.myQueen.health
			game.setSitesOrderedByDistanceFromStart()
		}
		game.sites.setDistancesFromQueens(game.myQueen, game.enemyQueen)
		fmt.Fprintln(os.Stderr, "We have", game.numberOfMyUnits[Knight], "Knights")

		game.remainingGold = game.calculateRemainingGold()
		fmt.Fprintln(os.Stderr, "Game Remaining Gold:", game.remainingGold)
		game.strategy = game.determineStrategy()
		fmt.Fprintln(os.Stderr, "Game Strategy:", game.strategy)

		fmt.Println(game.getQueenAction())
		fmt.Println(game.getTrainAction())
		fmt.Fprintln(os.Stderr, "BuildOrder:", game.getBuildOrder())
		game.turn++
		fmt.Fprintln(os.Stderr, "There are", len(game.enemyTowers), "enemyTowers")
	}
}

type SiteAndDistance struct {
	ID    int
	value int
}
type SitesByDistanceFromStart []SiteAndDistance

func (d SitesByDistanceFromStart) Len() int {
	return len(d)
}
func (d SitesByDistanceFromStart) Less(i, j int) bool {
	return d[i].value < d[j].value
}
func (d SitesByDistanceFromStart) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func returnSortedByDistance(sites map[int]*Site) SitesByDistanceFromStart {
	// Copy entries into a slice.
	slice := make(SitesByDistanceFromStart, 0, len(sites))
	for ID, value := range sites {
		slice = append(slice, SiteAndDistance{ID, value.distanceFromStartingLocation})
	}

	// Sort the slice.
	sort.Sort(slice)
	return slice
}

/************************************************
Game Methods
*************************************************/
func (game *Game) determineStrategy() int {
	strategy := DefaultStrategy
	if len(game.enemyTowers) > 3 && game.turn > 100 {
		strategy = TooManyTowersStrategy
	}
	return strategy
}

func (game *Game) calculateRemainingGold() int {
	subtract := 0
	for _, unitType := range game.unitBuildQueue {
		subtract += game.getCostOfUnit(unitType)
	}
	return game.gold - subtract
}

func (game *Game) setSitesOrderedByDistanceFromStart() {
	for ID := range game.sites {
		game.sites[ID].distanceFromStartingLocation = int(distanceBetween(game.sites[ID].position, game.myQueenStartingPosition))
	}
	game.sitesOrderedByDistanceFromStart = returnSortedByDistance(game.sites)
}

//func (game *Game) leftSideStart
func (game *Game) buildUnit(x int, y int, owner int, unitType int, health int) {
	newUnit := Unit{
		position: Position{
			x: x,
			y: y,
		},
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

func (game *Game) changeSite(ID int, structureType int, owner int, param1 int, param2 int, goldRemaining int, maxMineSize int) {
	structureType = getRealStructureType(structureType, param2)
	// Subtract changing sites from count
	// if game changed owner and the site was owner
	if game.sites[ID].owner != owner {
		if game.sites[ID].owner == Friendly {
			if game.sites[ID].getStructureType() == Tower {
				game.numberOfTowers--
				fmt.Fprintln(os.Stderr, "Substract game towers, total", game.numberOfTowers)
			}
			if game.sites[ID].getStructureType() == Barracks {
				(*game.numberOfBarracks)[param2]--
				fmt.Fprintln(os.Stderr, "Substract number to ", param2, " To get total of ", strconv.Itoa((*game.numberOfBarracks)[param2]))
			}
		} else {
			if game.sites[ID].getStructureType() == Tower {
				fmt.Fprintln(os.Stderr, "remove enemy tower")
				delete(game.enemyTowers, ID)
			}
		}
	}

	//fmt.Fprintln(os.Stderr, "changin site ", ID, structureType, param2)
	// Add towers
	if structureType != game.sites[ID].getStructureType() {
		if owner == Friendly {
			if structureType == Tower {
				game.numberOfTowers++
				fmt.Fprintln(os.Stderr, "Add Game towers, total", game.numberOfTowers)
			}
			if structureType == Barracks {
				(*game.numberOfBarracks)[param2]++
				fmt.Fprintln(os.Stderr, "Add number to ", param2, " To get total of ", strconv.Itoa((*game.numberOfBarracks)[param2]))
			}
		} else {
			if structureType == Tower {
				fmt.Fprintln(os.Stderr, "add enemy tower")
				game.enemyTowers[ID] = game.sites[ID]
			} else if game.sites[ID].getStructureType() == Tower {
				delete(game.enemyTowers, ID)
			}
		}
	}
	game.sites[ID].structureType = structureType
	game.sites[ID].owner = owner
	game.sites[ID].param1 = param1
	game.sites[ID].param2 = param2
	game.sites[ID].goldRemaining = goldRemaining
	game.sites[ID].maxMineSize = maxMineSize
}

func (game *Game) hasCountOfUnit(unitType int) int {
	count := 0
	for _, unit := range game.myUnits {
		if unit.unitType == unitType {
			count++
		}
	}
	return count
}

func (game *Game) getCostOfUnit(unitType int) int {
	cost := 0
	switch unitType {
	case Knight:
		cost = KnightCost
	case Archer:
		cost = ArcherCost
	case Giant:
		cost = GiantCost
	}

	if cost == 0 {
		fmt.Fprintln(os.Stderr, "Undefined cost of unit type:", unitType)
	}

	return cost
}

func (game *Game) getTrainAction() string {
	if game.strategy == DefaultStrategy {
		if len(game.enemyTowers) > 1 && game.remainingGold >= 160 && len(game.unitBuildQueue) == 0 {
			game.unitBuildQueue = append(game.unitBuildQueue, Knight, Knight)
		} else if len(game.enemyTowers) <= 1 && game.remainingGold >= 80 && len(game.unitBuildQueue) == 0 {
			game.unitBuildQueue = append(game.unitBuildQueue, Knight)
		}
	}
	if game.strategy == TooManyTowersStrategy {
		if len(game.unitBuildQueue) == 0 && game.remainingGold >= 200 {
			game.unitBuildQueue = append(game.unitBuildQueue, Giant)
		}
		if game.hasCountOfUnit(Giant) > 0 && game.remainingGold >= 80 {
			game.unitBuildQueue = append(game.unitBuildQueue, Knight)
		}
	}
	// Decide train step
	trainingLocations := ""
	if len(game.unitBuildQueue) > 0 {
		var unitToTrain int
		unitToTrain = game.unitBuildQueue[0]
		knightLocation := false
		giantLocation := false
		archerLocation := false
		if unitToTrain == Knight {
			knightLocation = true
		} else if unitToTrain == Giant {
			giantLocation = true
		} else if unitToTrain == Archer {
			archerLocation = true
		}
		closestAttackSiteID, _ := game.sites.findClosestSiteID(game.enemyQueen.position, true, false, false, knightLocation, false, archerLocation, false, giantLocation)
		// Found a location and can train here.
		if closestAttackSiteID != -1 && game.sites[closestAttackSiteID].param1 == 0 {
			trainingLocations = trainingLocations + " " + strconv.Itoa(closestAttackSiteID)
			game.unitBuildQueue = game.unitBuildQueue[1:]
		}
	}

	return "TRAIN" + trainingLocations
}

func (game *Game) getBuildCommand(siteID int, structureType int) string {
	fmt.Fprintln(os.Stderr, "Building ", structureType, " AT ", siteID)
	building := "MINE"
	switch structureType {
	case Barracks:
		building = "BARRACKS-KNIGHT"
	case GiantBarracks:
		building = "BARRACKS-GIANT"
	case ArcherBarracks:
		building = "BARRACKS-ARCHER"
	case Tower:
		building = "TOWER"
	}
	return "BUILD " + strconv.Itoa(siteID) + " " + building
}

func (game *Game) areEnemyUnitsNear(position Position) bool {
	for _, unit := range game.enemyUnits {
		distance := distanceBetween(position, unit.position)
		if distance < 150 {
			return true
		}
	}
	return false
}

func (game *Game) getQueenAction() string {
	fmt.Fprintln(os.Stderr, "TouchsiteID", game.touchedSite)
	areEnemiesNear := game.areEnemyUnitsNear(game.myQueen.position)
	if areEnemiesNear && game.numberOfTowers == 0 {
		fmt.Fprintln(os.Stderr, "Panic mode, build tower (enemies near and no towers)")
		// There are enemies close, and we have no defences!
		closestSiteID, _ := game.sites.findClosestSiteID(game.myQueen.position, false, true, true, false, false, false, false, false)
		if game.touchedSite == closestSiteID {
			// Build the Tower! you're close enough
			return game.getBuildCommand(game.touchedSite, Tower)
		}
	}

	if game.myQueen.health < 10 {
		return game.getMoveToEdge()
	}

	// Follow build order when touching a site.
	buildOrder := game.getBuildOrder()

	//if game.touchedSite != Neutral && game.sites[game.touchedSite].owner == Neutral {
	if game.touchedSite != Neutral {
		for order, siteAndDistance := range game.sitesOrderedByDistanceFromStart {
			if order >= len(buildOrder) {
				continue
			}
			//fmt.Fprintln(os.Stderr, "Order", order, game.sitesOrderedByDistanceFromStart[order].ID)
			//fmt.Fprintln(os.Stderr, "buildorder", buildOrder[order])
			//fmt.Fprintln(os.Stderr, "ID", game.sites[game.sitesOrderedByDistanceFromStart[order].ID].ID)
			if siteAndDistance.ID == game.touchedSite &&
				game.sites[game.sitesOrderedByDistanceFromStart[order].ID].getStructureType() != buildOrder[order] {
				if buildOrder[order] == Goldmine && (game.areEnemyUnitsNear(game.sites[game.touchedSite].position) || game.sites[game.touchedSite].goldRemaining <= IgnoreGoldmine) {
					// If the gold has run out or enemies are near, build a Tower instead
					fmt.Fprintln(os.Stderr, "Replace goldmine with Tower")
					return game.getBuildCommand(game.touchedSite, Tower)
				}
				fmt.Fprintln(os.Stderr, "Build the build order on touched site, seeing:", game.sites[game.sitesOrderedByDistanceFromStart[order].ID].getStructureType(), "Building:", buildOrder[order])
				return game.getBuildCommand(game.touchedSite, buildOrder[order])
			}
		}
	}

	// Upgrade mine logic
	if game.touchedSite != Neutral &&
		game.sites[game.touchedSite].owner == Friendly &&
		game.sites[game.touchedSite].maxMineSize != game.sites[game.touchedSite].param1 &&
		game.sites[game.touchedSite].getStructureType() == Goldmine &&
		game.sites[game.touchedSite].goldRemaining > IgnoreGoldmine {
		fmt.Fprintln(os.Stderr, "Upgrade Goldmine")
		return game.getBuildCommand(game.touchedSite, Goldmine)
	}

	// Upgrade Tower logic
	if game.touchedSite != Neutral &&
		game.sites[game.touchedSite].owner == Friendly &&
		game.sites[game.touchedSite].getStructureType() == Tower &&
		game.sites[game.touchedSite].param2 < MinTowerRangeConstruction {
		fmt.Fprintln(os.Stderr, "Upgrade Tower")
		return game.getBuildCommand(game.touchedSite, Tower)
	}

	// Move to next build order location
	for order := range buildOrder {
		next := len(buildOrder) - 1 - order
		targetSite := game.sites[game.sitesOrderedByDistanceFromStart[next].ID]

		fmt.Fprintln(os.Stderr, "compare structure type", targetSite.ID, targetSite.getStructureType(), buildOrder[next])
		//if targetSite.owner != Friendly && targetSite.structureType != structureType {
		if targetSite.getStructureType() != buildOrder[next] {
			fmt.Fprintln(os.Stderr, "Move to next build order #", next)
			return game.getMoveOrderForSite(targetSite)
		}
	}

	// Everything's done! Move to safety (aka your corner of the map)
	fmt.Fprintln(os.Stderr, "Move to edge!")
	return game.getMoveToEdge()
}

func (game *Game) getBuildOrder() []int {
	buildOrder := []int{Goldmine, Goldmine, Goldmine, Goldmine, Tower, Tower, Tower, Goldmine, Tower, Barracks}
	// If the Goldmine has been emptied out, replace with a Tower
	for order, structureType := range buildOrder {
		if structureType == Goldmine && game.sites[game.sitesOrderedByDistanceFromStart[order].ID].goldRemaining <= IgnoreGoldmine {
			buildOrder[order] = Tower
		}
	}

	if game.strategy == TooManyTowersStrategy {
		buildOrder[0] = GiantBarracks
	}
	return buildOrder
}

func (game *Game) getMoveOrderForSite(site *Site) string {
	return "MOVE " + strconv.Itoa(site.position.x) + " " + strconv.Itoa(site.position.y)
}

func (game *Game) getMoveToEdge() string {
	edgePosition := game.findClosestEdge()
	return "MOVE " + strconv.Itoa(edgePosition.x) + " " + strconv.Itoa(edgePosition.y)
}

//func (game *Game) getMoveToClosestFriendlyTower() string {
//closestSiteID, _ = game.sites.findClosestSiteID(game.myQueen.position.x, game.myQueen.position.y, true, false, false, false, true, false)
//fmt.Fprintln(os.Stderr, "Moving to closest friendly tower!", closestSiteID)
//}

func (game *Game) findClosestEdge() Position {
	x := 0
	y := 0
	if game.myQueenStartingPosition.x > (FieldWidth / 2) {
		x = FieldWidth
		y = FieldHeight
	}

	return Position{
		x: x,
		y: y,
	}
}

/************************************************
Sites Methods
*************************************************/
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
		distanceFromMyQueen := distanceBetween(myQueen.position, site.position)
		site.distanceFromMyQueen = int(distanceFromMyQueen)
		distanceFromEnemyQueen := distanceBetween(enemyQueen.position, site.position)
		site.distanceFromEnemyQueen = int(distanceFromEnemyQueen)
	}
}

func (sites Sites) findClosestSiteID(position Position, owned bool, enemy bool, neutral bool, knightBarracks bool, tower bool, archerBarracks bool, goldmine bool, giantBarracks bool) (int, float64) {
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
		if knightBarracks == false && site.getStructureType() == Barracks {
			continue
		}
		if archerBarracks == false && site.getStructureType() == ArcherBarracks {
			continue
		}
		if giantBarracks == false && site.getStructureType() == GiantBarracks {
			continue
		}
		if tower == false && site.getStructureType() == Tower {
			continue
		}
		if goldmine == false && site.getStructureType() == Goldmine {
			continue
		}

		distance := distanceBetween(position, site.position)
		if distance < closestDistance {
			closestSiteID = id
			closestDistance = distance
			//fmt.Fprintln(os.Stderr, "SiteID", site.ID, " has a distance of ", distance)
		}
	}
	return closestSiteID, closestDistance
}

/************************************************
Site Methods
*************************************************/

func (site Site) getStructureType() int {
	return getRealStructureType(site.structureType, site.param2)
}

/************************************************
Helper functions
*************************************************/

func getRealStructureType(structureType int, param2 int) int {
	if structureType == Barracks && param2 == Giant {
		structureType = GiantBarracks
	} else if structureType == Barracks && param2 == Archer {
		structureType = ArcherBarracks
	}

	return structureType
}

func distanceBetween(fromPosition Position, targetPosition Position) float64 {
	return math.Sqrt(math.Pow(float64(fromPosition.x-targetPosition.x), 2) + math.Pow(float64(fromPosition.y-targetPosition.y), 2))
}
