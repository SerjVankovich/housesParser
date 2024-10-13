package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"houseParser/utils"
	"io"
	"log"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	osrm "github.com/gojuno/go.osrm"
	geo "github.com/paulmach/go.geo"
	"github.com/paulmach/osm"
	"github.com/twpayne/go-geom"
	"golang.org/x/net/context"
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
const AREA_PER_PUPIL = 7.6
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
	//PlayCrossroads()
	//BuildDistancesAndLinesMatrix()
	b := 0.065
	eps := 0.001
	T := iterativeGravityModel(b, eps)
	SaveFloat64Matrix("correspondence2.txt", T)
}

func BuildCorrespondenceMatrix() {
	var crossroads []*CrossRoad
	LoadJson("weightedCrossroads.json", &crossroads)
	n := len(crossroads)

	distances := LoadDistances()

	corrMatrix := make([][]float64, n)
	for i := range corrMatrix {
		corrMatrix[i] = make([]float64, n)
	}
	for k := range 10 {
		print(k)
		Q_is, R_js := GetQ_is_R_js(crossroads, corrMatrix)
		CorrectMatrix(corrMatrix, crossroads)
		for i := range n {
			for j := range n {
				Q_i := Q_is[i]
				R_j := R_js[j]

				f_ij := DistFunc(distances[i][j])
				if Q_i != 0 && R_j != 0 && f_ij != 0 {
					corrMatrix[i][j] = corrMatrix[i][j] + Q_i*R_j*f_ij/SumRowApplied(distances[i], R_js, j)
				}
			}
		}
	}

	var nonEmpty []float64
	for _, row := range corrMatrix {
		for _, val := range row {
			if val != 0 {
				nonEmpty = append(nonEmpty, val)
			}
		}
	}
	println(len(nonEmpty))
	max := 0.0
	min := 100000.0
	sum := 0.0
	for _, val := range nonEmpty {
		if val > max {
			max = val
		}
		if val < min {
			min = val
		}
		sum += val
	}

	fmt.Printf("%10.f\n", max)
	println(min)
	fmt.Printf("%10.f\n", sum)

	SaveFloat64Matrix("correspondence.txt", corrMatrix)

}

// Итерационный алгоритм
func iterativeGravityModel(b float64, epsilon float64) [][]float64 {

	var crossroads []*CrossRoad
	LoadJson("weightedCrossroads.json", &crossroads)

	var d [][]float64 = LoadDistances()
	n := len(crossroads)
	T := make([][]float64, n)
	for i := range T {
		T[i] = make([]float64, n)
	}
	A := make([]float64, n)
	B := make([]float64, n)
	for k := range A {
		A[k] = 1.0
		B[k] = 1.0
	}
	iteration := 0
	// Итерационный процесс
	for {
		println(iteration)
		previousT := make([][]float64, n)
		for i := range previousT {
			previousT[i] = make([]float64, n)
			copy(previousT[i], T[i])
		}
		// Обновление A[i]
		for i := range A {
			sum := 0.0
			for j := range B {
				if i != j {
					sum += B[j] * crossroads[j].InWeight * DistFunc(d[i][j])
				}
			}
			if sum != 0 && math.Round(sum*1e20) != 0 {
				A[i] = 1.0 / sum
			}
		}
		// Обновление B[j]
		for j := range B {
			sum := 0.0
			for i := range A {
				if i != j {
					x := A[i] * crossroads[i].OutWeight * DistFunc(d[i][j])
					sum += x
				}
			}
			if sum != 0 && math.Round(sum*1e20) != 0 {
				B[j] = 1.0 / sum
			}
		}
		// Обновление матрицы T[i][j]
		for i := range T {
			for j := range T[i] {
				if i != j {
					T[i][j] = A[i] * crossroads[i].OutWeight * B[j] * crossroads[j].InWeight * DistFunc(d[i][j])
				} else {
					T[i][j] = 0
				}
			}
		}
		// Проверка сходимости
		converged := true
		for i := range T {
			for j := range T[i] {
				if T[i][j] != 0 && math.Abs(T[i][j]-previousT[i][j])/T[i][j] > epsilon {
					converged = false
					break
				}
			}
			if !converged {
				break
			}
		}
		if converged || iteration > 500 {
			break
		}
		iteration += 1
	}
	return T
}

func GetQ_is_R_js(crossroads []*CrossRoad, corrMatrix [][]float64) ([]float64, []float64) {
	var Q_is []float64
	var R_js []float64
	for i, crossroad := range crossroads {
		Q_is = append(Q_is, crossroad.OutWeight-SumRow(corrMatrix, i))
		R_js = append(R_js, crossroad.InWeight-SumCol(corrMatrix, i))
	}

	return Q_is, R_js
}

func CorrectMatrix(corrMatrix [][]float64, crossroads []*CrossRoad) {
	n := len(crossroads)
	for i := range n {
		for j := range n {
			D_j := crossroads[j].InWeight
			if D_j != 0 {
				sumCol := SumCol(corrMatrix, j)
				if sumCol > D_j {
					corrMatrix[i][j] = corrMatrix[i][j] * D_j / sumCol
				}
			}
		}
	}
}

func SumCol(corrMatrix [][]float64, j int) float64 {
	sum := 0.0
	for _, row := range corrMatrix {
		sum += row[j]
	}
	return sum
}

func SumRow(corrMatrix [][]float64, i int) float64 {
	sum := 0.0
	for _, col := range corrMatrix[i] {
		sum += col
	}
	return sum
}

func SumRowApplied(distances []float64, R_js []float64, j int) float64 {
	sum := 0.0
	for _, distance := range distances {
		sum += R_js[j] * DistFunc(distance)
	}
	return sum
}

func DistFunc(c_ij float64) float64 {
	const beta = 0.065
	return math.Exp(-beta * c_ij)
}

func LoadDistances() [][]float64 {
	file, err := os.Open("distances.txt")
	utils.ProcessError(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var distances [][]float64
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			var distanceLine []float64
			strDists := strings.Split(line, " ")
			for _, dist := range strDists {
				floatDist, _ := strconv.ParseFloat(dist, 32)
				distanceLine = append(distanceLine, floatDist)
			}
			distances = append(distances, distanceLine)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return distances
}

func BuildDistancesAndLinesMatrix() {
	var crossroads []*CrossRoad
	LoadJson("weightedCrossroads.json", &crossroads)
	wg := sync.WaitGroup{}
	n := len(crossroads)
	nContainers := 10

	var distances = make([][]float32, n)
	for i := range distances {
		distances[i] = make([]float32, n)
	}
	var lines = make([][]string, n)
	for i := range lines {
		lines[i] = make([]string, n)

	}
	step := n / 10
	start := 0
	stop := step
	for i := range nContainers {
		wg.Add(1)
		go func(start, stop, i int) {
			defer wg.Done()
			FillSubMatrix(start, stop, i, crossroads, distances, lines)
		}(start, stop, i)
		start = stop
		stop += step
	}
	wg.Wait()
	for _, distance := range distances {
		fmt.Println(distance[:10])
	}

	SaveFloatMatrix("distances.txt", distances)
	SaveLines(lines)
}

func FillSubMatrix(start int, stop int, batchNum int, crossroads []*CrossRoad, distances [][]float32, lines [][]string) {
	port := strconv.Itoa(5000 + batchNum + 1)
	if batchNum == 9 {
		stop = len(crossroads)
	}
	for i, crossroad1 := range crossroads[start:stop] {
		fmt.Println(port + ">" + strconv.Itoa(i) + " start: " + strconv.Itoa(start) + " stop: " + strconv.Itoa(stop))
		for j, crossroad2 := range crossroads {
			distance, line := FetchDistance(crossroad1, crossroad2, port)
			distances[start+i][j] = distance
			lines[start+i][j] = line
		}
	}
}

func SaveLines(lines [][]string) {
	f, err := os.Create("lines.txt")

	utils.ProcessError(err)

	defer f.Close()
	for _, linesArr := range lines {
		linesStr := strings.Join(linesArr, " ")
		_, err = f.WriteString(linesStr + "\n")
		utils.ProcessError(err)
	}
	fmt.Println("done")
}

func SaveFloatMatrix(file string, distances [][]float32) {
	f, err := os.Create(file)

	utils.ProcessError(err)

	defer f.Close()
	for _, distanceArr := range distances {
		distancesStr := strings.Trim(fmt.Sprint(distanceArr), "[]")
		_, err = f.WriteString(distancesStr + "\n")
		utils.ProcessError(err)
	}
	fmt.Println("done")
}

func SaveFloat64Matrix(file string, distances [][]float64) {
	f, err := os.Create(file)

	utils.ProcessError(err)

	defer f.Close()
	for _, distanceArr := range distances {
		distancesStr := strings.Trim(fmt.Sprint(distanceArr), "[]")
		_, err = f.WriteString(distancesStr + "\n")
		utils.ProcessError(err)
	}
	fmt.Println("done")
}

func FetchDistance(crossroad1 *CrossRoad, crossroad2 *CrossRoad, port string) (float32, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	client := osrm.NewFromURLWithTimeout("http://172.28.239.75:"+port, 100*time.Second)
	response, err := client.Route(ctx, osrm.RouteRequest{
		Profile: "car",
		Coordinates: osrm.NewGeometryFromPointSet(geo.PointSet{
			{crossroad1.Node.Lon, crossroad1.Node.Lat},
			{crossroad2.Node.Lon, crossroad2.Node.Lat},
		}),
		Steps:       osrm.StepsTrue,
		Annotations: osrm.AnnotationsTrue,
		Geometries:  osrm.GeometriesPolyline6,
	})
	utils.ProcessError(err)
	return response.Routes[0].Distance, getLine(response.Routes[0])
}

func getLine(route osrm.Route) string {
	legs := route.Legs
	var points geo.PointSet
	for _, leg := range legs {
		for _, step := range leg.Steps {
			points = slices.Concat(points, step.Geometry.PointSet)
		}
	}

	geometry := &osrm.Geometry{Path: geo.Path{PointSet: points}}
	return geometry.Polyline()
}

func findDistance(crossroad1 CrossRoad, crossroad2 CrossRoad, i int, j int, n int, distances [][]float64, lines [][]string) {
	if i == j {
		distances[i][j] = 0
		return
	}
	if crossroad1.OutWeight == 0 || crossroad2.InWeight == 0 {
		distances[i][j] = 0
	}
}

func PlayCrossroads() {

	//var wg sync.WaitGroup
	var crossroads []*CrossRoad
	LoadJson("weightedCrossroads.json", &crossroads)
	firstCrossroad := crossroads[0]
	secondCrossroad := crossroads[1]

	println(firstCrossroad.Node.Lon, secondCrossroad.Node.Lat)
	println()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	client := osrm.NewFromURLWithTimeout("http://172.28.239.75:5001", 1*time.Second)
	request := osrm.RouteRequest{
		Profile: "car",
		Coordinates: osrm.NewGeometryFromPointSet(geo.PointSet{
			{firstCrossroad.Node.Lon, firstCrossroad.Node.Lat},
			{secondCrossroad.Node.Lon, secondCrossroad.Node.Lat},
		}),
		Steps:       osrm.StepsTrue,
		Annotations: osrm.AnnotationsTrue,
		Geometries:  osrm.GeometriesPolyline6,
	}
	route, err := client.Route(ctx, request)
	utils.ProcessError(err)
	legs := route.Routes[0].Legs
	for i, leg := range legs {
		println("LEG: ", i)
		var points geo.PointSet
		for _, step := range leg.Steps {
			points = slices.Concat(points, step.Geometry.PointSet)
		}
		geometry := &osrm.Geometry{Path: geo.Path{PointSet: points}}
		println(geometry.Polyline())

	}
	log.Printf("routes are: %+v", route.Routes[0])

	//for i, crossroad := range crossroads {
	//	time.Sleep(1000 * time.Millisecond)
	//	for _, crossroad1 := range crossroads {
	//		i := i
	//		crossroad := crossroad
	//
	//		wg.Add(1)
	//		time.Sleep(10 * time.Millisecond)
	//
	//		crossroad1 := crossroad1
	//		go func() {
	//			ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Second)
	//			defer cancel()
	//			defer wg.Done()
	//			client := osrm.NewFromURLWithTimeout("http://0.0.0.0:5000", 10000*time.Second)
	//			_, err := client.Route(ctx, osrm.RouteRequest{
	//				Profile: "car",
	//				Coordinates: osrm.NewGeometryFromPointSet(geo.PointSet{
	//					{crossroad1.Node.Lon, crossroad1.Node.Lat},
	//					{crossroad.Node.Lon, crossroad.Node.Lat},
	//				}),
	//				Steps:       osrm.StepsFalse,
	//				Annotations: osrm.AnnotationsTrue,
	//				Overview:    osrm.OverviewFalse,
	//				Geometries:  osrm.GeometriesPolyline6,
	//			})
	//			utils.ProcessError(err)
	//			log.Printf("routes are: %d", i)
	//		}()
	//	}
	//}
	//wg.Wait()

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
