import folium

def create_map():
    print("Creating map")
    map = folium.Map(location=[61.7878389, 34.3638430], min_lat= 61, min_lon=34, max_lat=62, max_lon=35, max_bounds=True, zoom_start=10)
    folium.Rectangle([(61.711, 34.164), (61.88, 34.752)], color="red").add_to(map)

    map.save("rect.html")

if __name__ == "__main__":
    create_map()