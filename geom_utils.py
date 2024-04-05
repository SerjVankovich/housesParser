from geographiclib.polygonarea import PolygonArea
from geographiclib.geodesic import Geodesic
import math
import json

def main():
    f = open("buildings.json", encoding="UTF-8")
    data = json.load(f)
    f.close()
    for building in data:
        polygon = building["polygon"]
        levels = building["levels"]
        polygon_area = PolygonArea(Geodesic.WGS84)
        for point in polygon:
            polygon_area.AddPoint(point["lat"], point["lon"])
        area = math.fabs(polygon_area.Compute(reverse=True)[2] * levels)
        building["area"] = area
    f = open("buildings.json", "w", encoding="UTF-8")
    data = json.dump(data, f, ensure_ascii=False)
    f.close()
    

main()