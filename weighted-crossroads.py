import folium
import json


def create_map():
    print("Creating map")
    map = folium.Map(location=[61.7878389, 34.3638430], min_lat= 61, min_lon=34, max_lat=62, max_lon=35, max_bounds=True, zoom_start=10)
    f = open("weightedCrossroads.json", encoding="utf8")
    data = json.load(f)
    for crossroad in data:
        folium.CircleMarker(location = (crossroad["Node"]["lat"], crossroad["Node"]["lon"]),
                                 fill_opacity = 0.6, 
                                 radius = 5).add_to(map)


    map.save("cross-weighted.html")
    f.close()


if __name__ == "__main__":
    create_map()