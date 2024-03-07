import folium
import json
import math



def create_map():
    print("Creating map")
    map = folium.Map(location=[61.7878389, 34.3638430], min_lat= 61, min_lon=34, max_lat=62, max_lon=35, max_bounds=True, zoom_start=10)
    f = open("weightedCrossroads.json", encoding="utf8")
    data = json.load(f)
    pop = 0
    for crossroad in data:
        population = calculate_population(crossroad["buildings"])
        folium.CircleMarker(location = (crossroad["Node"]["lat"], crossroad["Node"]["lon"]),
                                 fill_opacity = 0.6, 
                                 radius = population // 100,
                                 popup= "Node " + str(crossroad["Node"]["id"])).add_to(map)
        for building in crossroad["buildings"]:
            folium.PolyLine(locations=[(building["lat"], building["lon"]), (crossroad["Node"]["lat"], crossroad["Node"]["lon"])],
                             weight=2,
                             color = 'blue').add_to(map)
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




def calculate_population(buildings):
    population = 0
    for building in buildings:
        population += building["area"]*0.7 / 18
    return math.ceil(population)

if __name__ == "__main__":
    create_map()