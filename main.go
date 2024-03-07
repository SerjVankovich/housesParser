package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	//"houseParser/osmreader"
	"houseParser/utils"
	"math"
	"os"

	"github.com/paulmach/osm"
	"github.com/twpayne/go-geom"
)

func main() {
	//osm := osmreader.LoadOsm("petrozavodsk.osm")
	FindNearestCrossroads()
}

func PrintNodesWithAmenity(osm *osm.OSM) {
	for _, v := range osm.Nodes {
		amenity := v.Tags.Find("amenity")
		if amenity != "" {
			println(v.ID, amenity)
		}
	}
}

type Building struct {
	CenterLat float64 `json:"lat"`
	CenterLon float64 `json:"lon"`
	Area      float64 `json:"area"`
	Address   string  `json:"address"`
}

type CrossRoad struct {
	Buildings []*Building `json:"buildings"`
	Node      *osm.Node   `json:"Node"`
}

func LoadHouses(osm *osm.OSM) {
	var buildings []*Building
	for _, v := range osm.Ways {
		building := v.Tags.Find("building")
		if building != "" {
			street := v.Tags.Find("addr:street")
			houseNumber := v.Tags.Find("addr:housenumber")
			levels := ParseLevels(v.Tags.Find("building:levels"))
			var coords []geom.Coord
			var sumLat float64 = 0
			var sumLon float64 = 0
			for _, wayNode := range v.Nodes {
				node, err := FindNode(wayNode.ID, osm.Nodes)
				utils.ProcessError(err)
				point := geom.Coord{node.Lat, node.Lon}
				coords = append(coords, point)
				sumLat += node.Lat
				sumLon += node.Lon
			}
			numNodes := len(v.Nodes)
			area := CalculatePolygonArea(coords) * float64(levels)
			buildings = append(buildings, &Building{sumLat / float64(numNodes), sumLon / float64(numNodes), area, street + " " + houseNumber})
		}
	}
	p, err := json.Marshal(buildings)
	utils.ProcessError(err)

	f, err := os.Create("buildings.json")
	utils.ProcessError(err)
	defer f.Close()
	f.Write(p)
}

func FindNearestCrossroads() {
	var buildings []*Building
	var crossroads []*osm.Node
	var weightedCrossRoads = map[int64]*CrossRoad{}

	LoadJson("buildings.json", &buildings)
	LoadJson("crossroads.json", &crossroads)

	for _, crossroad := range crossroads {
		weightedCrossRoads[int64(crossroad.ID)] = &CrossRoad{[]*Building{}, crossroad}
	}

	for _, building := range buildings {
		minCrossRoad := crossroads[0]
		minDistance := Distance([2]float64{building.CenterLon, building.CenterLat}, minCrossRoad.Point())
		for _, crossroad := range crossroads {
			distance := Distance([2]float64{building.CenterLon, building.CenterLat}, crossroad.Point())
			if distance < minDistance {
				minDistance = distance
				minCrossRoad = crossroad
			}
		}
		crossroad, ok := weightedCrossRoads[int64(minCrossRoad.ID)]
		if ok {
			crossroad.Buildings = append(crossroad.Buildings, building)
		}
	}

	crossroadsToSave := []*CrossRoad{}
	for _, value := range weightedCrossRoads {
		crossroadsToSave = append(crossroadsToSave, value)
	}

	p, err := json.Marshal(crossroadsToSave)
	utils.ProcessError(err)

	f, err := os.Create("weightedCrossroads.json")
	utils.ProcessError(err)
	defer f.Close()
	f.Write(p)

}

func Distance(dot1, dot2 [2]float64) float64 {
	sum := 0.0
	for i, c := range dot1 {
		sum += math.Pow((c - dot2[i]), 2)
	}
	return math.Sqrt(sum)
}

func LoadJson(filename string, store any) {
	f, err := os.Open(filename)
	utils.ProcessError(err)

	bytes, err := io.ReadAll(f)
	utils.ProcessError(err)
	err = json.Unmarshal(bytes, store)
	utils.ProcessError(err)
	f.Close()
}

func ParseLevels(levels string) int {
	if levels == "" {
		return 1
	}
	intLevel, err := strconv.Atoi(levels)
	if err != nil {
		return 1
	}
	return intLevel
}

func PrintDotsOfBuilding(osm *osm.OSM) {
	for _, v := range osm.Ways {
		building := v.Tags.Find("building")
		if building != "" {
			street := v.Tags.Find("addr:street")
			houseNumber := v.Tags.Find("addr:housenumber")
			levels := v.Tags.Find("building:levels")
			if street == "проспект Александра Невского" && houseNumber == "63" {
				var coords []geom.Coord
				for _, wayNode := range v.Nodes {
					node, err := FindNode(wayNode.ID, osm.Nodes)
					utils.ProcessError(err)
					point := geom.Coord{node.Lat, node.Lon}
					coords = append(coords, point)
				}
				fmt.Println(coords)
				fmt.Println(CalculatePolygonArea(coords) * 5)
				fmt.Printf("Address: %s %s, Building type: %s, Levels: %s, ID: %d \n", street, houseNumber, building, levels, v.ID)
			}
		}

	}
}

var AcceptedHighWays = map[string]bool{
	"motorway":    true,
	"trunk":       true,
	"primary":     true,
	"secondary":   true,
	"tertiary":    true,
	"residential": true,
}

type Crossroad struct {
	Refs   int
	Street string
}

func LoadCrossRoads(mapData *osm.OSM) {
	var crossroads = map[osm.NodeID]*Crossroad{}
	for _, way := range mapData.Ways {
		highway := way.Tags.Find("highway")
		if AcceptedHighWays[highway] {
			for _, node := range way.Nodes {
				crossroad, contains := crossroads[node.ID]
				street := way.Tags.Find("name")
				if node.ID == 392987863 {
					fmt.Println(street)
				}

				if contains {
					if crossroad.Street != street {
						crossroads[node.ID].Refs += 1
					}
				} else {
					crossroads[node.ID] = &Crossroad{Refs: 1, Street: street}
				}
			}
		}
	}
	var nodes = osm.Nodes{}
	for k, v := range crossroads {
		if v.Refs > 1 {
			node, err := FindNode(k, mapData.Nodes)
			utils.ProcessError(err)
			nodes = append(nodes, node)
		}
	}
	p, err := json.Marshal(nodes)
	utils.ProcessError(err)

	f, err := os.Create("crossroads.json")
	utils.ProcessError(err)
	defer f.Close()
	f.Write(p)
}

func GetCenterDotOfWay(way osm.Way, nodes osm.Nodes) {
	var sumLat float64 = 0
	var sumLon float64 = 0
	for _, v := range way.Nodes {
		node, err := FindNode(v.ID, nodes)
		utils.ProcessError(err)
		sumLat += node.Lat
		sumLon += node.Lon
	}
	var nNodes float64 = float64(len(way.Nodes))
	fmt.Printf("Center dot: Lat: %f, Lon: %f \n", sumLat/nNodes, sumLon/nNodes)
}

func FindNode(id osm.NodeID, nodes osm.Nodes) (*osm.Node, error) {
	for _, v := range nodes {
		if id == v.ID {
			return v, nil
		}
	}
	return nil, fmt.Errorf("couldn't find node with id=%d", id)
}

func CalculatePolygonArea(coordinates []geom.Coord) float64 {
	var area float64
	for i := 0; i < len(coordinates)-1; i++ {
		p1 := coordinates[i]
		p2 := coordinates[i+1]
		area += ConvertToRadian(p2[1]-p1[1]) * (2 + math.Sin(ConvertToRadian(p1[0])) + math.Sin(ConvertToRadian(p2[0])))
	}

	area = area * 6378137 * 6378137 / 2

	return math.Abs(area)
}

func ConvertToRadian(input float64) float64 {
	return input * math.Pi / 180
}
