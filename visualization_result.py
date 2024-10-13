import folium
import json
import math
import polyline
import numpy as np

from matplotlib import pyplot as plt 
   

def get_lines():
    lines_str = open("lines.txt", 'r').read().split("\n")
    return [line.split(" ") for line in lines_str[:-1]]

def get_correspondence_matrix():
    data = open("correspondence2.txt", 'r').read().split("\n")
    return np.array([[float(x) for x in line.split(" ")] for line in data[:-1]])

def get_crossroads():
    f = open("weightedCrossroads.json", encoding="utf8")
    return json.load(f)

def create_map():

    crossroads = get_crossroads()
    matrix = get_correspondence_matrix()
    lines = get_lines()
    n = len(crossroads)

    print("Creating map")
    map = folium.Map(location=[61.7878389, 34.3638430], min_lat= 61, min_lon=34, max_lat=62, max_lon=35, max_bounds=True, zoom_start=10)
    for crossroad in crossroads:
        folium.CircleMarker(location = (crossroad["Node"]["lat"], crossroad["Node"]["lon"]),
                                 fill_opacity = 0.6)
    lines_clusters = [[], [], [], [], [], []]
    for i in range(n):
        for j in range(n):
            if i != j:
                value = matrix[i, j]
                if value > 0.000001 and value <= 20:
                    line = polyline.decode(lines[i][j])
                    lines_clusters[0].append(line)
                if value > 20 and value <= 50:
                    line = polyline.decode(lines[i][j])
                    lines_clusters[1].append(line)
                if value > 50 and value <= 100:
                    line = polyline.decode(lines[i][j])
                    lines_clusters[2].append(line)
                if value > 100 and value <= 500:
                    line = polyline.decode(lines[i][j])
                    lines_clusters[3].append(line)
                if value > 500 and value <= 1000:
                    line = polyline.decode(lines[i][j])
                    lines_clusters[4].append(line)
                if value > 1000:
                    line = polyline.decode(lines[i][j])
                    lines_clusters[5].append(line)

    for line in lines_clusters[0]:
        folium.PolyLine(locations=line, weight=5, color='#00FF41').add_to(map)
    for line in lines_clusters[1]:
        folium.PolyLine(locations=line, weight=5.5, color='#ffff01').add_to(map)
    for line in lines_clusters[2]:
        folium.PolyLine(locations=line, weight=6, color='#FFC000').add_to(map)
    for line in lines_clusters[3]:
        folium.PolyLine(locations=line, weight=6.5, color='#fe7e00').add_to(map)
    for line in lines_clusters[4]:
        folium.PolyLine(locations=line, weight=7, color='#ff4001').add_to(map)
    for line in lines_clusters[5]:
        folium.PolyLine(locations=line, weight=7.5, color='#fe0000').add_to(map)
    

    map.save("result.html")

# [1, 10, 20, 30, 40, 50, 70, 80, 100, 200, 300, 400, 500, 700, 1000, 2000]

def calc_statistic():
    matrix = get_correspondence_matrix()
    flatten = np.array([x for x in matrix.flatten() if x > 0])
    mean = flatten.mean()
    median = np.median(flatten)
    min = flatten.min()
    max = flatten.max()
    perc_95 = np.percentile(flatten, 95)
    perc_99 = np.percentile(flatten, 99.9)
    sum = np.sum(flatten)
    print("Min:", min)
    print("Max:", max)
    print("Mean:", mean)
    print("Median:", median)
    print("95% percentile:", perc_95)
    print("99.9% percentile:", perc_99)
    print("Sum:", sum)

if __name__ == "__main__":
    calc_statistic()
