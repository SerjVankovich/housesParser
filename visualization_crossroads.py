import folium
import json
import math

AREA_PER_PERSON = 18
AREA_PER_PUPIL = 7.5
AREA_PER_CHILD = 8
AREA_PER_OFFICE_EMPLOYEE = 4.5
AREA_PER_INDUSTRIAL_EMPLOYEE = 10
AREA_PER_HOSPITAL = 10

USEFUL_HOUSE_COEFF = 0.5
WAREHOUSE_COEFF = 0.01
RETAIL_COEFF = 0.1
MUSEUM_COEFF = 0.1

HOUSE = "house"
KINDERGARTEN = "kindergarten"
SCHOOL = "school"
COLLEGE = "college"
UNIVERSITY = "university"
OFFICE = "office"
WAREHOUSE = "warehouse"
INDUSTRIAL = "industrial"
RETAIL = "retail"
CINEMA = "cinema"
SUPERMARKET = "supermarket"
STADIUM = "stadium"
RESTAURANT = "restaurant"
HOSPITAL = "hospital"
THEATRE = "theatre"
MUSEUM = "museum"


def create_map():
    print("Creating map")
    map = folium.Map(location=[61.7878389, 34.3638430], min_lat= 61, min_lon=34, max_lat=62, max_lon=35, max_bounds=True, zoom_start=10)
    f = open("weightedCrossroads.json", encoding="utf8")
    data = json.load(f)
    pop = 0
    no_liv_pop = 0
    for crossroad in data:
        population = calculate_living_population(crossroad["buildings"])

        folium.CircleMarker(location = (crossroad["Node"]["lat"], crossroad["Node"]["lon"]),
                                 fill_opacity = 0.6, 
                                 radius = population // 100,
                                 popup= "Population " + str(population)).add_to(map)
        for building in crossroad["buildings"]:
            if building["type"] == HOUSE:
                
                folium.PolyLine(locations=[(building["center"]["lat"], building["center"]["lon"]), (crossroad["Node"]["lat"], crossroad["Node"]["lon"])],
                             weight=2,
                             color = 'blue').add_to(map)
                
                b_pop = calculate_pop_in_building(building)
                folium.CircleMarker(location = (building["center"]["lat"], building["center"]["lon"]),
                                 fill_opacity = 0.6, 
                                 radius = b_pop // 100,
                                 popup= "Population " + str(b_pop)).add_to(map)
        print(population, crossroad["Node"])
        pop += population
    print(pop)

    # f = open("buildings.json", encoding="utf8")
    # buildings = json.load(f)
    # for building in buildings:
    #     folium.CircleMarker(location = (building["lat"], building["lon"]),
    #                              fill_opacity = 0.6, 
    #                              radius = 6,
    #                              color='#FFFF00',
    #                              popup=building["address"]).add_to(map)

    map.save("map.html")
    f.close()




def calculate_living_population(buildings):
    population = 0
    for building in buildings:
        if building["type"] == "house":
            population += building["area"]*0.5 / 18
    return math.ceil(population)

def calculate_non_liv_pop(buildings):
    population = 0
    for building in buildings:
        if building["type"] != HOUSE:
            population += calculate_pop_in_building(building)
    return population
def calculate_pop_in_building(building):
    area = building["area"]
    b_type = building["type"]
    if b_type == KINDERGARTEN:
        return math.ceil(area / AREA_PER_CHILD)
    if b_type in [SCHOOL, COLLEGE, UNIVERSITY]:
        return math.ceil(area / AREA_PER_PUPIL)
    if b_type == OFFICE:
        return math.ceil(area / AREA_PER_OFFICE_EMPLOYEE)
    if b_type == WAREHOUSE:
        return math.ceil(area * WAREHOUSE_COEFF)
    if b_type == INDUSTRIAL:
        return math.ceil(area / AREA_PER_INDUSTRIAL_EMPLOYEE)
    if b_type in [RETAIL, SUPERMARKET, RESTAURANT]:
        return math.ceil(area * RETAIL_COEFF)
    if b_type == HOSPITAL:
        return math.ceil(area / AREA_PER_HOSPITAL)
    if b_type == MUSEUM:
        return math.ceil(area * MUSEUM_COEFF)
    if b_type == HOUSE:
        return math.ceil(area * USEFUL_HOUSE_COEFF / AREA_PER_PERSON)
    return 0

if __name__ == "__main__":
    create_map()