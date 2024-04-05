package main

import (
	"encoding/json"
	"fmt"
	osrm "github.com/gojuno/go.osrm"
	geo "github.com/paulmach/go.geo"
	"github.com/paulmach/osm"
	"github.com/twpayne/go-geom"
	"golang.org/x/net/context"
	"houseParser/utils"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"sync"
	"time"
	//osrm "github.com/gojuno/go.osrm"
	//geo "github.com/paulmach/go.geo"
)

const HOUSE = "house"
const KINDERGARTEN = "kindergarten"
const SCHOOL = "school"
const COLLEGE = "college"
const UNIVERSITY = "university"
const OFFICE = "office"
const WAREHOUSE = "warehouse"
const INDUSTRIAL = "industrial"
const RETAIL = "retail"
const CINEMA = "cinema"
const SUPERMARKET = "supermarket"
const STADIUM = "stadium"
const RESTAURANT = "restaurant"
const HOSPITAL = "hospital"
const THEATRE = "theatre"
const MUSEUM = "museum"

const AREA_PER_PERSON = 18
const AREA_PER_PUPIL = 7.5
const AREA_PER_CHILD = 8
const AREA_PER_OFFICE_EMPLOYEE = 4.5
const AREA_PER_INDUSTRIAL_EMPLOYEE = 10
const AREA_PER_HOSPITAL = 10

const USEFUL_HOUSE_COEFF = 0.5
const WAREHOUSE_COEFF = 0.01
const RETAIL_COEFF = 0.1
const MUSEUM_COEFF = 0.1

var BUILDING_MAP = map[string]string{
	HOUSE:                HOUSE,
	"apartments":         HOUSE,
	"detached":           HOUSE,
	"dormitory":          HOUSE,
	"hotel":              HOUSE,
	"residential":        HOUSE,
	"semidetached_house": HOUSE,
	KINDERGARTEN:         KINDERGARTEN,
	SCHOOL:               SCHOOL,
	COLLEGE:              COLLEGE,
	UNIVERSITY:           UNIVERSITY,
	OFFICE:               OFFICE,
	"government":         OFFICE,
	"public":             OFFICE,
	"commercial":         OFFICE,
	WAREHOUSE:            WAREHOUSE,
	INDUSTRIAL:           INDUSTRIAL,
	RETAIL:               RETAIL,
	CINEMA:               CINEMA,
	SUPERMARKET:          SUPERMARKET,
	STADIUM:              STADIUM,
	RESTAURANT:           RESTAURANT,
	HOSPITAL:             HOSPITAL,
	MUSEUM:               MUSEUM,
}

var AMENITY_MAP = map[string]string{
	KINDERGARTEN:         KINDERGARTEN,
	SCHOOL:               SCHOOL,
	COLLEGE:              COLLEGE,
	UNIVERSITY:           UNIVERSITY,
	"research_institute": UNIVERSITY,
	"post_office":        OFFICE,
	"police":             OFFICE,
	"government":         OFFICE,
	"public":             OFFICE,
	"commercial":         OFFICE,
	"courthouse":         OFFICE,
	"townhall":           OFFICE,
	"social_facility":    OFFICE,
	"bank":               OFFICE,
	"fire_station":       OFFICE,
	WAREHOUSE:            WAREHOUSE,
	INDUSTRIAL:           INDUSTRIAL,
	RETAIL:               RETAIL,
	CINEMA:               CINEMA,
	SUPERMARKET:          SUPERMARKET,
	STADIUM:              STADIUM,
	RESTAURANT:           RESTAURANT,
	HOSPITAL:             HOSPITAL,
	"clinic":             HOSPITAL,
	MUSEUM:               MUSEUM,
	"arts_centre":        MUSEUM,
	"community_centre":   MUSEUM,
}

var SHOP_MAP = map[string]string{
	"department_store": RETAIL,
	"mall":             RETAIL,
	SUPERMARKET:        SUPERMARKET,
}

func main() {
	//osm := osmreader.LoadOsm("petrozavodsk.osm")
	//ways := LoadAllWays(osm)
	//nodes := LoadAllNodes(osm)
	//relations := LoadAllRelations(osm)
	//buildings := GetBuildingsFromRelations(nodes, ways, relations)
	//buildings = append(buildings, GetBuildingsFromNodes(ways, nodes)...)
	//p, err := json.Marshal(buildings)
	//utils.ProcessError(err)
	//
	//f, err := os.Create("buildings.json")
	//utils.ProcessError(err)
	//defer f.Close()
	//f.Write(p)
	//FindNearestCrossroads()
	//MergeCrossroads()
	PlayCrossroads()
}

func PlayCrossroads() {

	var wg sync.WaitGroup
	var crossroads []*CrossRoad
	LoadJson("weightedCrossroads.json", &crossroads)

	for i, crossroad := range crossroads {
		time.Sleep(1000 * time.Millisecond)
		for _, crossroad1 := range crossroads {
			i := i
			crossroad := crossroad

			wg.Add(1)
			time.Sleep(10 * time.Millisecond)

			crossroad1 := crossroad1
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Second)
				defer cancel()
				defer wg.Done()
				client := osrm.NewFromURLWithTimeout("http://0.0.0.0:5000", 10000*time.Second)
				_, err := client.Route(ctx, osrm.RouteRequest{
					Profile: "car",
					Coordinates: osrm.NewGeometryFromPointSet(geo.PointSet{
						{crossroad1.Node.Lon, crossroad1.Node.Lat},
						{crossroad.Node.Lon, crossroad.Node.Lat},
					}),
					Steps:       osrm.StepsFalse,
					Annotations: osrm.AnnotationsTrue,
					Overview:    osrm.OverviewFalse,
					Geometries:  osrm.GeometriesPolyline6,
				})
				utils.ProcessError(err)
				log.Printf("routes are: %d", i)
			}()
		}
	}
	wg.Wait()

}

func LoadAllRelations(mapData *osm.OSM) map[int64]*osm.Relation {
	var relations = map[int64]*osm.Relation{}
	for _, relation := range mapData.Relations {
		relations[int64(relation.ID)] = relation
	}
	return relations
}

func LoadAllNodes(mapData *osm.OSM) map[int64]*osm.Node {
	var nodes = map[int64]*osm.Node{}
	for _, node := range mapData.Nodes {
		nodes[int64(node.ID)] = node
	}
	return nodes
}

func LoadAllWays(mapData *osm.OSM) map[int64]*osm.Way {
	var ways = map[int64]*osm.Way{}
	for _, way := range mapData.Ways {
		ways[int64(way.ID)] = way
	}
	return ways
}

func PrintNodesWithAmenity(osm *osm.OSM) {
	for _, v := range osm.Nodes {
		amenity := v.Tags.Find("amenity")
		if amenity != "" {
			println(v.ID, amenity)
		}
	}
}

func MergeCrossroads() {
	var crossroads []*CrossRoad
	LoadJson("weightedCrossroads.json", &crossroads)

	var removedCrossroads = make(map[int64]struct{})
	var leftCrossroads []*CrossRoad

	for i, crossroad := range crossroads {
		if _, removed := removedCrossroads[int64(crossroad.Node.ID)]; !removed {
			for _, mCrossroad := range crossroads[i+1:] {
				if _, mRemoved := removedCrossroads[int64(mCrossroad.Node.ID)]; !mRemoved {
					distance := DistanceTrue(
						[2]float64{crossroad.Node.Lat, crossroad.Node.Lon},
						[2]float64{mCrossroad.Node.Lat, mCrossroad.Node.Lon},
					)
					if distance <= 55 {
						inOutSum1 := crossroad.InWeight + crossroad.OutWeight
						inOutSum2 := mCrossroad.InWeight + mCrossroad.OutWeight
						if inOutSum1 < inOutSum2 {
							mCrossroad.InWeight += crossroad.InWeight
							mCrossroad.OutWeight += crossroad.OutWeight
							removedCrossroads[int64(crossroad.Node.ID)] = struct{}{}
							break
						} else {
							crossroad.InWeight += mCrossroad.InWeight
							crossroad.OutWeight += mCrossroad.OutWeight
							removedCrossroads[int64(mCrossroad.Node.ID)] = struct{}{}
						}
					}
				}

			}
		}
	}

	for _, crossroad := range crossroads {
		if _, removed := removedCrossroads[int64(crossroad.Node.ID)]; !removed {
			leftCrossroads = append(leftCrossroads, crossroad)
		}
	}
	println(len(removedCrossroads))
	println(len(leftCrossroads))

	MarshallToJsonAndSave(leftCrossroads, "weightedCrossroads.json")

	//sort.Slice(distances, func(i, j int) bool {
	//	return distances[i].Distance < distances[j].Distance
	//})
	//
	//println(firstCrossroad.Node.ID)
	//
	//for _, distance := range distances {
	//	fmt.Printf("Distance: %2f %2f\n", float64(distance.Id), distance.Distance)
	//}

}

func MarshallToJsonAndSave(leftCrossroads interface{}, filename string) {
	p, err := json.Marshal(leftCrossroads)
	utils.ProcessError(err)

	f, err := os.Create(filename)
	utils.ProcessError(err)
	defer f.Close()
	f.Write(p)
}

func getDistanceBatch(crossroads []*CrossRoad, client *osrm.OSRM, ctx context.Context) [][]float32 {
	var points geo.PointSet
	for _, crossroad := range crossroads[:100] {
		points = append(points, geo.Point{crossroad.Node.Lon, crossroad.Node.Lat})
	}

	resp, err := client.Table(ctx, osrm.TableRequest{
		Profile:     "car",
		Coordinates: osrm.NewGeometryFromPointSet(points),
	})
	if err != nil {
		log.Fatalf("route failed: %v", err)
	}

	return resp.Distances
}

func CalculateLivingPopulation(buildings []*Building) float64 {
	var population = 0.0
	for _, building := range buildings {
		if building.Type == HOUSE {
			population += CalculatePopInBuilding(building)
		}
	}

	return math.Ceil(population)
}

func CalculateNonLivingPopulation(buildings []*Building) float64 {
	var population = 0.0
	for _, building := range buildings {
		if building.Type != HOUSE {
			population += CalculatePopInBuilding(building)
		}
	}
	return population
}

func CalculatePopInBuilding(building *Building) float64 {
	var area = building.Area
	var bType = building.Type
	if bType == KINDERGARTEN {
		return math.Ceil(area / AREA_PER_CHILD)
	}
	if bType == SCHOOL || bType == COLLEGE || bType == UNIVERSITY {
		return math.Ceil(area / AREA_PER_PUPIL)
	}
	if bType == OFFICE {
		return math.Ceil(area / AREA_PER_OFFICE_EMPLOYEE)
	}
	if bType == WAREHOUSE {
		return math.Ceil(area * WAREHOUSE_COEFF)
	}
	if bType == INDUSTRIAL {
		return math.Ceil(area / AREA_PER_INDUSTRIAL_EMPLOYEE)
	}

	if bType == RETAIL || bType == SUPERMARKET || bType == RESTAURANT {
		return math.Ceil(area * RETAIL_COEFF)
	}

	if bType == HOSPITAL {
		return math.Ceil(area / AREA_PER_HOSPITAL)
	}

	if bType == MUSEUM {
		return math.Ceil(area * MUSEUM_COEFF)
	}

	if bType == HOUSE {
		return math.Ceil(area * USEFUL_HOUSE_COEFF / AREA_PER_PERSON)
	}
	return 0
}

type Point struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Building struct {
	Center  *Point   `json:"center"`
	Polygon []*Point `json:"polygon"`
	Area    float64  `json:"area"`
	Type    string   `json:"type"`
	Address string   `json:"address"`
	Levels  int      `json:"levels"`
}

type CrossRoad struct {
	Buildings []*Building `json:"buildings"`
	Node      *osm.Node   `json:"Node"`
	InWeight  float64     `json:"in_weight"`
	OutWeight float64     `json:"out_weight"`
}

func GetBuildingsFromNodes(ways map[int64]*osm.Way, nodes map[int64]*osm.Node) []*Building {
	var buildings []*Building
	for _, v := range ways {
		building := v.Tags.Find("building")
		if building != "" {
			amenity := v.Tags.Find("amenity")
			shop := v.Tags.Find("shop")
			street := v.Tags.Find("addr:street")
			houseNumber := v.Tags.Find("addr:housenumber")
			if street == "набережная Ла-Рошель" {
				fmt.Println(building, amenity, street, houseNumber)
			}
			bType, ok := GetType(building, amenity, shop, street)
			if ok {

				levels := ParseLevels(v.Tags.Find("building:levels"))
				var coords []*Point
				var sumLat float64 = 0
				var sumLon float64 = 0
				for _, wayNode := range v.Nodes {
					if node, ok := nodes[int64(wayNode.ID)]; ok {
						point := &Point{node.Lat, node.Lon}
						coords = append(coords, point)
						sumLat += node.Lat
						sumLon += node.Lon
					}
				}
				numNodes := len(v.Nodes)
				center := &Point{sumLat / float64(numNodes), sumLon / float64(numNodes)}
				buildings = append(buildings, &Building{center, coords, 0, bType, street + " " + houseNumber, levels})
			}
		}
	}
	return buildings
}

func GetType(building, amenity, shop, street string) (string, bool) {
	if shopType, sOk := SHOP_MAP[shop]; sOk {
		return shopType, true
	}
	if buildingType, bOk := BUILDING_MAP[building]; bOk {
		return buildingType, true
	}
	if amenityType, aOk := AMENITY_MAP[amenity]; aOk {
		return amenityType, true
	}

	if building == "yes" && street != "" {
		return HOUSE, true
	}
	return "", false
}

func FindNearestCrossroads() {
	var buildings []*Building
	var crossroads []*osm.Node
	var weightedCrossRoads = map[int64]*CrossRoad{}

	LoadJson("buildings.json", &buildings)
	LoadJson("crossroads.json", &crossroads)

	for _, crossroad := range crossroads {
		weightedCrossRoads[int64(crossroad.ID)] = &CrossRoad{[]*Building{}, crossroad, 0, 0}
	}

	for _, building := range buildings {
		minCrossRoad := crossroads[0]
		minDistance := Distance([2]float64{building.Center.Lon, building.Center.Lat}, minCrossRoad.Point())
		for _, crossroad := range crossroads {
			distance := Distance([2]float64{building.Center.Lon, building.Center.Lat}, crossroad.Point())
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

	var crossroadsToSave []*CrossRoad
	for _, value := range weightedCrossRoads {
		value.InWeight = CalculateNonLivingPopulation(value.Buildings)
		value.OutWeight = CalculateLivingPopulation(value.Buildings)
		crossroadsToSave = append(crossroadsToSave, value)
	}

	MarshallToJsonAndSave(crossroadsToSave, "weightedCrossroads.json")
}

func Distance(dot1, dot2 [2]float64) float64 {
	sum := 0.0
	for i, c := range dot1 {
		sum += math.Pow(c-dot2[i], 2)
	}
	return math.Sqrt(sum)
}

func DistanceTrue(dot1, dot2 [2]float64) float64 {
	lat1 := dot1[0]
	lat2 := dot2[0]
	lon1 := dot1[1]
	lon2 := dot2[1]
	var R = 6378137. // Earth’s mean radius in meter
	var dLat = rad(lat2 - lat1)
	var dLong = rad(lon2 - lon1)
	var a = math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rad(lat1))*math.Cos(rad(lat2))*
			math.Sin(dLong/2)*math.Sin(dLong/2)
	var c = 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	var d = R * c
	return d
}

func rad(x float64) float64 {
	return x * math.Pi / 180
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

func GetBuildingsFromRelations(nodes map[int64]*osm.Node, ways map[int64]*osm.Way, relations map[int64]*osm.Relation) []*Building {
	var buildings []*Building
	for _, relation := range relations {
		building := relation.Tags.Find("building")
		if building != "" {
			amenity := relation.Tags.Find("amenity")
			shop := relation.Tags.Find("shop")
			street := relation.Tags.Find("addr:street")
			houseNumber := relation.Tags.Find("addr:housenumber")
			bType, ok := GetType(building, amenity, shop, street)
			if ok {
				coords, center := getAllDotsFromRelation(relation, ways, nodes)
				if !math.IsNaN(center.Lat) {
					levels := ParseLevels(relation.Tags.Find("building:levels"))
					buildings = append(buildings, &Building{center, coords, 0, bType, street + " " + houseNumber, levels})
				}
			}
		}
	}
	return buildings
}

func getAllDotsFromRelation(v *osm.Relation, ways map[int64]*osm.Way, nodes map[int64]*osm.Node) ([]*Point, *Point) {
	var mWays []*osm.Way
	var coords []*Point
	var sumLat float64 = 0
	var sumLon float64 = 0
	for _, member := range v.Members {
		mType, mRole, mRef := member.Type, member.Role, member.Ref
		if mType == osm.TypeWay && mRole == "outer" {
			if way, ok := ways[mRef]; ok {
				mWays = append(mWays, way)
				for _, wayNode := range way.Nodes {
					if node, ok := nodes[int64(wayNode.ID)]; ok {
						point := &Point{node.Lat, node.Lon}
						coords = append(coords, point)
						sumLat += node.Lat
						sumLon += node.Lon
					}

				}
			}
		}
	}

	numNodes := len(coords)
	center := &Point{sumLat / float64(numNodes), sumLon / float64(numNodes)}

	return coords, center
}

func ConvertToRadian(input float64) float64 {
	return input * math.Pi / 180
}
