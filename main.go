package main

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// all required constants
const (
	BASE_32               = "0123456789bcdefghjkmnpqrstuvwxyz"
	MAX_LATITUDE  float64 = 90
	MIN_LATITUDE  float64 = -90
	MAX_LONGITUDE float64 = 180
	MIN_LONGITUDE float64 = -180
	PRECISION     int     = 15
)

// Location struct
type Location struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Distance  float64 `json:"distance"`
}

type LocationData struct {
	Lat       float64    `json:"lat"`
	Lon       float64    `json:"lon"`
	PlaceData []Location `json:"placeData"`
}

// basic trie struct that contain nodes
type TrieNode struct {
	children map[byte]*TrieNode
	isEnd    bool
}

// trie intialization for GeoHash
type GeoHashTrie struct {
	root *TrieNode
}

// global trie initialization
var trie *GeoHashTrie

var centralLat float64
var centralLon float64

// a map that contains geohash as key and place name as value
var geohashAndPlaceMap = map[string]string{}

// function that initialize geohashtrie
func NewGeoHashTrie() *GeoHashTrie {
	return &GeoHashTrie{
		root: &TrieNode{
			children: make(map[byte]*TrieNode),
			isEnd:    false,
		},
	}
}

// insert in geohash trie
func (ght *GeoHashTrie) Insert(geohash string) {
	node := ght.root
	for i := 0; i < len(geohash); i++ {
		c := geohash[i]
		if _, ok := node.children[c]; !ok {
			node.children[c] = &TrieNode{
				children: make(map[byte]*TrieNode),
				isEnd:    false,
			}
		}
		node = node.children[c]
	}
	node.isEnd = true
}

// search for geohash in trie
func (ght *GeoHashTrie) Search(geohash string) bool {
	node := ght.root
	for i := 0; i < len(geohash); i++ {
		c := geohash[i]
		if _, ok := node.children[c]; !ok {
			return false
		}
		node = node.children[c]
	}
	return node.isEnd
}

// delete geohash from trie
func (ght *GeoHashTrie) Delete(geohash string) bool {
	return ght.deleteHelper(ght.root, geohash, 0)
}

func (ght *GeoHashTrie) deleteHelper(node *TrieNode, geohash string, depth int) bool {
	if node == nil {
		return false
	}

	// If we've reached the end of the geohash
	if depth == len(geohash) {
		// If this node is not the end of a geohash, it can't be deleted
		if !node.isEnd {
			return false
		}
		// Mark this node as not the end of a geohash
		node.isEnd = false
		// Return true if this node has no children, indicating it can be deleted
		return len(node.children) == 0
	}

	// Recursively delete from the appropriate child
	c := geohash[depth]
	if childNode, exists := node.children[c]; exists {
		shouldDeleteChild := ght.deleteHelper(childNode, geohash, depth+1)
		if shouldDeleteChild {
			delete(node.children, c)
			// Return true if this node has no children and is not the end of another geohash
			return len(node.children) == 0 && !node.isEnd
		}
	}

	return false
}

// Encode lat, lon into geohash string
func EncodeGeoHash(latitude, longitude float64, precision int) string {
	var (
		geohash          = make([]byte, precision)
		bits             = 0
		hash_value       = 0
		lat_min, lat_max = MIN_LATITUDE, MAX_LATITUDE
		lon_min, lon_max = MIN_LONGITUDE, MAX_LONGITUDE
	)

	for i := 0; i < precision; i++ {
		for j := 0; j < 5; j++ {
			bits++
			if bits%2 == 1 {
				mid := (lon_min + lon_max) / 2
				if longitude > mid {
					hash_value = (hash_value << 1) + 1
					lon_min = mid
				} else {
					hash_value = hash_value << 1
					lon_max = mid
				}
			} else {
				mid := (lat_min + lat_max) / 2
				if latitude > mid {
					hash_value = (hash_value << 1) + 1
					lat_min = mid
				} else {
					hash_value = hash_value << 1
					lat_max = mid
				}
			}
		}
		geohash[i] = BASE_32[hash_value]
		hash_value = 0
	}

	return string(geohash)
}

// Decode geohash string and return lat,lon
func DecodeGeoHash(geohash string) (float64, float64) {
	var (
		lat_min, lat_max = MIN_LATITUDE, MAX_LATITUDE
		lon_min, lon_max = MIN_LONGITUDE, MAX_LONGITUDE
		is_lon           = true
	)

	for _, c := range geohash {
		idx := strings.IndexByte(BASE_32, byte(c))
		for i := 4; i >= 0; i-- {
			bit := (idx >> i) & 1
			if is_lon {
				mid := (lon_min + lon_max) / 2
				if bit == 1 {
					lon_min = mid
				} else {
					lon_max = mid
				}
			} else {
				mid := (lat_min + lat_max) / 2
				if bit == 1 {
					lat_min = mid
				} else {
					lat_max = mid
				}
			}
			is_lon = !is_lon
		}
	}

	return (lat_min + lat_max) / 2, (lon_min + lon_max) / 2
}

// find nearby place from given geohash while considering distance
func (ght *GeoHashTrie) FindNearby(geohash string, distance float64) []string {
	lat, lon := DecodeGeoHash(geohash)
	precision := len(geohash)

	// Calculate bounding box
	latDelta := distance / 111.0 // Approximate degrees latitude per km
	lonDelta := distance / (111 * math.Cos(lat*math.Pi/180.0))

	minLat := lat - latDelta
	maxLat := lat + latDelta
	minLon := lon - lonDelta
	maxLon := lon + lonDelta

	// Encode bounding box corners
	sw := EncodeGeoHash(minLat, minLon, precision)
	ne := EncodeGeoHash(maxLat, maxLon, precision)

	return ght.RangeQuery(sw, ne)
}

// this function finds all the co-ordinates that come between given range
func (ght *GeoHashTrie) RangeQuery(sw, ne string) []string {
	commonPrefix := ght.commonPrefix(sw, ne)
	return ght.collectGeohashes(ght.root, commonPrefix, "")
}

func (ght *GeoHashTrie) commonPrefix(a, b string) string {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	return a[:i]
}

// collect geohashes
func (ght *GeoHashTrie) collectGeohashes(node *TrieNode, prefix string, current string) []string {
	var results []string
	if node.isEnd && strings.HasPrefix(current, prefix) {
		results = append(results, current)
	}

	for ch, child := range node.children {
		newCurrent := current + string(ch)
		if strings.HasPrefix(prefix, newCurrent) || strings.HasPrefix(newCurrent, prefix) {
			results = append(results, ght.collectGeohashes(child, prefix, newCurrent)...)
		}
	}

	return results
}

// calculate distance between two lat,lon using Haversine formula
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371

	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLon := (lon2 - lon1) * (math.Pi / 180)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func main() {
	log.Println("GeoHash with Tries")
	// trie := NewGeoHashTrie()

	// // Example usage
	// lat, lon := 48.858844, 2.294351 // Eiffel Tower
	// geohash := EncodeGeoHash(lat, lon, 7)
	// log.Printf("Encoded GeoHash: %s\n", geohash)

	// trie.Insert(geohash)
	// log.Printf("GeoHash found in trie: %v\n", trie.Search(geohash))

	// decoded_lat, decoded_lon := DecodeGeoHash(geohash)
	// log.Printf("Decoded coordinates: %.6f, %.6f\n", decoded_lat, decoded_lon)

	// Example usage
	// locations := []Location{
	// 	{"Gateway of India", 18.922002, 72.837783},
	// 	{"Juhu Beach", 19.075985, 72.843816},
	// 	{"Marine Drive", 18.920414, 72.835517},
	// 	{"Elephanta Caves", 18.963012, 72.890401},
	// 	{"Sanjay Gandhi National Park", 19.112514, 72.864211},
	// 	{"Bandra Fort", 19.041002, 72.839835},
	// 	{"Colaba Causeway", 18.922908, 72.837129},
	// 	{"Haji Ali Dargah", 18.924459, 72.833873},
	// 	{"Essel World", 19.118528, 72.891479},
	// 	{"Kanheri Caves", 19.140275, 72.872463},
	// }

	// Insert locations into the trie
	// for _, loc := range locations {
	// 	geohash := EncodeGeoHash(loc.Latitude, loc.Longitude, 15)
	// 	trie.Insert(geohash)
	// 	log.Printf("Inserted %s: %s\n", loc.Name, geohash)
	// }

	// const DIST_KM = 10.0

	// Find nearby locations
	// centerLat, centerLon := 48.861111, 2.336389 // Approximate center of Paris
	// centerLat, centerLon := 19.076090, 72.877426 // Approximate center of Mumbai
	// centerGeohash := EncodeGeoHash(centerLat, centerLon, 15)
	// nearbyGeohashes := trie.FindNearby(centerGeohash, DIST_KM) // 3 km radius

	// log.Printf("\nLocations within %.2f km of center of Mumbai:\n", DIST_KM)
	// for _, geohash := range nearbyGeohashes {
	// 	lat, lon := DecodeGeoHash(geohash)
	// 	distance := CalculateDistance(centerLat, centerLon, lat, lon)

	// 	if distance <= DIST_KM {
	// 		log.Printf("GeoHash: %s, Coordinates: (%.6f, %.6f), Distance: %.2f km\n", geohash, lat, lon, distance)
	// 	}
	// }

	// Range query example
	// swLat, swLon := 48.850, 2.290 // Southwest corner
	// neLat, neLon := 48.870, 2.340 // Northeast corner
	// swGeohash := EncodeGeoHash(swLat, swLon, 7)
	// neGeohash := EncodeGeoHash(neLat, neLon, 7)

	// log.Printf("\nLocations within range (%.6f, %.6f) to (%.6f, %.6f):\n", swLat, swLon, neLat, neLon)
	// rangeGeohashes := trie.RangeQuery(swGeohash, neGeohash)
	// for _, geohash := range rangeGeohashes {
	// 	lat, lon := DecodeGeoHash(geohash)
	// 	log.Printf("GeoHash: %s, Coordinates: (%.6f, %.6f)\n", geohash, lat, lon)
	// }
	handleHttpServer()
}

func GetDummyData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	q := r.URL.Query()
	val, exists := q["city"]
	fileName := ""
	if !exists {
		http.Error(w, "City code doesn't exists in query, Please provide city code", http.StatusBadRequest)
		return
	}
	cityCode := val[0]
	if cityCode == "1" {
		fileName = "mumbai.json"
	} else if cityCode == "2" {
		fileName = "ny.json"
	} else {
		http.Error(w, "Invalid city code", http.StatusBadRequest)
		return
	}
	data, err := os.ReadFile(fileName)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Parse the JSON data into a LocationData
	var jsonData LocationData
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	loadDataInTrie(jsonData)

	json.NewEncoder(w).Encode(jsonData)
}

// function to load json data into global trie
func loadDataInTrie(jsonData LocationData) {
	trie = NewGeoHashTrie()
	centralLat = jsonData.Lat
	centralLon = jsonData.Lon

	for _, place := range jsonData.PlaceData {
		geohash := EncodeGeoHash(place.Latitude, place.Longitude, PRECISION)
		trie.Insert(geohash)
		geohashAndPlaceMap[geohash] = place.Name
	}
}

func Welcome(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Welcome to Geo Hash service",
	})
}

func FindNearByHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	lat := r.URL.Query().Get("lat")
	lon := r.URL.Query().Get("lon")
	rad := r.URL.Query().Get("radius")
	if lat == "" || lon == "" || rad == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
	}
	radius, _ := strconv.ParseFloat(rad, 64)
	lati, _ := strconv.ParseFloat(lat, 64)
	long, err := strconv.ParseFloat(lon, 64)
	if err != nil {
		http.Error(w, "Invalid radius", http.StatusBadRequest)
	}
	geohash := EncodeGeoHash(lati, long, PRECISION)
	places := trie.FindNearby(geohash, radius)
	var locationsDist []Location
	for _, geohash := range places {
		var loc Location
		dlat, dlon := DecodeGeoHash(geohash)
		distance := CalculateDistance(lati, long, dlat, dlon)

		if distance <= radius {
			loc.Name = geohashAndPlaceMap[geohash]
			loc.Latitude = dlat
			loc.Longitude = dlon
			loc.Distance = distance
			locationsDist = append(locationsDist, loc)
		}
	}
	json.NewEncoder(w).Encode(locationsDist)
}

func handleHttpServer() {
	router := mux.NewRouter()
	router.HandleFunc("/", Welcome).Methods("GET")
	router.HandleFunc("/load-dummy", GetDummyData).Methods("GET")
	router.HandleFunc("/find-nearby", FindNearByHandler).Methods("GET")
	log.Println("Server is running on port : 8000")
	log.Panicln(http.ListenAndServe(":8000", router))
}
