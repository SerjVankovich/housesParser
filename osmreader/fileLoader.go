package osmreader

import (
	"encoding/xml"
	"houseParser/utils"
	"io"
	"os"

	"github.com/paulmach/osm"
)

func LoadOsm(filepath string) *osm.OSM {
	allData := new(osm.OSM)
	f, err := os.Open(filepath)
	utils.ProcessError(err)
	defer f.Close()
	bytes, err := io.ReadAll(f)

	utils.ProcessError(err)
	err = xml.Unmarshal(bytes, allData)
	utils.ProcessError(err)
	return allData
}
